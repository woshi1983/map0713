// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type MetricsKubernetes struct {
	*Kubernetes
}

func (w MetricsKubernetes) GetNodeMetric(name string) (*metricsv1beta1.NodeMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetNodeMetricContext(ctx, name)
}

func (w MetricsKubernetes) GetNodeMetricContext(ctx context.Context, name string) (*metricsv1beta1.NodeMetrics, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	return clientset.MetricsV1beta1().NodeMetricses().Get(ctx, name, metav1.GetOptions{})
}

func (w MetricsKubernetes) GetNodeMetrics(selectors ...Selector) (*metricsv1beta1.NodeMetricsList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetNodeMetricsContext(ctx, selectors...)
}

func (w MetricsKubernetes) GetNodeMetricsContext(ctx context.Context, selectors ...Selector) (*metricsv1beta1.NodeMetricsList, error) {
	clientset, err := versioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	return clientset.MetricsV1beta1().NodeMetricses().List(ctx, opts)
}

func NewMetrics(w *Kubernetes) *MetricsKubernetes {
	return &MetricsKubernetes{w}
}

func NewMetricsFromToken(endpoint string, token []byte, clientCfg rest.TLSClientConfig) (*MetricsKubernetes, error) {
	w, err := NewKubernetesFromToken(endpoint, token, clientCfg)
	if err != nil {
		return nil, err
	}
	return NewMetrics(w), nil
}
