// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	tanzuv2 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/tanzu/v2"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// TKG2SKubernetes implements Handler of TKG2S provisioned cluster.
type TKG2SKubernetes struct {
	*Kubernetes

	*model.ClusterProvider
}

func (k *TKG2SKubernetes) newSupervisorClient() (*tanzuv2.VSphereWithTanzu, error) {
	spec := &model.ClusterProviderTKG2SSpec{}
	if err := json.Unmarshal(k.ClusterProvider.Spec, spec); err != nil {
		return nil, err
	}
	block, err := cryptoutil.NewAESGCMCipher(k.ClusterProvider.ID[:]) // AES-128
	if err != nil {
		return nil, err
	}
	return tanzuv2.NewVSphereWithTanzu(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
}

func (k *TKG2SKubernetes) ProvisionPrepare() error {
	return nil
}

func (k *TKG2SKubernetes) ProvisionKubernetes() (string, model.ClusterState, []byte, error) {
	sc, err := k.newSupervisorClient()
	if err != nil {
		return "", model.ClusterStateUnknown, nil, err
	}

	defer k.log.Sync()

	spec := &model.ClusterTKG2SSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}
	for _, each := range k.RelatedImageRegistries {
		if len(each.CACert) != 0 {
			spec.TrustedCAs = append(spec.TrustedCAs, each.CACert)
		}
	}

	k.log.Infof("checking state of cluster %s...", strconv.Quote(k.Cluster.Name))
	info, err := sc.GetClusterInfo(spec.Namespace, k.Cluster.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			k.log.Infof("cluster %s does not exist. A new workload cluster will be provisioned.", strconv.Quote(k.Cluster.Name))
			// provisioning
			p, err := sc.CreateCluster(k.Cluster, spec, false)
			k.log.Debugf("apply cluster with the following manifest:\n%s\n", strings.ReplaceAll(string(p), "\n", "\n> "))
			if err != nil {
				k.log.Errorf("failed to provision workload cluster due to: %v.", err)
				k.log.Infof("cleanup cluster provision for %s...", strconv.Quote(k.Cluster.Name))
				if err := sc.DeleteCluster(spec.Namespace, k.Cluster.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
					k.log.Errorf("failed to delete workload cluster due to: %v.", err)
				}
				return "", model.ClusterStateUnknown, nil, err
			}

			k.log.Infof("workload cluster %s is spawning now.", strconv.Quote(k.Cluster.Name))
			return "", model.ClusterStateCreating, nil, nil
		}
		k.log.Errorf("could not retrieve state of cluster %s: %v.", strconv.Quote(k.Cluster.Name), err)
		return "", model.ClusterStateUnknown, nil, err
	}

	k.log.Debugf("found workload cluster %s in namespace %s.", strconv.Quote(k.Cluster.Name), spec.Namespace)

	switch tanzuv2.TKGClusterPhase(info.Status) {
	case tanzuv2.TKGClusterPhaseEmpty, tanzuv2.TKGClusterPhaseCreating:
		k.log.Infof("workload cluster %s is still spawning.", strconv.Quote(k.Cluster.Name))
		return "", model.ClusterStateCreating, nil, nil
	case tanzuv2.TKGClusterPhaseRunning:
		updated, err := sc.EZUpdateClusterOnDemand(k.Cluster.Name, k.Cluster.Version, spec)
		if err != nil {
			k.log.Errorf("failed to update workload cluster %s on demand: %v", strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateRunning, nil, err
		}

		if updated {
			k.log.Infof("workload cluster %s is updating now.", strconv.Quote(k.Cluster.Name))
			return "", model.ClusterStateUpdating, nil, nil
		}

		_, actualMatch, err := sc.DoesControlPlaneVersionMatch(spec.Namespace, k.Cluster.Name, k.Cluster.Version)
		if err != nil {
			k.log.Errorf("failed to check status for workload cluster %s: %v", strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateRunning, nil, err
		}
		if !actualMatch {
			k.log.Infof("cluster %s is pending update of control plane.", strconv.Quote(k.Cluster.Name))
			return "", model.ClusterStateUpdating, nil, nil
		}

		c, err := sc.GetCluster(spec.Namespace, k.Cluster.Name)
		if err != nil {
			k.log.Errorf("failed to get state of cluster %s: %v", strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateRunning, nil, err
		}

		b, err := getTKGClusterCA(sc.Kubernetes, spec.Namespace, k.Cluster.Name)
		if err != nil {
			k.log.Errorf("failed to get CA certificate of cluster %s: %v", strconv.Quote(k.Cluster.Name), err)
			return "", model.ClusterStateRunning, nil, err
		}
		return "https://" + c.Spec.ControlPlaneEndpoint.String(), model.ClusterStateRunning, b, nil
	case tanzuv2.TKGClusterPhaseFailed:
		k.log.Errorf("workload cluster %s entered %s state.", strconv.Quote(k.Cluster.Name), info.Status)
		return "", model.ClusterStateUnknown, nil, &ErrCriticalFailure{fmt.Errorf("cluster is in failed phase")}
	case tanzuv2.TKGClusterPhaseUpdating:
		k.log.Infof("workload cluster %s is still updating.", strconv.Quote(k.Cluster.Name))
		return "", model.ClusterStateUpdating, nil, nil
	case tanzuv2.TKGClusterPhaseDeleting:
		k.log.Warnf("workload cluster %s is deleting. We assume that this is an existing cluster which is being deleted. Standby for creation...", strconv.Quote(k.Cluster.Name))
		return "", model.ClusterStateDeleting, nil, nil
	}
	return "", model.ClusterStateUnknown, nil, fmt.Errorf("cluster status is currently unknown")
}

func (k *TKG2SKubernetes) ProvisionNamespace() error {
	defer k.log.Sync()

	sc, err := k.newSupervisorClient()
	if err != nil {
		return err
	}

	spec := &model.ClusterTKG2SSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	c, err := sc.GetClusterClient(spec.Namespace, k.Cluster.Name)
	if err != nil {
		return err
	}
	if err = k.Kubernetes.provisionNamespace(c); err != nil {
		return err
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

func (k *TKG2SKubernetes) ProvisionServiceAccount() error {
	sc, err := k.newSupervisorClient()
	if err != nil {
		return err
	}

	spec := &model.ClusterTKG2SSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	c, err := sc.GetClusterClient(spec.Namespace, k.Cluster.Name)
	if err != nil {
		return err
	}
	return k.Kubernetes.provisionServiceAccount(c)
}

func (k *TKG2SKubernetes) ProvisionClusterRoleBinding() (token []byte, err error) {
	sc, err := k.newSupervisorClient()
	if err != nil {
		return nil, err
	}

	spec := &model.ClusterTKG2SSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	c, err := sc.GetClusterClient(spec.Namespace, k.Cluster.Name)
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
		k.log.Warnf("unable to find a valid token of service account %s/%s.", defNamespace, defServiceAccount)
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

func (k *TKG2SKubernetes) ProvisionValidation(namespace ...string) error {
	c, err := k.newClient()
	if err != nil {
		return err
	}
	return k.Kubernetes.provisionValidation(c, "secretgen-controller", "vmware-system-auth", "vmware-system-cloud-provider", "vmware-system-csi", "vmware-system-tkg")
}

func (k *TKG2SKubernetes) ProvisionPostScripts() error {
	return k.Kubernetes.ProvisionPostScripts()
}

func (k *TKG2SKubernetes) ProvisionFinish() error {
	return k.Kubernetes.ProvisionFinish()
}

func (k *TKG2SKubernetes) CheckHealth() (model.ClusterState, error) {
	return k.Kubernetes.CheckHealth()
}

func (k *TKG2SKubernetes) GetTrustedCACerts() ([]byte, error) {
	spec := &model.ClusterTKG2SSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	for _, each := range spec.TrustedCAs {
		buf.Write(bytes.TrimSpace(each))
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

func NewTKG2SKubernetes(cluster *model.Cluster, imageRegistries []model.ImageRegistry, provider *model.ClusterProvider, logger *zap.SugaredLogger) *TKG2SKubernetes {
	return &TKG2SKubernetes{
		Kubernetes:      NewExternalKubernetes(cluster, imageRegistries, logger),
		ClusterProvider: provider,
	}
}

func getTKGClusterCA(k *k8s.Kubernetes, namespace, name string) ([]byte, error) {
	secret, err := k.GetSecret(namespace, name+"-ca")
	if err != nil {
		return nil, err
	}
	if secret.Data != nil && secret.Data["tls.crt"] != nil {
		return secret.Data["tls.crt"], nil
	}
	return nil, fmt.Errorf("cluster CA not found")
}
