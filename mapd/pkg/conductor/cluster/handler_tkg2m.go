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
)

// TKG2MKubernetes implements Handler of TKG2M provisioned cluster.
type TKG2MKubernetes struct {
	*Kubernetes

	*model.ClusterProvider
}

func (k *TKG2MKubernetes) newTanzuClient() (*tanzuv2.TanzuKubernetesGrid, error) {
	spec := &model.ClusterProviderTKG2MSpec{}
	if err := json.Unmarshal(k.ClusterProvider.Spec, spec); err != nil {
		return nil, err
	}

	kubeConfig := spec.KubeConfig
	if len(spec.KubeConfig) > 0 {
		block, err := cryptoutil.NewAESGCMCipher(k.ClusterProvider.ID[:]) // AES-128
		if err != nil {
			return nil, err
		}
		kubeConfig = block.Decrypt(spec.KubeConfig)
	}
	return tanzuv2.NewTanzuKubernetesGrid(kubeConfig)
}

func (k *TKG2MKubernetes) ProvisionPrepare() error {
	return nil
}

func (k *TKG2MKubernetes) ProvisionKubernetes() (string, model.ClusterState, []byte, error) {
	sc, err := k.newTanzuClient()
	if err != nil {
		return "", model.ClusterStateUnknown, nil, err
	}

	defer k.log.Sync()

	spec := &model.ClusterTKG2MSpec{}
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
			_, err = sc.GetNamespace(spec.Namespace)
			if err != nil {
				if !k8s.IsErrObjectNotExists(err) {
					k.log.Errorf("failed to check cluster namespace %s due to: %v.", spec.Namespace, err)
					return "", model.ClusterStateUnknown, nil, err
				}
				_, err = createNamespace(sc.Kubernetes, spec.Namespace)
				if err != nil && !k8s.IsErrObjectAlreadyExists(err) {
					k.log.Errorf("failed to create cluster namespace %s due to: %v.", spec.Namespace, err)
					return "", model.ClusterStateUnknown, nil, err
				}
			}

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

func (k *TKG2MKubernetes) ProvisionNamespace() error {
	defer k.log.Sync()

	mc, err := k.newTanzuClient()
	if err != nil {
		return err
	}

	spec := &model.ClusterTKGMSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	c, err := mc.GetClusterClient(spec.Namespace, k.Cluster.Name)
	if err != nil {
		return err
	}
	if err = k.Kubernetes.provisionNamespace(c); err != nil {
		return err
	}

	if len(spec.SSHPrivateKey) < 1 {
		k.log.Warnf("SSH private key was not specified. Skipped for creating a mirrored secret in cluster %s.", strconv.Quote(k.Cluster.Name))
		return nil
	}

	k.log.Infof("checking ssh auth secret %s/%s state...", defNamespace, defSSHPrivateKeySecret)
	secret, err := c.GetSecret(defNamespace, defSSHPrivateKeySecret)
	if err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			k.log.Errorf("unexpected error: %v.", err)
			return err
		}

		secret, err = createSSHPrivateKeySecret(c, defNamespace, defSSHPrivateKeySecret, spec.SSHPrivateKey)
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
	if !ok || string(val) != string(spec.SSHPrivateKey) {
		secret.Data["ssh-privatekey"] = spec.SSHPrivateKey
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

func (k *TKG2MKubernetes) ProvisionServiceAccount() error {
	mc, err := k.newTanzuClient()
	if err != nil {
		return err
	}

	spec := &model.ClusterTKG2MSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	c, err := mc.GetClusterClient(spec.Namespace, k.Cluster.Name)
	if err != nil {
		return err
	}
	return k.Kubernetes.provisionServiceAccount(c)
}

func (k *TKG2MKubernetes) ProvisionClusterRoleBinding() (token []byte, err error) {
	mc, err := k.newTanzuClient()
	if err != nil {
		return nil, err
	}

	spec := &model.ClusterTKG2MSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		panic(err)
	}

	c, err := mc.GetClusterClient(spec.Namespace, k.Cluster.Name)
	if err != nil {
		return nil, err
	}
	return k.Kubernetes.provisionClusterRoleBinding(c)
}

func (k *TKG2MKubernetes) ProvisionValidation(namespace ...string) error {
	return k.Kubernetes.ProvisionValidation(append([]string{
		"avi-system",
		"kube-node-lease",
		"kube-public",
		"tanzu-package-repo-global",
		"tkg-system",
		"tkg-system-public",
	}, namespace...)...)
}

func (k *TKG2MKubernetes) ProvisionPostScripts() error {
	return k.Kubernetes.ProvisionPostScripts()
}

func (k *TKG2MKubernetes) ProvisionFinish() error {
	return k.Kubernetes.ProvisionFinish()
}

func (k *TKG2MKubernetes) CheckHealth() (model.ClusterState, error) {
	return k.Kubernetes.CheckHealth()
}

func (k *TKG2MKubernetes) GetTrustedCACerts() ([]byte, error) {
	spec := &model.ClusterTKG2MSpec{}
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

func NewTKG2MKubernetes(cluster *model.Cluster, imageRegistries []model.ImageRegistry, provider *model.ClusterProvider, logger *zap.SugaredLogger) *TKG2MKubernetes {
	return &TKG2MKubernetes{
		Kubernetes:      NewExternalKubernetes(cluster, imageRegistries, logger),
		ClusterProvider: provider,
	}
}
