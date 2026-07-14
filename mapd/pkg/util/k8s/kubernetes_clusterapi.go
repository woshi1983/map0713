// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"bytes"
	"context"
	"text/template"

	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

type ClusterAPI struct {
	*Kubernetes

	tpl *template.Template
}

func (k *ClusterAPI) ApplyCluster(cfg ClusterTemplateConfig) ([]byte, error) {
	var buf bytes.Buffer
	if err := k.tpl.Execute(&buf, cfg); err != nil {
		return nil, &ClusterTemplateRenderingError{err}
	}

	if cfg.Dryrun {
		return buf.Bytes(), nil
	}
	manifest := buf.Bytes()
	return manifest, k.Kubernetes.Apply(manifest, mapd.GlobalConfig.HTTPClientTimeout)
}

func (k *ClusterAPI) UpdateCluster(obj *capiv1beta1.Cluster) (*capiv1beta1.Cluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.UpdateClusterContext(ctx, obj)
}

func (k *ClusterAPI) UpdateClusterContext(ctx context.Context, obj *capiv1beta1.Cluster) (*capiv1beta1.Cluster, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	result := &capiv1beta1.Cluster{}
	err = client.Put().
		Namespace(obj.Namespace).
		Resource("clusters").
		Name(obj.Name).
		VersionedParams(&metav1.UpdateOptions{}, ClusterAPIParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *ClusterAPI) PatchCluster(namespace, name string, pt types.PatchType, patch []byte) (*capiv1beta1.Cluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.PatchClusterContext(ctx, namespace, name, pt, patch)
}

func (k *ClusterAPI) PatchClusterContext(ctx context.Context, namespace, name string, pt types.PatchType, patch []byte) (*capiv1beta1.Cluster, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	result := &capiv1beta1.Cluster{}
	err = client.Patch(pt).
		Namespace(namespace).
		Resource("clusters").
		Name(name).
		VersionedParams(&metav1.PatchOptions{}, ClusterAPIParameterCodec).
		Body(patch).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *ClusterAPI) GetCluster(namespace, name string) (*capiv1beta1.Cluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetClusterContext(ctx, namespace, name)
}

func (k *ClusterAPI) GetClusterContext(ctx context.Context, namespace, name string) (*capiv1beta1.Cluster, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &capiv1beta1.Cluster{}
	err = client.Get().
		Namespace(namespace).
		Resource("clusters").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, ClusterAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *Kubernetes) GetClusters(namespace string, selectors ...Selector) (*capiv1beta1.ClusterList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetClustersContext(ctx, namespace, selectors...)
}

func (k *Kubernetes) GetClustersContext(ctx context.Context, namespace string, selectors ...Selector) (*capiv1beta1.ClusterList, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &capiv1beta1.ClusterList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("clusters").
		VersionedParams(&opts, ClusterAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *ClusterAPI) DeleteCluster(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.DeleteClusterContext(ctx, namespace, name)
}

func (k *ClusterAPI) DeleteClusterContext(ctx context.Context, namespace, name string) error {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	err = client.Delete().
		Namespace(namespace).
		Resource("clusters").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
	if err != nil {
		return err
	}
	return nil
}

func (k *ClusterAPI) GetClusterClass(namespace, name string) (*capiv1beta1.ClusterClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetClusterClassContext(ctx, namespace, name)
}

func (k *ClusterAPI) GetClusterClassContext(ctx context.Context, namespace, name string) (*capiv1beta1.ClusterClass, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &capiv1beta1.ClusterClass{}
	err = client.Get().
		Namespace(namespace).
		Resource("clusterclasses").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, ClusterAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *Kubernetes) GetClusterClasses(namespace string, selectors ...Selector) (*capiv1beta1.ClusterClassList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetClusterClassesContext(ctx, namespace, selectors...)
}

func (k *Kubernetes) GetClusterClassesContext(ctx context.Context, namespace string, selectors ...Selector) (*capiv1beta1.ClusterClassList, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &capiv1beta1.ClusterClassList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("clusterclasses").
		VersionedParams(&opts, ClusterAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *ClusterAPI) UpdateKubeadmControlPlane(obj *kubeadmv1beta1.KubeadmControlPlane) (*kubeadmv1beta1.KubeadmControlPlane, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.UpdateKubeadmControlPlaneContext(ctx, obj)
}

func (k *ClusterAPI) UpdateKubeadmControlPlaneContext(ctx context.Context, obj *kubeadmv1beta1.KubeadmControlPlane) (*kubeadmv1beta1.KubeadmControlPlane, error) {
	config := *k.Config
	config.GroupVersion = &kubeadmv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	result := &kubeadmv1beta1.KubeadmControlPlane{}
	err = client.Put().
		Namespace(obj.Namespace).
		Resource("kubeadmcontrolplanes").
		Name(obj.Name).
		VersionedParams(&metav1.UpdateOptions{}, ClusterAPIParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *ClusterAPI) GetKubeadmControlPlanes(namespace string, selectors ...Selector) (*kubeadmv1beta1.KubeadmControlPlaneList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetKubeadmControlPlanesContext(ctx, namespace, selectors...)
}

func (k *ClusterAPI) GetKubeadmControlPlanesContext(ctx context.Context, namespace string, selectors ...Selector) (*kubeadmv1beta1.KubeadmControlPlaneList, error) {
	config := *k.Config
	config.GroupVersion = &kubeadmv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &kubeadmv1beta1.KubeadmControlPlaneList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("kubeadmcontrolplanes").
		VersionedParams(&opts, ClusterAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *ClusterAPI) UpdateMachineDeployment(obj *capiv1beta1.MachineDeployment) (*capiv1beta1.MachineDeployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.UpdateMachineDeploymentContext(ctx, obj)
}

func (k *ClusterAPI) UpdateMachineDeploymentContext(ctx context.Context, obj *capiv1beta1.MachineDeployment) (*capiv1beta1.MachineDeployment, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	result := &capiv1beta1.MachineDeployment{}
	err = client.Put().
		Namespace(obj.Namespace).
		Resource("machinedeployments").
		Name(obj.Name).
		VersionedParams(&metav1.UpdateOptions{}, ClusterAPIParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *ClusterAPI) GetMachineDeployments(namespace string, selectors ...Selector) (*capiv1beta1.MachineDeploymentList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetMachineDeploymentsContext(ctx, namespace, selectors...)
}

func (k *ClusterAPI) GetMachineDeploymentsContext(ctx context.Context, namespace string, selectors ...Selector) (*capiv1beta1.MachineDeploymentList, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &capiv1beta1.MachineDeploymentList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("machinedeployments").
		VersionedParams(&opts, ClusterAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *ClusterAPI) GetMachines(namespace string, selectors ...Selector) (*capiv1beta1.MachineList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetMachinesContext(ctx, namespace, selectors...)
}

func (k *ClusterAPI) GetMachinesContext(ctx context.Context, namespace string, selectors ...Selector) (*capiv1beta1.MachineList, error) {
	config := *k.Config
	config.GroupVersion = &capiv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = ClusterAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &capiv1beta1.MachineList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("machines").
		VersionedParams(&opts, ClusterAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewClusterAPI(k *Kubernetes, clusterTemplate string) (*ClusterAPI, error) {
	t, err := textutil.EZTemplate(clusterTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing cluster template")
	}
	return &ClusterAPI{Kubernetes: k, tpl: t}, nil
}

func NewClusterAPIFromKubeConfig(p []byte, clusterTemplate string) (*ClusterAPI, error) {
	k, err := NewKubernetesFromKubeConfig(p)
	if err != nil {
		return nil, err
	}
	return NewClusterAPI(k, clusterTemplate)
}

func NewClusterAPIFromToken(endpoint string, token []byte, clientCfg rest.TLSClientConfig, clusterTemplate string) (*ClusterAPI, error) {
	k, err := NewKubernetesFromToken(endpoint, token, clientCfg)
	if err != nil {
		return nil, err
	}
	return NewClusterAPI(k, clusterTemplate)
}

type ClusterTemplateRenderingError struct {
	Err error
}

func (e *ClusterTemplateRenderingError) Error() string {
	return e.Err.Error()
}

func IsClusterTemplateRenderingError(err error) bool {
	if _, ok := err.(*ClusterTemplateRenderingError); ok {
		return true
	}
	return false
}

type ClusterTemplateConfig struct {
	Name      string
	Namespace string
	Params    map[string]string
	Dryrun    bool
}

var (
	ClusterAPIScheme         = runtime.NewScheme()
	ClusterAPICodecs         = serializer.NewCodecFactory(ClusterAPIScheme)
	ClusterAPIParameterCodec = runtime.NewParameterCodec(ClusterAPIScheme)
)

func init() {
	metav1.AddToGroupVersion(ClusterAPIScheme, metav1.Unversioned)
	metav1.AddToGroupVersion(ClusterAPIScheme, capiv1beta1.GroupVersion)
	metav1.AddToGroupVersion(ClusterAPIScheme, kubeadmv1beta1.GroupVersion)

	capiv1beta1.AddToScheme(ClusterAPIScheme)
	kubeadmv1beta1.AddToScheme(ClusterAPIScheme)
}
