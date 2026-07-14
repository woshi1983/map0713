// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	runv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/run.tanzu/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// TKGSKubernetes implements Handler of TKGS provisioned cluster.
type TKGSKubernetes struct {
	*Kubernetes

	*model.ClusterProvider
}

func (k *TKGSKubernetes) newSupervisorClient() (*k8s.TanzuKubernetesGridService, error) {
	spec := &model.ClusterProviderTKGSSpec{}
	if err := json.Unmarshal(k.ClusterProvider.Spec, spec); err != nil {
		return nil, err
	}
	block, err := cryptoutil.NewAESGCMCipher(k.ClusterProvider.ID[:]) // AES-128
	if err != nil {
		return nil, err
	}
	return k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
}

func (k *TKGSKubernetes) newClient() (*k8s.Kubernetes, error) {
	if len(k.Cluster.AuthToken) > 0 {
		return k.Kubernetes.newClient()
	}

	sc, err := k.newSupervisorClient()
	if err != nil {
		return nil, err
	}
	spec := &model.ClusterTKGSSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		return nil, err
	}
	token, err := sc.Login(spec.Namespace, k.Cluster.Name)
	if err != nil {
		return nil, err
	}
	return k8s.NewKubernetesFromToken(k.Cluster.Endpoint, []byte(token), rest.TLSClientConfig{
		Insecure: k.Cluster.InsecureTLS,
		CAData:   k.Cluster.CACert,
	})
}

func (k *TKGSKubernetes) newInsecureClient() (*k8s.Kubernetes, error) {
	if len(k.Cluster.AuthToken) > 0 {
		return k.Kubernetes.newClient()
	}

	sc, err := k.newSupervisorClient()
	if err != nil {
		return nil, err
	}
	spec := &model.ClusterTKGSSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		return nil, err
	}
	token, err := sc.Login(spec.Namespace, k.Cluster.Name)
	if err != nil {
		return nil, err
	}
	return k8s.NewKubernetesFromToken(k.Cluster.Endpoint, []byte(token), rest.TLSClientConfig{Insecure: true})
}

func (k *TKGSKubernetes) ProvisionPrepare() error {
	return nil
}

