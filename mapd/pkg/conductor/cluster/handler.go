// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	versionsv1alpha1 "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"github.com/wI2L/jsondiff"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/conductor/cluster/extension"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	mapv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/mapd/v1alpha1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/randutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/tanzu"
	"go.uber.org/zap"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// Handler handle lifecycle of Cluster.
type Handler interface {
	ProvisionPrepare() error
	ProvisionKubernetes() (endpoint string, state model.ClusterState, caCert []byte, err error)
	ProvisionNamespace() error
	ProvisionServiceAccount() error
	ProvisionClusterRoleBinding() (token []byte, err error)
	ProvisionValidation(namespace ...string) error
	ProvisionPostScripts() error
	ProvisionFinish() error
	CheckHealth() (model.ClusterState, error)
	GetTrustedCACerts() ([]byte, error)
}

// Kubernetes implements Handler of attached cluster.
type Kubernetes struct {
	Cluster                *model.Cluster
	PrimaryImageRegistry   *model.ImageRegistry
	RelatedImageRegistries []model.ImageRegistry
	LoadBalancer           *model.LoadBalancer

	log *zap.SugaredLogger
}

func (k *Kubernetes) newClient() (*k8s.Kubernetes, error) {
	authToken := k.Cluster.AuthToken
	if len(k.Cluster.AuthToken) > 0 {
		block, err := cryptoutil.NewAESGCMCipher(k.Cluster.ID[:]) // AES-128
		if err != nil {
			return nil, err
		}
		authToken = block.Decrypt(k.Cluster.AuthToken)
	}
	return k8s.NewKubernetesFromToken(k.Cluster.Endpoint, authToken, rest.TLSClientConfig{
		Insecure: k.Cluster.InsecureTLS,
		CAData:   k.Cluster.CACert,
	})
}

func (k *Kubernetes) ProvisionPrepare() error {
	return nil
}

func (k *Kubernetes) ProvisionKubernetes() (endpoint string, state model.ClusterState, caCert []byte, err error) {
	return "", model.ClusterStateUnknown, nil, nil
}

func (k *Kubernetes) ProvisionNamespace() error {
	c, err := k.newClient()
	if err != nil {
		return err
	}
	return k.provisionNamespace(c)
}

func (k *Kubernetes) provisionNamespace(c *k8s.Kubernetes) error {
	defer k.log.Sync()

	k.log.Infof("checking namespace %s state on cluster %s...", defNamespace, strconv.Quote(k.Cluster.Name))
	if _, err := c.GetNamespace(defNamespace); err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("unexpected error: %v.", err)
			return err
		}
		k.log.Infof("namespace %s does not exist on cluster %s, now creating...", defNamespace, strconv.Quote(k.Cluster.Name))
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: defNamespace,
				Labels: map[string]string{
					defLabelInternal: "true",
					defLabelManaged:  "true",
				},
			},
		}
		if _, err := c.CreateNamespace(namespace); err != nil {
			k.log.Errorf("failed to provision namespace %s on cluster %s due to: %v.", defNamespace, strconv.Quote(k.Cluster.Name), err)
			return err
		}
		k.log.Infof("namespace %s is created but requires initialization.", defNamespace)
	} else {
		k.log.Warnf("namespace %s already exists on cluster %s.", defNamespace, strconv.Quote(k.Cluster.Name))
	}
	k.log.Infof("checking docker registry secret %s/%s state...", defNamespace, defRegistrySecret)
	if _, err := c.GetSecret(defNamespace, defRegistrySecret); err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("unexpected error: %v.", err)
			return err
		}
		k.log.Infof("docker registry secret %s/%s does not exist on cluster %s, now creating...", defNamespace, defRegistrySecret, strconv.Quote(k.Cluster.Name))

		cred := &k8s.DockerConfigJSON{
			Auths: make(map[string]*k8s.DockerConfigJSONAuth),
		}

		for _, registry := range k.RelatedImageRegistries {
			block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
			if err != nil {
				k.log.Errorf("unexpected error: %v.", err)
				return err
			}

			if registry.AuthUser != "" {
				password := block.Decrypt(registry.AuthPassword)
				cred.Auths[registry.Endpoint] = &k8s.DockerConfigJSONAuth{
					Username: registry.AuthUser,
					Password: string(password),
					Auth:     []byte(strings.Join([]string{registry.AuthUser, string(password)}, ":")),
				}
			}
		}
		b, err := json.Marshal(cred)
		if err != nil {
			k.log.Errorf("failed to provision docker registry secret %s/%s on cluster %s due to: %v.", defNamespace, defRegistrySecret, strconv.Quote(k.Cluster.Name), err)
			return fmt.Errorf("error constructing .dockerconfigjson: %v", err)
		}

		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      defRegistrySecret,
				Namespace: defNamespace,
				Labels: map[string]string{
					defLabelInternal: "true",
					defLabelManaged:  "true",
				},
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{".dockerconfigjson": b},
		}
		if _, err = c.CreateSecret(secret); err != nil {
			k.log.Errorf("failed to provision docker registry secret %s/%s on cluster %s due to: %v.", defNamespace, defRegistrySecret, strconv.Quote(k.Cluster.Name), err)
			return err
		}
		k.log.Infof("docker registry secret %s is created, and namespace %s is now ready for use.", defRegistrySecret, defNamespace)
	} else {
		k.log.Warnf("docker registry secret %s/%s already exists on cluster %s.", defNamespace, defRegistrySecret, strconv.Quote(k.Cluster.Name))
	}
	return nil
}

