// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	versioned "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type APIExtensions struct {
	*Kubernetes
}

func (w APIExtensions) CreateCustomResourceDefinition(obj *apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.CreateCustomResourceDefinitionContext(ctx, obj)
}

func (w APIExtensions) CreateCustomResourceDefinitionContext(ctx context.Context, obj *apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, obj, metav1.CreateOptions{})
}

func (w APIExtensions) UpdateCustomResourceDefinition(obj *apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.UpdateCustomResourceDefinitionContext(ctx, obj)
}

func (w APIExtensions) UpdateCustomResourceDefinitionContext(ctx context.Context, obj *apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, obj, metav1.UpdateOptions{})
}

func (w APIExtensions) GetCustomResourceDefinition(name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetCustomResourceDefinitionContext(ctx, name)
}

func (w APIExtensions) GetCustomResourceDefinitionContext(ctx context.Context, name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
}

func (w APIExtensions) GetCustomResourceDefinitions(selectors ...Selector) (*apiextensionsv1.CustomResourceDefinitionList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetCustomResourceDefinitionsContext(ctx, selectors...)
}

func (w APIExtensions) GetCustomResourceDefinitionsContext(ctx context.Context, selectors ...Selector) (*apiextensionsv1.CustomResourceDefinitionList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	return clientset.ApiextensionsV1().CustomResourceDefinitions().List(ctx, opts)
}

func (w APIExtensions) DeleteCustomResourceDefinition(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.DeleteCustomResourceDefinitionContext(ctx, name)
}

func (w APIExtensions) DeleteCustomResourceDefinitionContext(ctx context.Context, name string) error {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return err
	}
	bg := metav1.DeletePropagationBackground
	return clientset.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &bg})
}

func NewAPIExtensions(w *Kubernetes) *APIExtensions {
	return &APIExtensions{w}
}

func NewAPIExtensionsFromToken(endpoint string, token []byte, clientCfg rest.TLSClientConfig) (*APIExtensions, error) {
	w, err := NewKubernetesFromToken(endpoint, token, clientCfg)
	if err != nil {
		return nil, err
	}
	return NewAPIExtensions(w), nil
}