func (k *TKGSKubernetes) ProvisionKubernetes() (string, model.ClusterState, []byte, error) {
	sc, err := k.newSupervisorClient()
	if err != nil {
		return "", model.ClusterStateUnknown, nil, err
	}

	defer k.log.Sync()

	spec := &model.ClusterTKGSSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	k.log.Infof("checking state of cluster %s...", strconv.Quote(k.Cluster.Name))
	tkc, err := sc.GetTanzuKubernetesCluster(spec.Namespace, k.Cluster.Name)
	if err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("could not retrieve state of cluster %s: %v.", strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateUnknown, nil, err
		}

		k.log.Infof("cluster %s does not exist. A new tanzu kubernetes cluster will be provisioned.", strconv.Quote(k.Cluster.Name))

		tsc, err := sc.GetTkgServiceConfiguration(defTkgServiceConfiguration)
		if err != nil {
			k.log.Errorf("could not retrieve TkgServiceConfiguration %s for cluster %s due to: %v.", defTkgServiceConfiguration, strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateUnknown, nil, err
		}
		// provisioning
		tkc = generateTanzuKubernetesCluster(k.Cluster, spec.Namespace)
		tkc.Spec.Settings.Network.Trust = tsc.Spec.Trust
		if tkc.Spec.Settings.Network.Trust == nil {
			tkc.Spec.Settings.Network.Trust = &runv1alpha1.TrustConfiguration{}
		}
		for _, each := range k.RelatedImageRegistries {
			caCert := bytes.TrimSpace(each.CACert)
			if len(caCert) != 0 {
				tkc.Spec.Settings.Network.Trust.AdditionalTrustedCAs = append(tkc.Spec.Settings.Network.Trust.AdditionalTrustedCAs, runv1alpha1.TLSCertificate{
					Name: each.ID.String(),
					Data: base64.StdEncoding.EncodeToString(caCert),
				})
			}
		}

		_, err = sc.CreateTanzuKubernetesCluster(tkc)
		if err != nil {
			k.log.Errorf("failed to provision tanzu kubernetes cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateUnknown, nil, err
		}

		k.log.Infof("tanzu kubernetes cluster %s is spawning now.", strconv.Quote(k.Cluster.Name))
		return "", model.ClusterStateCreating, nil, nil
	}

	k.log.Debugf("found tanzu kubernetes cluster %s in namespace %s.", strconv.Quote(k.Cluster.Name), spec.Namespace)
	sameVersion, err := compairK8sVersion(tkc.Spec.Distribution.VersionHint, k.Cluster.Version)
	if err != nil {
		k.log.Errorf("failed to parsing cluster %s's version due to: %v.", strconv.Quote(k.Cluster.Name), err)
		return "", model.ClusterStateUnknown, nil, err
	}
	var requiresUpdate int
	if !sameVersion {
		tkc.Spec.Distribution.Version = ""
		tkc.Spec.Distribution.VersionHint = "v" + strings.TrimPrefix(k.Cluster.Version, "v")
		requiresUpdate++
	}
	if doesTKCRequiresReconfig(tkc, spec) {
		tkc.Spec.Topology.ControlPlane.Count = int32(spec.CtlPlaneCount)
		tkc.Spec.Topology.ControlPlane.Class = spec.CtlPlaneClass
		tkc.Spec.Topology.Workers.Count = int32(spec.DataPlaneCount)
		tkc.Spec.Topology.Workers.Class = spec.DataPlaneClass
		requiresUpdate++
	}
	if requiresUpdate > 0 {
		_, err := sc.UpdateTanzuKubernetesCluster(tkc)
		if err != nil {
			k.log.Errorf("failed to update tanzu kubernetes cluster %s due to: %v. It will be retried later.", strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateUnknown, nil, err
		}
		k.log.Infof("tanzu kubernetes cluster %s will be updated automatically.", strconv.Quote(k.Cluster.Name))
		return "", model.ClusterStateUpdating, nil, fmt.Errorf("cluster pending update")
	}

	// TODO: This property will be removed in v1alpha3
	switch tkc.Status.Phase {
	case runv1alpha1.TanzuKubernetesClusterPhaseCreating:
		k.log.Infof("tanzu kubernetes cluster %s is still spawning.", strconv.Quote(k.Cluster.Name))
		return "", model.ClusterStateCreating, nil, nil
	case runv1alpha1.TanzuKubernetesClusterPhaseRunning:
		if tkc.Status.ClusterAPIStatus == nil || len(tkc.Status.ClusterAPIStatus.APIEndpoints) == 0 {
			k.log.Infof("endpoint of tanzu kubernetes cluster %s was not provisioned yet, waiting for it to be available...", strconv.Quote(k.Cluster.Name))
			return "", model.ClusterStateRunning, nil, fmt.Errorf("lack of endpoint")
		}
		end := tkc.Status.ClusterAPIStatus.APIEndpoints[0]
		endpoint := fmt.Sprintf("https://%s:%d", end.Host, end.Port)
		k.Cluster.Endpoint = endpoint

		// check if there is necessary to update cluster.
		c, err := k.newInsecureClient()
		if err != nil {
			return endpoint, model.ClusterStateRunning, nil, err
		}
		nodeList, err := c.GetNodes()
		if err != nil {
			k.log.Errorf("failed to retrieve nodes of tanzu kubernetes cluster %s due to: %v. It will be retried later.", strconv.Quote(k.Cluster.Name), err)
			return endpoint, model.ClusterStateRunning, nil, err
		}
		nodeCount := int(tkc.Spec.Topology.ControlPlane.Count) + int(tkc.Spec.Topology.Workers.Count)
		if len(nodeList.Items) != nodeCount {
			k.log.Infof("nodes of tanzu kubernetes cluster %s do not match the desired count (%d/%d).", strconv.Quote(k.Cluster.Name), len(nodeList.Items), nodeCount)
			return endpoint, model.ClusterStateRunning, nil, fmt.Errorf("cluster is scaling")
		}

		version, err := c.Version()
		if err != nil {
			k.log.Errorf("failed to retrieve version of tanzu kubernetes cluster %s due to: %v.", strconv.Quote(k.Cluster.Name), err)
			return endpoint, model.ClusterStateRunning, nil, err
		}
		sameVersion, err := compairK8sVersion(version.GitVersion, k.Cluster.Version)
		if err != nil {
			k.log.Errorf("failed to parsing cluster version due to: %v.", err)
			return endpoint, model.ClusterStateRunning, nil, err
		}
		if !sameVersion {
			k.log.Infof("tanzu kubernetes cluster %s is pending update and its Kubernetes version does not match (%s/%s).", strconv.Quote(k.Cluster.Name), version.GitVersion, k.Cluster.Version)
			return endpoint, model.ClusterStateRunning, nil, fmt.Errorf("version pending update")
		}

		b, err := getTKGSClusterCA(sc, spec.Namespace, k.Cluster.Name)
		if err != nil {
			k.log.Errorf("failed to get CA certificate of cluster %s: %v", strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateRunning, nil, err
		}

		return endpoint, model.ClusterStateRunning, b, nil
	case runv1alpha1.TanzuKubernetesClusterPhaseUpdating:
		k.log.Infof("tanzu kubernetes cluster %s is in updating state.", strconv.Quote(k.Cluster.Name))
		return "", model.ClusterStateUpdating, nil, nil
	case runv1alpha1.TanzuKubernetesClusterPhaseFailed:
		k.log.Errorf("tanzu kubernetes cluster %s has failed to provision. You can check it manually with command: %s", strconv.Quote(k.Cluster.Name), strconv.Quote(fmt.Sprintf("kubectl get tkc %s -n %s", k.Cluster.Name, spec.Namespace)))
		return "", model.ClusterStateFailed, nil, &ErrCriticalFailure{fmt.Errorf("cluster requires user intervention")}
	case runv1alpha1.TanzuKubernetesClusterPhaseDeleting:
		k.log.Warnf("tanzu kubernetes cluster %s should be provisioning but was found DELETING!!! We assume that this is an existing cluster which is being deleted. Standby for creation...", strconv.Quote(k.Cluster.Name))
		return "", model.ClusterStateDeleting, nil, nil
	case runv1alpha1.TanzuKubernetesClusterPhaseEmpty:
		return "", model.ClusterStateCreating, nil, nil
	}
	return "", model.ClusterStateUnknown, nil, &ErrCriticalFailure{fmt.Errorf("cluster requires user intervention")}
}

func (k *TKGSKubernetes) ProvisionNamespace() error {
	defer k.log.Sync()

	sc, err := k.newSupervisorClient()
	if err != nil {
		return err
	}
	c, err := k.newClient()
	if err != nil {
		return err
	}

	if err := k.Kubernetes.provisionNamespace(c); err != nil {
		return err
	}

	spec := &model.ClusterTKGSSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	privateKeyName := k.Cluster.Name + "-ssh"
	k.log.Infof("retrieving secret %s/%s from supervisor cluster for cluster %s...", spec.Namespace, privateKeyName, strconv.Quote(k.Cluster.Name))
	ssecret, err := sc.GetSecret(spec.Namespace, privateKeyName)
	if err != nil {
		k.log.Errorf("failed to retrieve secret %s/%s on cluster %s due to: %v.", spec.Namespace, privateKeyName, strconv.Quote(k.Cluster.Name), err)
		return err
	}
	if ssecret.Type != corev1.SecretTypeSSHAuth {
		k.log.Warnf("secret %s/%s should be a %s, but it is %s. Skipped for creating a mirrored secret in cluster %s.", spec.Namespace, privateKeyName, corev1.SecretTypeSSHAuth, ssecret.Type, strconv.Quote(k.Cluster.Name))
		return nil
	}
	privateKey, ok := ssecret.Data["ssh-privatekey"]
	if !ok {
		k.log.Warnf("secret %s/%s was found, but it does not seem to be a valid SSH auth secret. Skipped for creating a mirrored secret in cluster %s.", spec.Namespace, privateKeyName, strconv.Quote(k.Cluster.Name))
		return nil
	}

	k.log.Infof("checking ssh auth secret %s/%s state...", defNamespace, defSSHPrivateKeySecret)
	secret, err := c.GetSecret(defNamespace, defSSHPrivateKeySecret)
	if err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("unexpected error: %v.", err)
			return err
		}

		secret, err = createSSHPrivateKeySecret(c, defNamespace, defSSHPrivateKeySecret, privateKey)
		if err != nil {
			k.log.Errorf("failed to create secret %s/%s on cluster %s due to: %v.", defNamespace, defSSHPrivateKeySecret, strconv.Quote(k.Cluster.Name), err)
			return err
		}
		k.log.Infof("SSH private key is saved as secret %s/%s on cluster %s.", secret.GetNamespace(), secret.GetName(), strconv.Quote(k.Cluster.Name))
		return nil
	}

	var updatesRequired int
	if secret.Type != corev1.SecretTypeSSHAuth {
		secret.Type = corev1.SecretTypeSSHAuth
		secret.Data = make(map[string][]byte)
		updatesRequired++
	}
	val, ok := secret.Data["ssh-privatekey"]
	if !ok || string(val) != string(privateKey) {
		secret.Data["ssh-privatekey"] = privateKey
		updatesRequired++
	}
	if updatesRequired > 0 {
		if _, err = c.UpdateSecret(secret); err != nil {
			k.log.Errorf("failed to update secret %s/%s on cluster %s due to: %v.", secret.GetNamespace(), secret.GetName(), strconv.Quote(k.Cluster.Name), err)
			return err
		}
		k.log.Warnf("SSH private key was changed on cluster %s.", strconv.Quote(k.Cluster.Name))
	}
	k.log.Infof("SSH private key is saved as secret %s/%s on cluster %s.", secret.GetNamespace(), secret.GetName(), strconv.Quote(k.Cluster.Name))
	return nil
}

func (k *TKGSKubernetes) ProvisionServiceAccount() error {
	c, err := k.newClient()
	if err != nil {
		return err
	}
	return k.Kubernetes.provisionServiceAccount(c)
}

func (k *TKGSKubernetes) ProvisionClusterRoleBinding() (token []byte, err error) {
	c, err := k.newClient()
	if err != nil {
		return nil, err
	}

	defer k.log.Sync()

	for _, each := range []string{defClusterAdminClusterRoleBinding, defPspPrivilegedClusterRoleBinding} {
		k.log.Infof("checking cluster role binding %s state on tanzu kubernetes cluster...", each)
		if _, err := c.GetClusterRoleBinding(each); err != nil {
			if !k8s.IsErrObjectNotExists(err) {
				k.log.Errorf("unexpected error: %v.", err)
				return nil, err
			}
			k.log.Infof("cluster role binding %s does not exist on cluster %s, now creating...", each, strconv.Quote(k.Cluster.Name))
			switch each {
			case defClusterAdminClusterRoleBinding:
				if _, err := createClusterRoleBindingClusterAdmin(c, each, defNamespace, defServiceAccount); err != nil {
					k.log.Errorf("failed to provision cluster role binding %s on cluster %s due to: %v.", each, strconv.Quote(k.Cluster.Name), err)
					return nil, err
				}
			case defPspPrivilegedClusterRoleBinding:
				if _, err := createClusterRoleBindingPSPPrivileged(c, each); err != nil {
					k.log.Errorf("failed to provision cluster role binding %s on cluster %s due to: %v.", each, strconv.Quote(k.Cluster.Name), err)
					return nil, err
				}
			}
			k.log.Infof("cluster role binding %s is created and assigned to service account %s/%s.", each, defNamespace, defServiceAccount)
		} else {
			k.log.Warnf("cluster role binding %s already exists.", each)
		}
	}

	k.log.Infof("checking service account %s state on tanzu kubernetes cluster...", defServiceAccount)
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
		k.log.Infof("secret %s/%s seems to be creating as we cannot find any data on it, and we will check it again later.", defNamespace, secret.Name)
		return nil, fmt.Errorf("token of service account is missing")
	}
	token, ok := secret.Data["token"]
	if !ok {
		k.log.Warnf("data was found but token is missing on secret %s/%s, and we will check it again later.", defNamespace, secret.Name)
		return nil, fmt.Errorf("token of service account is missing")
	}
	return token, nil
}