func (k *Kubernetes) ProvisionServiceAccount() error {
	c, err := k.newClient()
	if err != nil {
		return err
	}
	return k.provisionServiceAccount(c)
}

func (k *Kubernetes) provisionServiceAccount(c *k8s.Kubernetes) error {
	defer k.log.Sync()

	k.log.Infof("checking service account %s state on kubernetes cluster...", defServiceAccount)
	sa, err := c.GetServiceAccount(defNamespace, defServiceAccount)
	if err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("unexpected error: %v.", err)
			return err
		}
		k.log.Infof("service account %s/%s does not exist on cluster %s, now creating...", defNamespace, defServiceAccount, strconv.Quote(k.Cluster.Name))
		if _, err := createServiceAccount(c, defNamespace, defServiceAccount); err != nil {
			k.log.Errorf("failed to provision service account %s/%s on cluster %s due to: %v.", defNamespace, defServiceAccount, strconv.Quote(k.Cluster.Name), err)
			return err
		}
		k.log.Infof("service account %s/%s is creating and we will check it again later.", defNamespace, defServiceAccount)
		return err
	}
	if len(sa.Secrets) < 1 {
		k8sVersion, err := c.Version()
		if err != nil {
			k.log.Errorf("cannot obtain kubernetes version from cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
			return err
		}
		ver, err := semver.NewVersion(k8sVersion.GitVersion)
		if err != nil {
			k.log.Errorf("cluster %s's version %s does not seem like a valid semantic version: %v.", strconv.Quote(k.Cluster.Name), k8sVersion.GitVersion, err)
			return err
		}
		if ver.GreaterThan(semver.MustParse("v1.23")) || (k8sVersion.Major == "1" && k8sVersion.Minor == "23") {
			k.log.Infof("create token for service account %s/%s as cluster %s's version %s is greater than 1.23.", defNamespace, defServiceAccount, strconv.Quote(k.Cluster.Name), k8sVersion.GitVersion)
			secret, err := createServiceAccountToken(c, defNamespace, defServiceAccount+"-"+randutil.StringN(5, randutil.LOWERCASE), defServiceAccount)
			if err != nil {
				k.log.Errorf("cannot create token for service account %s/%s on cluster %s due to: %v.", defNamespace, defServiceAccount, strconv.Quote(k.Cluster.Name), err)
				return err
			}

			sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
				Kind:       secret.Kind,
				Namespace:  secret.Namespace,
				Name:       secret.Name,
				APIVersion: secret.APIVersion,
			})
			if _, err := c.UpdateServiceAccount(sa); err != nil {
				k.log.Errorf("failed to update service account %s/%s on cluster %s due to: %v.", sa.Namespace, sa.Name, strconv.Quote(k.Cluster.Name), err)
				return err
			}
			return fmt.Errorf("token of service account is missing")
		}

		k.log.Infof("service account %s/%s seems to be creating as we cannot find a valid secret under it, and we will check it again later.", defNamespace, defServiceAccount)
		return fmt.Errorf("number of secrets is not satisfied")
	}

	// update default service account for tanzu build service, no matter
	// whether it will be enabled or not.
	for _, each := range []string{"default", defServiceAccount} {
		sa, err = c.GetServiceAccount(defNamespace, each)
		if err != nil {
			if !k8s.IsErrObjectNotExists(err) {
				k.log.Errorf("unexpected error: %v.", err)
				return err
			}
			k.log.Infof("service account %s/%s does not exist on cluster %s, now creating...", defNamespace, each, strconv.Quote(k.Cluster.Name))
			if sa, err = createServiceAccount(c, defNamespace, each); err != nil {
				k.log.Errorf("failed to provision service account %s/%s on cluster %s due to: %v.", defNamespace, each, strconv.Quote(k.Cluster.Name), err)
				return err
			}
			k.log.Infof("service account %s/%s is created.", defNamespace, each)
		}

		if sa.ObjectMeta.Labels == nil {
			sa.ObjectMeta.Labels = make(map[string]string)
		}
		sa.ObjectMeta.Labels[defLabelManaged] = "true"
		sm := make(map[string]corev1.ObjectReference)
		for _, each := range sa.Secrets {
			sm[each.Name] = each
		}
		sm[defRegistrySecret] = corev1.ObjectReference{
			Kind:       "Secret",
			Namespace:  sa.Namespace,
			Name:       defRegistrySecret,
			APIVersion: corev1.SchemeGroupVersion.String(),
		}
		sa.Secrets = sa.Secrets[:0]
		for _, each := range sm {
			sa.Secrets = append(sa.Secrets, each)
		}

		if sa.ObjectMeta.Annotations == nil {
			sa.ObjectMeta.Annotations = make(map[string]string)
		}
		kpackVal, ok := sa.ObjectMeta.Annotations["kpack.io/managedSecret"]
		if !ok {
			kpackVal = "{}"
		}
		kpackM := make(map[string]string)
		json.Unmarshal([]byte(kpackVal), &kpackM)

		u, err := url.Parse(k.PrimaryImageRegistry.Endpoint)
		if err != nil {
			k.log.Errorf("unexpected error: %v.", err)
			return err
		}

		kpackM[defRegistrySecret] = u.Host
		b, _ := json.Marshal(kpackM)
		sa.ObjectMeta.Annotations["kpack.io/managedSecret"] = string(b)
		if _, err := c.UpdateServiceAccount(sa); err != nil {
			k.log.Errorf("failed to update service account %s/%s on cluster %s due to: %v.", defNamespace, each, strconv.Quote(k.Cluster.Name), err)
			return err
		}
	}

	sa, err = c.GetServiceAccount(defNamespace, defServiceAccount)
	if err != nil {
		return err
	}
	var defRegistrySecretFound bool
	for _, each := range sa.ImagePullSecrets {
		if each.Name == defRegistrySecret {
			defRegistrySecretFound = true
			break
		}
	}
	if !defRegistrySecretFound {
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
			Name: defRegistrySecret,
		})
		if _, err := c.UpdateServiceAccount(sa); err != nil {
			k.log.Errorf("failed to update service account %s/%s on cluster %s due to: %v.", sa.Namespace, sa.Name, strconv.Quote(k.Cluster.Name), err)
			return err
		}
	}
	return nil
}

