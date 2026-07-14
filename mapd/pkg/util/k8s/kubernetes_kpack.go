// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"

	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	scheme "github.com/pivotal/kpack/pkg/client/clientset/versioned/scheme"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type KpackKubernetes struct {
	*Kubernetes
}

func (w KpackKubernetes) CreateImage(obj *buildv1alpha1.Image) (*buildv1alpha1.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateImageContext(ctx, obj)
}

func (w KpackKubernetes) CreateImageContext(ctx context.Context, obj *buildv1alpha1.Image) (*buildv1alpha1.Image, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &buildv1alpha1.Image{}
	err = clientset.KpackV1alpha1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("images").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w KpackKubernetes) UpdateImage(obj *buildv1alpha1.Image) (*buildv1alpha1.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateImageContext(ctx, obj)
}

func (w KpackKubernetes) UpdateImageContext(ctx context.Context, obj *buildv1alpha1.Image) (*buildv1alpha1.Image, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &buildv1alpha1.Image{}
	err = clientset.KpackV1alpha1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("images").
		Name(obj.Name).
		VersionedParams(&metav1.UpdateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w KpackKubernetes) GetImage(namespace, name string) (*buildv1alpha1.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetImageContext(ctx, namespace, name)
}

func (w KpackKubernetes) GetImageContext(ctx context.Context, namespace, name string) (*buildv1alpha1.Image, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &buildv1alpha1.Image{}
	err = clientset.KpackV1alpha1().RESTClient().Get().
		Namespace(namespace).
		Resource("images").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w KpackKubernetes) GetImages(namespace string, selectors ...Selector) (*buildv1alpha1.ImageList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetImagesContext(ctx, namespace, selectors...)
}

func (w KpackKubernetes) GetImagesContext(ctx context.Context, namespace string, selectors ...Selector) (*buildv1alpha1.ImageList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &buildv1alpha1.ImageList{}
	err = clientset.KpackV1alpha1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("images").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w KpackKubernetes) DeleteImage(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteImageContext(ctx, namespace, name)
}

func (w KpackKubernetes) DeleteImageContext(ctx context.Context, namespace, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.KpackV1alpha1().RESTClient().Delete().
		Namespace(namespace).
		Resource("images").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func (w KpackKubernetes) CreateBuild(obj *buildv1alpha1.Build) (*buildv1alpha1.Build, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateBuildContext(ctx, obj)
}

func (w KpackKubernetes) CreateBuildContext(ctx context.Context, obj *buildv1alpha1.Build) (*buildv1alpha1.Build, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &buildv1alpha1.Build{}
	err = clientset.KpackV1alpha1().RESTClient().Post().
		Namespace(obj.Namespace).
		Resource("builds").
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w KpackKubernetes) UpdateBuild(obj *buildv1alpha1.Build) (*buildv1alpha1.Build, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateBuildContext(ctx, obj)
}

func (w KpackKubernetes) UpdateBuildContext(ctx context.Context, obj *buildv1alpha1.Build) (*buildv1alpha1.Build, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &buildv1alpha1.Build{}
	err = clientset.KpackV1alpha1().RESTClient().Put().
		Namespace(obj.Namespace).
		Resource("builds").
		Name(obj.Name).
		VersionedParams(&metav1.UpdateOptions{}, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w KpackKubernetes) GetBuild(namespace, name string) (*buildv1alpha1.Build, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetBuildContext(ctx, namespace, name)
}

func (w KpackKubernetes) GetBuildContext(ctx context.Context, namespace, name string) (*buildv1alpha1.Build, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &buildv1alpha1.Build{}
	err = clientset.KpackV1alpha1().RESTClient().Get().
		Namespace(namespace).
		Resource("builds").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w KpackKubernetes) GetBuilds(namespace string, selectors ...Selector) (*buildv1alpha1.BuildList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetBuildsContext(ctx, namespace, selectors...)
}

func (w KpackKubernetes) GetBuildsContext(ctx context.Context, namespace string, selectors ...Selector) (*buildv1alpha1.BuildList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &buildv1alpha1.BuildList{}
	err = clientset.KpackV1alpha1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("builds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w KpackKubernetes) DeleteBuild(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteBuildContext(ctx, namespace, name)
}

func (w KpackKubernetes) DeleteBuildContext(ctx context.Context, namespace, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground

	return clientset.KpackV1alpha1().RESTClient().Delete().
		Namespace(namespace).
		Resource("builds").
		Name(name).
		Body(&metav1.DeleteOptions{PropagationPolicy: &bg}).
		Do(ctx).
		Error()
}

func NewKpack(w *Kubernetes) *KpackKubernetes {
	return &KpackKubernetes{w}
}

func NewKpackFromToken(endpoint string, token []byte, clientCfg rest.TLSClientConfig) (*KpackKubernetes, error) {
	w, err := NewKubernetesFromToken(endpoint, token, clientCfg)
	if err != nil {
		return nil, err
	}
	return NewKpack(w), nil
}