func (k *TKGSKubernetes) ProvisionValidation(namespace ...string) error {
	return k.Kubernetes.ProvisionValidation("vmware-system-auth", "vmware-system-cloud-provider", "vmware-system-csi")
}

func (k *TKGSKubernetes) ProvisionPostScripts() error {
	return k.Kubernetes.ProvisionPostScripts()
}

func (k *TKGSKubernetes) ProvisionFinish() error {
	return k.Kubernetes.ProvisionFinish()
}

func (k *TKGSKubernetes) CheckHealth() (model.ClusterState, error) {
	sc, err := k.newSupervisorClient()
	if err != nil {
		return model.ClusterStateUnknown, err
	}

	defer k.log.Sync()

	spec := &model.ClusterTKGSSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	tkc, err := sc.GetTanzuKubernetesCluster(spec.Namespace, k.Cluster.Name)
	if err != nil {
		k.log.Errorf("could not retrieve state of cluster %s: %v", strconv.Quote(k.Cluster.Name), err)
		return model.ClusterStateUnknown, err
	}

	if tkc.Status.Phase == runv1alpha1.TanzuKubernetesClusterPhaseEmpty {
		k.log.Warnf("could not find any information about cluster's current status.")
		return model.ClusterStateUnknown, fmt.Errorf("lack of cluster state")
	}

	switch tkc.Status.Phase {
	case runv1alpha1.TanzuKubernetesClusterPhaseCreating:
		k.log.Warnf("tanzu kubernetes cluster %s should be provisioned but it was found being spawning and creating.", strconv.Quote(k.Cluster.Name))
		return model.ClusterStateCreating, fmt.Errorf("cluster being creating")
	case runv1alpha1.TanzuKubernetesClusterPhaseRunning:
	case runv1alpha1.TanzuKubernetesClusterPhaseUpdating:
		k.log.Infof("tanzu kubernetes cluster %s is in updating state.", strconv.Quote(k.Cluster.Name))
		return model.ClusterStateUpdating, fmt.Errorf("cluster being updating")
	case runv1alpha1.TanzuKubernetesClusterPhaseFailed:
		k.log.Warnf("tanzu kubernetes cluster %s is in failed state for some reason. You can check it manually with command: %s", strconv.Quote(k.Cluster.Name), strconv.Quote(fmt.Sprintf("kubectl get tkc %s -n %s", k.Cluster.Name, spec.Namespace)))
		return model.ClusterStateFailed, fmt.Errorf("cluster being failed")
	case runv1alpha1.TanzuKubernetesClusterPhaseDeleting:
		k.log.Warnf("tanzu kubernetes cluster %s was found DELETING!!! which will not happen in common cases. Administrators shall prevent operating those managed tanzu kubernetes clusters directly.", strconv.Quote(k.Cluster.Name))
		return model.ClusterStateDeleting, fmt.Errorf("cluster being deleting")
	default:
		k.log.Warnf("tanzu kubernetes cluster %s is in unknown state: %s.", strconv.Quote(k.Cluster.Name), tkc.Status.Phase)
		return model.ClusterStateUnknown, fmt.Errorf("cluster state unknown")
	}

	if _, err := k.Kubernetes.CheckHealth(); err != nil {
		return model.ClusterStateRunning, err
	}
	return model.ClusterStateRunning, nil
}