func (k *Kubernetes) ProvisionClusterRoleBinding() ([]byte, error) {
	c, err := k.newClient()
	if err != nil {
		return nil, err
	}
	return k.provisionClusterRoleBinding(c)
}

func (k *Kubernetes) provisionClusterRoleBinding(c *k8s.Kubernetes) ([]byte, error) {
	defer k.log.Sync()

	k.log.Infof("checking cluster role binding %s state on cluster %s...", defClusterAdminClusterRoleBinding, strconv.Quote(k.Cluster.Name))
	if _, err := c.GetClusterRoleBinding(defClusterAdminClusterRoleBinding); err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("unexpected error: %v.", err)
			return nil, err
		}
		k.log.Infof("cluster role binding %s does not exist on cluster %s, now creating...", defClusterAdminClusterRoleBinding, strconv.Quote(k.Cluster.Name))
		if _, err := createClusterRoleBindingClusterAdmin(c, defClusterAdminClusterRoleBinding, defNamespace, defServiceAccount); err != nil {
			k.log.Errorf("failed to provision cluster role binding %s on cluster %s due to: %v.", defClusterAdminClusterRoleBinding, strconv.Quote(k.Cluster.Name), err)
			return nil, err
		}
		k.log.Infof("cluster role binding %s is created and assigned to service account %s/%s.", defClusterAdminClusterRoleBinding, defNamespace, defServiceAccount)
	} else {
		k.log.Warnf("cluster role binding %s already exists.", defClusterAdminClusterRoleBinding)
	}

	k.log.Infof("checking service account %s state on cluster %s...", defServiceAccount, strconv.Quote(k.Cluster.Name))
	sa, err := c.GetServiceAccount(defNamespace, defServiceAccount)
	if err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("unexpected error: %v.", err)
			return nil, err
		}
		return nil, &ErrCriticalFailure{err}
	}

	var secret *corev1.Secret
	for _, each := range sa.Secrets {
		obj, err := c.GetSecret(defNamespace, each.Name)
		if err != nil {
			k.log.Errorf("cannot obtain service token from secret %s/%s on cluster %s due to: %v.", defNamespace, each.Name, strconv.Quote(k.Cluster.Name), err)
			return nil, err
		}
		if obj.Type != corev1.SecretTypeServiceAccountToken {
			continue
		}
		secret = obj
	}
	if secret == nil {
		k.log.Errorf("unable to find a valid token of service account %s/%s on cluster %s.", defNamespace, defServiceAccount, strconv.Quote(k.Cluster.Name))
		return nil, fmt.Errorf("token of service account is missing")
	}
	if secret.Data == nil {
		k.log.Infof("secret %s/%s seems to be creating on cluster %s as we cannot find any data on it, and we will check it again later.", defNamespace, secret.Name, strconv.Quote(k.Cluster.Name))
		return nil, fmt.Errorf("token of service account is missing")
	}
	token, ok := secret.Data["token"]
	if !ok {
		k.log.Warnf("data was found but token is missing on secret %s/%s, and we will check it again later.", defNamespace, secret.Name)
		return nil, fmt.Errorf("token of service account is missing")
	}
	return token, nil
}

