package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"go.uber.org/zap"
)

var ErrTKGM1NoLongerSupport = fmt.Errorf("TKGm 1.X is no longer supported. Please consider to migrate your clusters to TKG2")

// TKGMKubernetes implements Handler of TKGM provisioned cluster.
type TKGMKubernetes struct {
	*Kubernetes

	*model.ClusterProvider
}

func (k *TKGMKubernetes) ProvisionPrepare() error {
	return nil
}

func (k *TKGMKubernetes) ProvisionKubernetes() (endpoint string, state model.ClusterState, caCert []byte, err error) {
	return "", model.ClusterStateUnknown, nil, &ErrCriticalFailure{ErrTKGM1NoLongerSupport}
}

func (k *TKGMKubernetes) ProvisionNamespace() error {
	return &ErrCriticalFailure{ErrTKGM1NoLongerSupport}
}

func (k *TKGMKubernetes) ProvisionServiceAccount() error {
	return &ErrCriticalFailure{ErrTKGM1NoLongerSupport}
}

func (k *TKGMKubernetes) ProvisionClusterRoleBinding() (token []byte, err error) {
	return nil, &ErrCriticalFailure{ErrTKGM1NoLongerSupport}
}

func (k *TKGMKubernetes) ProvisionValidation(namespace ...string) error {
	return k.Kubernetes.ProvisionValidation(append([]string{
		"avi-system",
		"kube-node-lease",
		"kube-public",
		"tanzu-package-repo-global",
		"tkg-system",
		"tkg-system-public",
	}, namespace...)...)
}

func (k *TKGMKubernetes) ProvisionPostScripts() error {
	c, err := k.newClient()
	if err != nil {
		return err
	}
	return k.Kubernetes.provisionPostScripts(c)
}

func (k *TKGMKubernetes) ProvisionFinish() error {
	return k.Kubernetes.ProvisionFinish()
}

func (k *TKGMKubernetes) CheckHealth() (model.ClusterState, error) {
	return k.Kubernetes.CheckHealth()
}

func (k *TKGMKubernetes) GetTrustedCACerts() ([]byte, error) {
	spec := &model.ClusterTKGMSpec{}
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

func NewTKGMKubernetes(cluster *model.Cluster, imageRegistries []model.ImageRegistry, provider *model.ClusterProvider, logger *zap.SugaredLogger) *TKGMKubernetes {
	return &TKGMKubernetes{
		Kubernetes:      NewExternalKubernetes(cluster, imageRegistries, logger),
		ClusterProvider: provider,
	}
}
