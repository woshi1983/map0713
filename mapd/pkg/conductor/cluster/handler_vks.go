package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/vks"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
)

type VKSKubernetes struct {
	*Kubernetes
	*model.ClusterProvider
}

func NewVKSKubernetes(c *model.Cluster, registries []model.ImageRegistry, provider *model.ClusterProvider, log *zap.SugaredLogger) *VKSKubernetes {
	return &VKSKubernetes{Kubernetes: NewExternalKubernetes(c, registries, log), ClusterProvider: provider}
}

func (k *VKSKubernetes) spec() (*model.ClusterVKSSpec, error) {
	spec := &model.ClusterVKSSpec{}
	if err := json.Unmarshal(k.Cluster.Spec, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func (k *VKSKubernetes) supervisor() (*vks.Client, error) {
	spec := &model.ClusterProviderVKSSpec{}
	if err := json.Unmarshal(k.ClusterProvider.Spec, spec); err != nil {
		return nil, err
	}
	client, _, err := vks.NewClientFromProvider(k.ClusterProvider.ID, spec)
	return client, err
}

func (k *VKSKubernetes) ProvisionKubernetes() (string, model.ClusterState, []byte, error) {
	spec, err := k.spec()
	if err != nil {
		return "", model.ClusterStateUnknown, nil, err
	}
	client, err := k.supervisor()
	if err != nil {
		return "", model.ClusterStateUnknown, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()
	release, err := client.ResolveRelease(ctx, k.Cluster.Version)
	if err != nil {
		return "", model.ClusterStateUnknown, nil, err
	}
	obj, err := client.CreateCluster(ctx, k.Cluster.ID, k.Cluster.Name, release, *spec)
	if err != nil {
		return "", model.ClusterStateUnknown, nil, err
	}
	if spec.RemoteUID == "" {
		spec.RemoteUID = string(obj.GetUID())
		if len(spec.RemoteUID) == 0 {
			return "", model.ClusterStateCreating, nil, nil
		}
		k.Cluster.Spec, err = json.Marshal(spec)
		if err != nil {
			return "", model.ClusterStateUnknown, nil, err
		}
	}
	obj, err = client.GetClusterWithStatus(ctx, spec.Namespace, k.Cluster.Name)
	if err != nil {
		return "", model.ClusterStateCreating, nil, err
	}
	if !vks.Available(obj) || !vks.ReplicasReady(obj) {
		return "", model.ClusterStateCreating, nil, nil
	}
	endpoint := vks.Endpoint(obj)
	if endpoint == "" {
		return "", model.ClusterStateCreating, nil, nil
	}
	ca, err := client.GetCA(ctx, spec.Namespace, k.Cluster.Name)
	if err != nil {
		return "", model.ClusterStateCreating, nil, err
	}
	config, err := client.GetKubeconfig(ctx, spec.Namespace, k.Cluster.Name)
	if err != nil {
		return "", model.ClusterStateCreating, nil, err
	}
	workload, err := k8s.NewKubernetesFromKubeConfig(config)
	if err != nil {
		return "", model.ClusterStateCreating, nil, err
	}
	if err := workload.PingContext(ctx); err != nil {
		return "", model.ClusterStateCreating, nil, err
	}
	return endpoint, model.ClusterStateRunning, ca, nil
}

func vksHealthState(current model.ClusterState, obj *unstructured.Unstructured) (model.ClusterState, error) {
	if vks.UpgradeInProgress(obj) || current == model.ClusterStateUpdating && (vks.ReconciliationPending(obj) || !vks.Available(obj) || !vks.ReplicasReady(obj)) {
		return model.ClusterStateUpdating, fmt.Errorf("VKS cluster topology is still reconciling")
	}
	if !vks.ReplicasReady(obj) {
		return model.ClusterStateCreating, fmt.Errorf("VKS cluster replicas are not ready")
	}
	if !vks.Available(obj) {
		return model.ClusterStateCreating, fmt.Errorf("VKS cluster is not available")
	}
	return model.ClusterStateRunning, nil
}

func (k *VKSKubernetes) CheckHealth() (model.ClusterState, error) {
	spec, err := k.spec()
	if err != nil {
		return model.ClusterStateUnknown, err
	}
	client, err := k.supervisor()
	if err != nil {
		return model.ClusterStateUnknown, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()
	obj, err := client.GetClusterWithStatus(ctx, spec.Namespace, k.Cluster.Name)
	if err != nil {
		return model.ClusterStateUnknown, err
	}
	if obj.GetLabels()[vks.OwnerLabel] != k.Cluster.ID.String() || spec.RemoteUID == "" || string(obj.GetUID()) != spec.RemoteUID {
		return model.ClusterStateUnknown, fmt.Errorf("VKS cluster ownership or UID mismatch")
	}
	state, err := vksHealthState(k.Cluster.State, obj)
	if err != nil {
		return state, err
	}
	return k.Kubernetes.CheckHealth()
}

func (k *VKSKubernetes) EnsureDefaultStorageClass() (bool, error) {
	spec, err := k.spec()
	if err != nil {
		return false, err
	}
	client, err := k.supervisor()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()
	return client.EnsureDefaultStorageClass(ctx, k.Cluster.ID, k.Cluster.Name, *spec)
}

func (k *VKSKubernetes) workloadClient() (*k8s.Kubernetes, error) {
	spec, err := k.spec()
	if err != nil {
		return nil, err
	}
	client, err := k.supervisor()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()
	config, err := client.GetKubeconfig(ctx, spec.Namespace, k.Cluster.Name)
	if err != nil {
		return nil, err
	}
	return k8s.NewKubernetesFromKubeConfig(config)
}

func (k *VKSKubernetes) ProvisionNamespace() error {
	c, err := k.workloadClient()
	if err != nil {
		return err
	}
	return k.Kubernetes.provisionNamespace(c)
}

func (k *VKSKubernetes) ProvisionServiceAccount() error {
	c, err := k.workloadClient()
	if err != nil {
		return err
	}
	return k.Kubernetes.provisionServiceAccount(c)
}

func (k *VKSKubernetes) ProvisionClusterRoleBinding() ([]byte, error) {
	c, err := k.workloadClient()
	if err != nil {
		return nil, err
	}
	// Current VKS versions do not auto-create long-lived service account
	// secrets. Create one explicitly for MAP's existing token-based access.
	client, err := kubernetes.NewForConfig(c.Config)
	if err != nil {
		return nil, err
	}
	secretName := defServiceAccount + "-token"
	secrets := client.CoreV1().Secrets(defNamespace)
	if _, err := secrets.Get(context.Background(), secretName, metav1.GetOptions{}); apierrors.IsNotFound(err) {
		_, err = secrets.Create(context.Background(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: defNamespace, Annotations: map[string]string{corev1.ServiceAccountNameKey: defServiceAccount}},
			Type:       corev1.SecretTypeServiceAccountToken,
		}, metav1.CreateOptions{})
	}
	if err != nil {
		return nil, err
	}
	sa, err := client.CoreV1().ServiceAccounts(defNamespace).Get(context.Background(), defServiceAccount, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	found := false
	for _, ref := range sa.Secrets {
		if ref.Name == secretName {
			found = true
			break
		}
	}
	if !found {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{Name: secretName})
		if _, err := client.CoreV1().ServiceAccounts(defNamespace).Update(context.Background(), sa, metav1.UpdateOptions{}); err != nil {
			return nil, err
		}
	}
	return k.Kubernetes.provisionClusterRoleBinding(c)
}

func (k *VKSKubernetes) GetTrustedCACerts() ([]byte, error) {
	spec, err := k.spec()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	for _, ca := range spec.TrustedCAs {
		buf.Write(bytes.TrimSpace(ca))
		buf.WriteByte('\n')
	}
	for _, registry := range k.RelatedImageRegistries {
		if len(registry.CACert) > 0 {
			buf.Write(bytes.TrimSpace(registry.CACert))
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes(), nil
}

var _ Handler = (*VKSKubernetes)(nil)