func (k *Kubernetes) ProvisionValidation(namespace ...string) error {
	c, err := k.newClient()
	if err != nil {
		return err
	}
	return k.provisionValidation(c, namespace...)
}

func (k *Kubernetes) provisionValidation(c *k8s.Kubernetes, namespace ...string) error {
	defer k.log.Sync()

	k.log.Infof("validating critical kubernetes components...")
	nss := append([]string{"kube-system"}, namespace...)
	for _, ns := range nss {
		pods, err := c.GetPods(ns)
		if err != nil {
			k.log.Errorf("failed to retrieve pods under namespace %s of cluster %s due to: %v.", ns, strconv.Quote(k.Cluster.Name), err)
			return err
		}

	LOOP:
		for _, pod := range pods.Items {
			switch pod.Status.Phase {
			case corev1.PodPending:
				k.log.Warnf("pod %s/%s is pending on cluster %s. The cluster may not be fully functional.", ns, pod.Name, strconv.Quote(k.Cluster.Name))
				continue LOOP
			case corev1.PodSucceeded:
				continue LOOP
			case corev1.PodFailed:
				k.log.Warnf("pod %s/%s is failed on cluster %s. The cluster may not be fully functional.", ns, pod.Name, strconv.Quote(k.Cluster.Name))
				continue LOOP
			}
			ready := 0
			for _, container := range pod.Status.ContainerStatuses {
				if container.Ready {
					ready++
				}
			}
			if ready < len(pod.Status.ContainerStatuses) {
				k.log.Warnf("provision will not be proceeded until pod %s/%s in cluster %s is ready or failed (current: %d/%d).", ns, pod.Name, strconv.Quote(k.Cluster.Name), ready, len(pod.Status.ContainerStatuses))
				return fmt.Errorf("pod %s is starting", pod.Name)
			}
		}
	}
	return nil
}

