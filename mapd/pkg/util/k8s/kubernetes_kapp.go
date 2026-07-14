// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"

	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

type KappController struct {
	*Kubernetes
}

func (w *KappController) CreatePackageRepository(obj *packagingv1alpha1.PackageRepository) (*packagingv1alpha1.PackageRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreatePackageRepositoryContext(ctx, obj)
}

func (w *KappController) CreatePackageRepositoryContext(ctx context.Context, obj *packagingv1alpha1.PackageRepository) (*packagingv1alpha1.PackageRepository, error) {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &packagingv1alpha1.PackageRepository{}
	err = client.Post().
		Namespace(obj.Namespace).
		Resource("packagerepositories").
		VersionedParams(&metav1.CreateOptions{}, KappParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) UpdatePackageRepository(obj *packagingv1alpha1.PackageRepository) (*packagingv1alpha1.PackageRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdatePackageRepositoryContext(ctx, obj)
}

func (w *KappController) UpdatePackageRepositoryContext(ctx context.Context, obj *packagingv1alpha1.PackageRepository) (*packagingv1alpha1.PackageRepository, error) {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &packagingv1alpha1.PackageRepository{}
	err = client.Put().
		Namespace(obj.Namespace).
		Resource("packagerepositories").
		Name(obj.Name).
		VersionedParams(&metav1.UpdateOptions{}, KappParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) PatchPackageRepository(namespace, name string, pt types.PatchType, patch []byte) (*packagingv1alpha1.PackageRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.PatchPackageRepositoryContext(ctx, namespace, name, pt, patch)
}

func (w *KappController) PatchPackageRepositoryContext(ctx context.Context, namespace, name string, pt types.PatchType, patch []byte) (*packagingv1alpha1.PackageRepository, error) {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	result := &packagingv1alpha1.PackageRepository{}
	err = client.Patch(pt).
		Namespace(namespace).
		Resource("packagerepositories").
		Name(name).
		VersionedParams(&metav1.PatchOptions{}, KappParameterCodec).
		Body(patch).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) DeletePackageRepository(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeletePackageRepositoryContext(ctx, namespace, name)
}

func (w *KappController) DeletePackageRepositoryContext(ctx context.Context, namespace, name string) error {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	err = client.Delete().
		Namespace(namespace).
		Resource("packagerepositories").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
	if err != nil {
		return err
	}
	return nil
}

func (w *KappController) GetPackageRepository(namespace, name string) (*packagingv1alpha1.PackageRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPackageRepositoryContext(ctx, namespace, name)
}

func (w *KappController) GetPackageRepositoryContext(ctx context.Context, namespace, name string) (*packagingv1alpha1.PackageRepository, error) {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &packagingv1alpha1.PackageRepository{}
	err = client.Get().
		Namespace(namespace).
		Resource("packagerepositories").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, KappParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) GetPackageRepositories(namespace string, selectors ...Selector) (*packagingv1alpha1.PackageRepositoryList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPackageRepositoriesContext(ctx, namespace, selectors...)
}

func (w *KappController) GetPackageRepositoriesContext(ctx context.Context, namespace string, selectors ...Selector) (*packagingv1alpha1.PackageRepositoryList, error) {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &packagingv1alpha1.PackageRepositoryList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("packagerepositories").
		VersionedParams(&opts, KappParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) CreatePackageInstall(obj *packagingv1alpha1.PackageInstall) (*packagingv1alpha1.PackageInstall, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreatePackageInstallContext(ctx, obj)
}

func (w *KappController) CreatePackageInstallContext(ctx context.Context, obj *packagingv1alpha1.PackageInstall) (*packagingv1alpha1.PackageInstall, error) {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &packagingv1alpha1.PackageInstall{}
	err = client.Post().
		Namespace(obj.Namespace).
		Resource("packageinstalls").
		VersionedParams(&metav1.CreateOptions{}, KappParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) GetPackageInstall(namespace, name string) (*packagingv1alpha1.PackageInstall, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPackageInstallContext(ctx, namespace, name)
}

func (w *KappController) GetPackageInstallContext(ctx context.Context, namespace, name string) (*packagingv1alpha1.PackageInstall, error) {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &packagingv1alpha1.PackageInstall{}
	err = client.Get().
		Namespace(namespace).
		Resource("packageinstalls").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, KappParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) GetPackageInstalls(namespace string, selectors ...Selector) (*packagingv1alpha1.PackageInstallList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPackageInstallsContext(ctx, namespace, selectors...)
}

func (w *KappController) GetPackageInstallsContext(ctx context.Context, namespace string, selectors ...Selector) (*packagingv1alpha1.PackageInstallList, error) {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &packagingv1alpha1.PackageInstallList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("packageinstalls").
		VersionedParams(&opts, KappParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) DeletePackageInstall(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeletePackageInstallContext(ctx, namespace, name)
}

func (w *KappController) DeletePackageInstallContext(ctx context.Context, namespace, name string) error {
	config := *w.Config
	config.GroupVersion = &packagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	err = client.Delete().
		Namespace(namespace).
		Resource("packageinstalls").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
	if err != nil {
		return err
	}
	return nil
}

func (w *KappController) GetPackage(namespace, name string) (*datapackagingv1alpha1.Package, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPackageContext(ctx, namespace, name)
}

func (w *KappController) GetPackageContext(ctx context.Context, namespace, name string) (*datapackagingv1alpha1.Package, error) {
	config := *w.Config
	config.GroupVersion = &datapackagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &datapackagingv1alpha1.Package{}
	err = client.Get().
		Namespace(namespace).
		Resource("packages").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, KappParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *KappController) GetPackages(namespace string, selectors ...Selector) (*datapackagingv1alpha1.PackageList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetPackagesContext(ctx, namespace, selectors...)
}

func (w *KappController) GetPackagesContext(ctx context.Context, namespace string, selectors ...Selector) (*datapackagingv1alpha1.PackageList, error) {
	config := *w.Config
	config.GroupVersion = &datapackagingv1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = KappCodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &datapackagingv1alpha1.PackageList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("packages").
		VersionedParams(&opts, KappParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewKappController(k *Kubernetes) *KappController {
	return &KappController{k}
}

func NewKappControllerFromKubeConfig(p []byte) (*KappController, error) {
	k, err := NewKubernetesFromKubeConfig(p)
	if err != nil {
		return nil, err
	}
	return NewKappController(k), nil
}

func NewKappControllerFromToken(endpoint string, token []byte, clientCfg rest.TLSClientConfig) (*KappController, error) {
	return &KappController{Kubernetes: &Kubernetes{Config: &rest.Config{
		Host:            endpoint,
		BearerToken:     string(token),
		TLSClientConfig: clientCfg,
	}}}, nil
}

var (
	KappScheme         = runtime.NewScheme()
	KappCodecs         = serializer.NewCodecFactory(KappScheme)
	KappParameterCodec = runtime.NewParameterCodec(KappScheme)
)

func init() {
	metav1.AddToGroupVersion(KappScheme, metav1.Unversioned)
	metav1.AddToGroupVersion(KappScheme, packagingv1alpha1.SchemeGroupVersion)
	metav1.AddToGroupVersion(KappScheme, datapackagingv1alpha1.SchemeGroupVersion)

	packagingv1alpha1.AddToScheme(KappScheme)
	datapackagingv1alpha1.AddToScheme(KappScheme)
}