func (k *TKGSKubernetes) GetTrustedCACerts() ([]byte, error) {
	spec := &model.ClusterTKGSSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	sc, err := k.newSupervisorClient()
	if err != nil {
		return nil, err
	}
	tkc, err := sc.GetTanzuKubernetesCluster(spec.Namespace, k.Cluster.Name)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if tkc.Spec.Settings != nil && tkc.Spec.Settings.Network != nil && tkc.Spec.Settings.Network.Trust != nil {
		for _, each := range tkc.Spec.Settings.Network.Trust.AdditionalTrustedCAs {
			b, err := base64.StdEncoding.DecodeString(each.Data)
			if err != nil {
				continue
			}
			buf.Write(bytes.TrimSpace(b))
			buf.WriteRune('\n')
		}
	}
	for _, each := range k.RelatedImageRegistries {
		if len(each.CACert) != 0 {
			buf.Write(bytes.TrimSpace(each.CACert))
			buf.WriteRune('\n')
		}
	}
	return buf.Bytes(), nil
}

func NewTKGSKubernetes(cluster *model.Cluster, imageRegistries []model.ImageRegistry, provider *model.ClusterProvider, logger *zap.SugaredLogger) *TKGSKubernetes {
	return &TKGSKubernetes{
		Kubernetes:      NewExternalKubernetes(cluster, imageRegistries, logger),
		ClusterProvider: provider,
	}
}