func (k *Kubernetes) ProvisionPostScripts() error {
	c, err := k.newClient()
	if err != nil {
		return err
	}

	return k.provisionPostScripts(c)
}

func (k *Kubernetes) defaultStandardPackageRepositoryImage() (string, error) {
	repo := &model.StandardPackageRepository{}
	if err := mapd.Database().Where("`default` = ?", true).Order("`updated_at` DESC").Take(repo).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}

		u, err := url.Parse(k.PrimaryImageRegistry.Endpoint)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%s", filepath.Join(u.Host, tanzu.TKGDefaultStandardRepo), tanzu.TKGDefaultStandardRepoVersion), nil
	}
	return repo.Image, nil
}

func (k *Kubernetes) provisionPostScripts(c *k8s.Kubernetes) error {
	defer k.log.Sync()

	// cluster initiator is indeed a very special extension.
	initiator := newClusterInitiator(k.Cluster.Name, defNamespace, c, k.PrimaryImageRegistry, k.log)
	if err := initiator.Install(); err != nil {
		if extension.IsCriticalError(err) {
			return &ErrCriticalFailure{err}
		}
		return err
	}

	expectedRepoImage, err := k.defaultStandardPackageRepositoryImage()
	if err != nil {
		k.log.Errorf("failed to get standard package repository image for cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
		return err
	}
	k.log.Infof("expected standard package repository image for cluster %s is %s.", strconv.Quote(k.Cluster.Name), expectedRepoImage)

	k.log.Infof("checking state of package repository for cluster %s...", strconv.Quote(k.Cluster.Name))
	kapp := k8s.NewKappController(c)

	repoList, err := kapp.GetPackageRepositories("", k8s.FieldEqual("metadata.name", "tanzu-standard"))
	if err != nil {
		k.log.Errorf("failed to retrieve package repository for cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
		return err
	}
	var repo *packagingv1alpha1.PackageRepository
	if len(repoList.Items) < 1 {
		repoNS := "tanzu-package-repo-global"
		_, err = kapp.GetNamespace(repoNS)
		if err != nil {
			if !k8s.IsErrObjectNotExists(err) {
				k.log.Errorf("failed to check namespace for cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
				return err
			}
			repoNS = defNamespace
		}

		repoSpec := packagingv1alpha1.PackageRepositorySpec{
			Fetch: &packagingv1alpha1.PackageRepositoryFetch{
				ImgpkgBundle: &kappctrlv1alpha1.AppFetchImgpkgBundle{
					Image: expectedRepoImage,
				},
			},
		}

		repo = &packagingv1alpha1.PackageRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "tanzu-standard",
				Namespace: repoNS,
			},
			Spec: repoSpec,
		}
		if repo, err = kapp.CreatePackageRepository(repo); err != nil {
			k.log.Errorf("failed to create package repository for cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
			return err
		}

		k.log.Infof("package repository %s/%s is creating on cluster %s.", repo.Namespace, repo.Name, strconv.Quote(k.Cluster.Name))
		return fmt.Errorf("package repository %s/%s is not ready", repo.Namespace, repo.Name)
	} else {
		repo = &repoList.Items[0]
	}

	newRepo := repo.DeepCopy()
	if newRepo.Spec.Fetch == nil {
		newRepo.Spec.Fetch = &packagingv1alpha1.PackageRepositoryFetch{
			ImgpkgBundle: &kappctrlv1alpha1.AppFetchImgpkgBundle{
				Image: expectedRepoImage,
			},
		}
	} else if newRepo.Spec.Fetch.ImgpkgBundle == nil {
		newRepo.Spec.Fetch.ImgpkgBundle = &kappctrlv1alpha1.AppFetchImgpkgBundle{
			Image: expectedRepoImage,
		}
		newRepo.Spec.Fetch.Image = nil
		newRepo.Spec.Fetch.HTTP = nil
		newRepo.Spec.Fetch.Git = nil
	} else if newRepo.Spec.Fetch.ImgpkgBundle.Image != expectedRepoImage {
		newRepo.Spec.Fetch.ImgpkgBundle.Image = expectedRepoImage
		newRepo.Spec.Fetch.Image = nil
		newRepo.Spec.Fetch.HTTP = nil
		newRepo.Spec.Fetch.Git = nil
	}
	patch, err := jsondiff.Compare(repo, newRepo)
	if err != nil {
		k.log.Errorf("failed to compare package repository for cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
		return err
	}
	if len(patch) > 0 {
		k.log.Infof("package repository %s/%s requires update on cluster %s.", repo.Namespace, repo.Name, strconv.Quote(k.Cluster.Name))

		p, err := json.Marshal(patch)
		if err != nil {
			k.log.Errorf("failed to generate patch for package repository %s/%s on cluster %s due to: %v.", repo.Namespace, repo.Name, strconv.Quote(k.Cluster.Name), err)
			return err
		}

		_, err = kapp.PatchPackageRepository(repo.Namespace, repo.Name, types.JSONPatchType, p)
		if err != nil {
			k.log.Errorf("failed to patch for package repository %s/%s on cluster %s due to: %v.", repo.Namespace, repo.Name, strconv.Quote(k.Cluster.Name), err)
			return err
		}

		k.log.Infof("package repository %s/%s is updating on cluster %s.", repo.Namespace, repo.Name, strconv.Quote(k.Cluster.Name))
		return fmt.Errorf("package repository %s/%s is not ready", repo.Namespace, repo.Name)
	}

	var repoReady bool
	for _, cond := range repo.Status.Conditions {
		if cond.Type == kappctrlv1alpha1.ReconcileSucceeded && cond.Status == corev1.ConditionTrue {
			repoReady = true
			break
		}
	}
	if !repoReady {
		k.log.Warnf("waiting for package repository %s/%s to be ready for proceeding provision of cluster %s.", repo.GetNamespace(), repo.GetName(), strconv.Quote(k.Cluster.Name))
		return fmt.Errorf("package repository is not ready")
	}
	k.log.Infof("package repository %s/%s is ready on cluster %s.", repo.Namespace, repo.Name, strconv.Quote(k.Cluster.Name))
	return nil
}

func (k *Kubernetes) ProvisionFinish() error {
	c, err := k.newClient()
	if err != nil {
		return err
	}
	return k.provisionFinish(c)
}

func (k *Kubernetes) provisionFinish(c *k8s.Kubernetes) error {
	defer k.log.Sync()

	k.log.Infof("checking state of cert manager for cluster %s...", strconv.Quote(k.Cluster.Name))
	certCRD, err := searchCertManagerCertificateCRD(k8s.NewAPIExtensions(c))
	if err != nil {
		k.log.Errorf("failed to retrieve certificate CRD of cert manager for cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
		return err
	}

	kapp := k8s.NewKappController(c)

	installedCertManagers, err := getInstalledCertManagers(kapp)
	if err != nil {
		k.log.Errorf("failed to retrieve state of cert manager for cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
		return err
	}

	var certManager *packagingv1alpha1.PackageInstall
	if len(installedCertManagers) < 1 && certCRD == nil {
		latestPackage, err := getLatestAvailableCertManager(kapp)
		if err != nil {
			k.log.Errorf("failed to retrieve latest available version of cert manager for cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
			return err
		}

		certManager = &packagingv1alpha1.PackageInstall{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cert-manager",
				Namespace: defNamespace,
			},
			Spec: packagingv1alpha1.PackageInstallSpec{
				PackageRef: &packagingv1alpha1.PackageRef{
					RefName: latestPackage.Spec.RefName,
					VersionSelection: &versionsv1alpha1.VersionSelectionSemver{
						Constraints: latestPackage.Spec.Version,
					},
				},
				ServiceAccountName: defServiceAccount,
				SyncPeriod:         &metav1.Duration{Duration: time.Minute * 5},
			},
		}
		if _, err := kapp.CreatePackageInstall(certManager); err != nil {
			k.log.Errorf("unable to deploy package for cert manager on cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
			return err
		}
	} else if certCRD != nil && len(installedCertManagers) < 1 {
		k.log.Warn("cert manager is already enabled on cluster %s, but we cannot validate its installation.", strconv.Quote(k.Cluster.Name))
	} else if len(installedCertManagers) > 0 {
		certManager = installedCertManagers[0]
	}

	if certManager != nil {
		var pkgReady bool
	LOOP:
		for _, cond := range certManager.Status.Conditions {
			switch cond.Type {
			case kappctrlv1alpha1.ReconcileSucceeded:
				pkgReady = true
				break LOOP
			case kappctrlv1alpha1.ReconcileFailed:
				k.log.Errorf("failed to deploy package %s/%s on cluster %s due to: %s.", certManager.GetNamespace(), certManager.GetName(), strconv.Quote(k.Cluster.Name), certManager.Status.FriendlyDescription)

				if certManager.Spec.PackageRef == nil || certManager.Spec.PackageRef.VersionSelection == nil || certManager.Spec.PackageRef.VersionSelection.Constraints != "1.5.3+vmware.2-tkg.1" {
					k.log.Infof("cleaning up the failed package %s/%s from cluster %s.", certManager.GetNamespace(), certManager.GetName(), strconv.Quote(k.Cluster.Name))
					if err := kapp.DeletePackageInstall(certManager.GetNamespace(), certManager.GetName()); err != nil {
						k.log.Errorf("failed to cleanup package for cert manager on cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
						return err
					}
					return fmt.Errorf("package is not created")
				}

				k.log.Errorf("Provision will be stopped on cluster %s due to an unexpected error. Before proceeding the provision again, please solve the problem manually and remove the failed package installation with command: \"kubectl delete packageinstall %s -n %s\".", strconv.Quote(k.Cluster.Name), certManager.GetName(), certManager.GetNamespace())
				return &ErrCriticalFailure{fmt.Errorf("failed to deploy cert-manager")}
			}
		}
		if !pkgReady {
			k.log.Infof("waiting for package %s/%s to be ready on cluster %s.", certManager.GetNamespace(), certManager.GetName(), strconv.Quote(k.Cluster.Name))
			return fmt.Errorf("package is not ready")
		}

		k.log.Infof("package %s/%s is ready on cluster %s now.", certManager.GetNamespace(), certManager.GetName(), strconv.Quote(k.Cluster.Name))
	}

	apiext := k8s.NewAPIExtensions(c)
	for _, each := range []*apiextensionsv1.CustomResourceDefinition{
		mapv1alpha1.CRDBinaryImage,
		mapv1alpha1.CRDDockerfileImage,
	} {
		k.log.Infof("installing API extension %s on cluster %s...", strconv.Quote(each.Name), strconv.Quote(k.Cluster.Name))
		if _, err := apiext.CreateCustomResourceDefinition(each); err != nil && !k8s.IsErrObjectAlreadyExists(err) {
			k.log.Errorf("failed to install API extension %s on cluster %s due to: %v.", strconv.Quote(each.Name), strconv.Quote(k.Cluster.Name), err)
			return &ErrCriticalFailure{fmt.Errorf("failed to install API extension: %v", err)}
		}
	}
	return nil
}

func (k *Kubernetes) CheckHealth() (model.ClusterState, error) {
	c, err := k.newClient()
	if err != nil {
		return model.ClusterStateUnknown, err
	}
	return model.ClusterStateRunning, c.Ping()
}

func (k *Kubernetes) GetTrustedCACerts() ([]byte, error) {
	spec := &model.ClusterExternalSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	for _, each := range spec.RootCAs {
		buf.WriteString(strings.TrimSpace(each))
		buf.WriteRune('\n')
	}
	for _, each := range k.RelatedImageRegistries {
		if len(each.CACert) != 0 {
			buf.Write(bytes.TrimSpace(each.CACert))
			buf.WriteRune('\n')
		}
	}
	return buf.Bytes(), nil
}

func NewExternalKubernetes(cluster *model.Cluster, imageRegistries []model.ImageRegistry, logger *zap.SugaredLogger) *Kubernetes {
	return &Kubernetes{
		Cluster:                cluster,
		PrimaryImageRegistry:   &imageRegistries[0],
		RelatedImageRegistries: imageRegistries,
		log:                    logger,
	}
}

type clusterInitiator struct {
	*extension.GenericInstaller
}

func newClusterInitiator(clusterName, namespace string, k *k8s.Kubernetes, registry *model.ImageRegistry, log *zap.SugaredLogger) *clusterInitiator {
	imageTag := "latest"
	if mapd.Version != "" {
		imageTag = mapd.Version
	}
	return &clusterInitiator{extension.NewGenericInstaller(extension.GenericInstallerConfig{
		Kubernetes:       k,
		Extension:        nil,
		ImageRegistry:    registry,
		Cluster:          clusterName,
		JobNamespace:     namespace,
		Job:              "cluster-initiator",
		Image:            fmt.Sprintf("%s:%s", mapd.GlobalConfig.ClusterInitiatorImage, imageTag),
		UnmarshalEnvFunc: func(ext *model.ClusterExtension, env map[string]string) error { return nil },
	}, log)}
}

func getInstalledCertManagers(kapp *k8s.KappController) ([]*packagingv1alpha1.PackageInstall, error) {
	list, err := kapp.GetPackageInstalls("", k8s.FieldEqual("metadata.name", "cert-manager"))
	if err != nil {
		return nil, err
	}

	var installedCertManagers []*packagingv1alpha1.PackageInstall
	for i := range list.Items {
		each := &list.Items[i]
		if each.Spec.PackageRef == nil {
			continue
		}
		if each.Spec.PackageRef.RefName == "cert-manager.tanzu.vmware.com" {
			installedCertManagers = append(installedCertManagers, each)
		}
	}
	return installedCertManagers, nil
}

func searchCertManagerCertificateCRD(k *k8s.APIExtensions) (*apiextensionsv1.CustomResourceDefinition, error) {
	list, err := k.GetCustomResourceDefinitions(k8s.FieldEqual("metadata.name", "certificates.cert-manager.io"))
	if err != nil {
		return nil, err
	}

	if len(list.Items) > 0 {
		return &list.Items[0], nil
	}
	return nil, nil
}

func getLatestAvailableCertManager(tkgm *k8s.KappController) (*datapackagingv1alpha1.Package, error) {
	list, err := tkgm.GetPackages("", k8s.FieldEqual("spec.refName", "cert-manager.tanzu.vmware.com"))
	if err != nil {
		return nil, err
	}
	if len(list.Items) < 1 {
		return nil, nil
	}
	sort.Slice(list.Items, func(i, j int) bool {
		pkg1, pkg2 := list.Items[i], list.Items[j]
		v1, err := semver.NewVersion(pkg1.Spec.Version)
		if err != nil {
			return false
		}
		v2, err := semver.NewVersion(pkg2.Spec.Version)
		if err != nil {
			return false
		}
		return v1.GreaterThan(v2)
	})
	return &list.Items[0], nil
}
