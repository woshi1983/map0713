// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	mapv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/mapd/v1alpha1"
	versioned "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/mapd/clientset/versioned"
	scheme "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/mapd/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type MapKubernetes struct {
	*Kubernetes
}

func (w MapKubernetes) CreateBinaryImage(obj *mapv1alpha1.BinaryImage) (*mapv1alpha1.BinaryImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateBinaryImageContext(ctx, obj)
}

func (w MapKubernetes) CreateBinaryImageContext(ctx context.Context, obj *mapv1alpha1.BinaryImage) (*mapv1alpha1.BinaryImage, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.MapV1alpha1().BinaryImages(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
}

func (w MapKubernetes) UpdateBinaryImage(obj *mapv1alpha1.BinaryImage) (*mapv1alpha1.BinaryImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateBinaryImageContext(ctx, obj)
}

func (w MapKubernetes) UpdateBinaryImageContext(ctx context.Context, obj *mapv1alpha1.BinaryImage) (*mapv1alpha1.BinaryImage, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.MapV1alpha1().BinaryImages(obj.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
}

func (w MapKubernetes) GetBinaryImage(namespace, name string) (*mapv1alpha1.BinaryImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetBinaryImageContext(ctx, namespace, name)
}

func (w MapKubernetes) GetBinaryImageContext(ctx context.Context, namespace, name string) (*mapv1alpha1.BinaryImage, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.MapV1alpha1().BinaryImages(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (w MapKubernetes) GetBinaryImages(namespace string, selectors ...Selector) (*mapv1alpha1.BinaryImageList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetBinaryImagesContext(ctx, namespace, selectors...)
}

func (w MapKubernetes) GetBinaryImagesContext(ctx context.Context, namespace string, selectors ...Selector) (*mapv1alpha1.BinaryImageList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &mapv1alpha1.BinaryImageList{}
	err = clientset.MapV1alpha1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("binaryimages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w MapKubernetes) DeleteBinaryImage(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteBinaryImageContext(ctx, namespace, name)
}

func (w MapKubernetes) DeleteBinaryImageContext(ctx context.Context, namespace, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground
	return clientset.MapV1alpha1().BinaryImages(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &bg})
}

func (w MapKubernetes) CreateDockerfileImage(obj *mapv1alpha1.DockerfileImage) (*mapv1alpha1.DockerfileImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateDockerfileImageContext(ctx, obj)
}

func (w MapKubernetes) CreateDockerfileImageContext(ctx context.Context, obj *mapv1alpha1.DockerfileImage) (*mapv1alpha1.DockerfileImage, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.MapV1alpha1().DockerfileImages(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
}

func (w MapKubernetes) UpdateDockerfileImage(obj *mapv1alpha1.DockerfileImage) (*mapv1alpha1.DockerfileImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateDockerfileImageContext(ctx, obj)
}

func (w MapKubernetes) UpdateDockerfileImageContext(ctx context.Context, obj *mapv1alpha1.DockerfileImage) (*mapv1alpha1.DockerfileImage, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.MapV1alpha1().DockerfileImages(obj.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
}

func (w MapKubernetes) GetDockerfileImage(namespace, name string) (*mapv1alpha1.DockerfileImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetDockerfileImageContext(ctx, namespace, name)
}

func (w MapKubernetes) GetDockerfileImageContext(ctx context.Context, namespace, name string) (*mapv1alpha1.DockerfileImage, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.MapV1alpha1().DockerfileImages(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (w MapKubernetes) GetDockerfileImages(namespace string, selectors ...Selector) (*mapv1alpha1.DockerfileImageList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetDockerfileImagesContext(ctx, namespace, selectors...)
}

func (w MapKubernetes) GetDockerfileImagesContext(ctx context.Context, namespace string, selectors ...Selector) (*mapv1alpha1.DockerfileImageList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &mapv1alpha1.DockerfileImageList{}
	err = clientset.MapV1alpha1().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("dockerfileimages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w MapKubernetes) DeleteDockerfileImage(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteDockerfileImageContext(ctx, namespace, name)
}

func (w MapKubernetes) DeleteDockerfileImageContext(ctx context.Context, namespace, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground
	return clientset.MapV1alpha1().DockerfileImages(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &bg})
}

func NewMap(w *Kubernetes) *MapKubernetes {
	return &MapKubernetes{w}
}

func NewMapFromToken(endpoint string, token []byte, clientCfg rest.TLSClientConfig) (*MapKubernetes, error) {
	w, err := NewKubernetesFromToken(endpoint, token, clientCfg)
	if err != nil {
		return nil, err
	}
	return NewMap(w), nil
}