func doesTKCRequiresReconfig(tkc *runv1alpha1.TanzuKubernetesCluster, spec *model.ClusterTKGSSpec) bool {
	if int64(tkc.Spec.Topology.ControlPlane.Count) != spec.CtlPlaneCount {
		return true
	}
	if tkc.Spec.Topology.ControlPlane.Class != spec.CtlPlaneClass {
		return true
	}
	if int64(tkc.Spec.Topology.Workers.Count) != spec.DataPlaneCount {
		return true
	}
	if tkc.Spec.Topology.Workers.Class != spec.DataPlaneClass {
		return true
	}
	return false
}

func compairK8sVersion(ver1, ver2 string) (bool, error) {
	v1, err := semver.NewVersion(ver1)
	if err != nil {
		return false, err
	}
	v2, err := semver.NewVersion(ver2)
	if err != nil {
		return false, err
	}

	v1p, err := v1.SetMetadata("")
	if err != nil {
		return false, err
	}
	v1p, err = v1p.SetPrerelease("")
	if err != nil {
		return false, err
	}

	v2p, err := v2.SetMetadata("")
	if err != nil {
		return false, err
	}
	v2p, err = v2p.SetPrerelease("")
	if err != nil {
		return false, err
	}

	return v1p.Equal(&v2p), nil
}

func getTKGSClusterCA(k *k8s.TanzuKubernetesGridService, namespace, name string) ([]byte, error) {
	secret, err := k.GetSecret(namespace, name+"-ca")
	if err != nil {
		return nil, err
	}
	if secret.Data != nil && secret.Data["tls.crt"] != nil {
		return secret.Data["tls.crt"], nil
	}
	return nil, fmt.Errorf("cluster CA not found")
}
