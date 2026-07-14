// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/mapd/clientset/versioned/scheme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func patchNodeLabelsAndSpec(client *k8s.Kubernetes, name string, value nodePatcher) (*corev1.Node, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	node, err := client.PatchNode(name, types.MergePatchType, b)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func cordonNodeBeforeDrain(ctx context.Context, cli kubernetes.Interface, name string) (*corev1.Node, error) {
	result := &corev1.Node{}
	err := cli.CoreV1().RESTClient().Patch(types.MergePatchType).
		Resource("nodes").
		Name(name).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body([]byte(`{"spec":{"unschedulable":true}}`)).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getNodes(client *k8s.Kubernetes, selectors ...k8s.Selector) ([]corev1.Node, error) {
	list, err := client.GetNodes(selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func getNodeMetrics(client *k8s.MetricsKubernetes, name string) (*metricsv1beta1.NodeMetrics, error) {
	metrics, err := client.GetNodeMetric(name)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func convertNode(node *model.Node, obj *corev1.Node) {
	node.Name = obj.ObjectMeta.Name
	node.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		node.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		node.CreatedAt = time.Now()
	}

	node.Spec = &model.NodeSpec{
		Unschedulable: obj.Spec.Unschedulable,
	}
	for _, each := range obj.Spec.Taints {
		node.Spec.Taints = append(node.Spec.Taints, &model.NodeTaint{
			Key:    each.Key,
			Value:  each.Value,
			Effect: string(each.Effect),
		})
	}

	node.Status = &model.NodeStatus{
		KernelVersion: obj.Status.NodeInfo.KernelVersion,
		Os:            obj.Status.NodeInfo.OperatingSystem,
		OsImage:       obj.Status.NodeInfo.OSImage,
		Allocatable:   make(map[string]string),
		Capacity:      make(map[string]string),
	}
	for k, v := range obj.Status.Allocatable {
		node.Status.Allocatable[k.String()] = v.String()

		p := make([]byte, 0, 18)
		number, suffix := v.CanonicalizeBytes(p)
		node.Status.Allocatable[fmt.Sprintf("%s_number", k)] = string(number)
		node.Status.Allocatable[fmt.Sprintf("%s_suffix", k)] = string(suffix)
	}
	for k, v := range obj.Status.Capacity {
		node.Status.Capacity[k.String()] = v.String()

		p := make([]byte, 0, 18)
		number, suffix := v.CanonicalizeBytes(p)
		node.Status.Capacity[fmt.Sprintf("%s_number", k)] = string(number)
		node.Status.Capacity[fmt.Sprintf("%s_suffix", k)] = string(suffix)
	}
	for _, cond := range obj.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			switch cond.Status {
			case corev1.ConditionTrue:
				node.Status.Ready = util.BoolPtr(true)
			case corev1.ConditionFalse:
				node.Status.Ready = util.BoolPtr(false)
			}
		}
	}
	for _, addr := range obj.Status.Addresses {
		switch addr.Type {
		case corev1.NodeInternalIP:
			ip := net.ParseIP(addr.Address)
			if ip != nil {
				node.Status.InternalIP = append(node.Status.InternalIP, ip)
			}
		case corev1.NodeExternalIP:
			ip := net.ParseIP(addr.Address)
			if ip != nil {
				node.Status.ExternalIP = append(node.Status.ExternalIP, ip)
			}
		}
	}

	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Unix(0, 0)
	}
	if node.LastModified.IsZero() {
		node.LastModified = node.CreatedAt
	}
}

func getPodsTotalRequestsAndLimits(podList []corev1.Pod) (reqs map[corev1.ResourceName]resource.Quantity, limits map[corev1.ResourceName]resource.Quantity) {
	reqs, limits = map[corev1.ResourceName]resource.Quantity{}, map[corev1.ResourceName]resource.Quantity{}
	for _, pod := range podList {
		podReqs, podLimits := resourcehelper.PodRequestsAndLimits(&pod)
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = podReqValue.DeepCopy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = podLimitValue.DeepCopy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return
}

func cordonAndDrainNode(ctx context.Context, cli kubernetes.Interface, nodeName string, timeout time.Duration, evChan chan<- *model.DrainNodeEvent) {
	defer close(evChan)

	defer func() {
		if r := recover(); r != nil {
			evChan <- &model.DrainNodeEvent{
				Stderr:    fmt.Sprintf("critical error occured while draining node %s: %v.\n", nodeName, r),
				Timestamp: time.Now(),
			}
		}
	}()

	evChan <- &model.DrainNodeEvent{
		Stdout:    fmt.Sprintf("start to cordon and drain node %s...\n", nodeName),
		Timestamp: time.Now(),
	}

	if err := util.Retry(func() error {
		_, err := cordonNodeBeforeDrain(ctx, cli, nodeName)
		return err
	}, 3, time.Millisecond*500); err != nil {
		evChan <- &model.DrainNodeEvent{
			Stderr:    fmt.Sprintf("failed to cordon node %s: %v.", nodeName, err),
			Timestamp: time.Now(),
		}
		return
	}

	evChan <- &model.DrainNodeEvent{
		Stdout:    fmt.Sprintf("node %s is successfully cordoned.\nstart to evict pods on node %s...\n", nodeName, nodeName),
		Timestamp: time.Now(),
	}

	drainer := &drain.Helper{
		Ctx:                 ctx,
		Client:              cli,
		GracePeriodSeconds:  -1,
		Timeout:             timeout,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
	}

	list, errs := drainer.GetPodsForDeletion(nodeName)
	if errs != nil {
		evChan <- &model.DrainNodeEvent{
			Stderr:    fmt.Sprintf("failed to get pods on node %s due to: %v.\n", nodeName, utilerrors.NewAggregate(errs).Error()),
			Timestamp: time.Now(),
		}
		return
	}

	if warnings := list.Warnings(); warnings != "" {
		evChan <- &model.DrainNodeEvent{
			Stderr:    fmt.Sprintf("WARNING: %s\n", warnings),
			Timestamp: time.Now(),
		}
	}

	pods := list.Pods()
	totalPods := int64(len(pods))

	evChan <- &model.DrainNodeEvent{
		TotalPods: totalPods,
		Stdout:    fmt.Sprintf("node %s is successfully cordoned.\nstart to evict pods on node %s...\n", nodeName, nodeName),
		Timestamp: time.Now(),
	}

	ed := &drainNodeLogEventDispatcher{totalPods: totalPods, eventChan: evChan}
	drainer.Out = &drainNodeLogEventWriter{patcher: ed}
	drainer.ErrOut = &drainNodeLogEventWriter{patcher: ed, isStderr: true}
	drainer.OnPodDeletedOrEvicted = func(pod *corev1.Pod, usingEviction bool) { ed.OnPodEvicted(pod) }

	if err := drainer.DeleteOrEvictPods(pods); err != nil {
		evChan <- &model.DrainNodeEvent{
			TotalPods: totalPods,
			Stderr:    fmt.Sprintf("progress is terminated due to a critical error: %s\n", err),
			Timestamp: time.Now(),
		}
		return
	}
	evChan <- &model.DrainNodeEvent{
		EvictedPods: ed.evictedPods,
		TotalPods:   ed.totalPods,
		Stdout:      fmt.Sprintf("node %s is successfully cordoned and drained.\n", nodeName),
		Timestamp:   time.Now(),
	}
}

type drainNodeLogEventDispatcher struct {
	mux         sync.RWMutex
	evictedPods int64
	totalPods   int64
	eventChan   chan<- *model.DrainNodeEvent
}

func (p *drainNodeLogEventDispatcher) OnPodEvicted(pod *corev1.Pod) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.evictedPods++
	p.eventChan <- &model.DrainNodeEvent{
		EvictedPods: p.evictedPods,
		TotalPods:   p.totalPods,
		Stdout:      fmt.Sprintf("evicted pod %s/%s.\n", pod.GetNamespace(), pod.GetName()),
		Timestamp:   time.Now(),
	}
}

func (p *drainNodeLogEventDispatcher) WriteStdout(b []byte) (n int, err error) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	p.eventChan <- &model.DrainNodeEvent{
		EvictedPods: p.evictedPods,
		TotalPods:   p.totalPods,
		Stdout:      string(b),
		Timestamp:   time.Now(),
	}
	return len(b), nil
}

func (p *drainNodeLogEventDispatcher) WriteStderr(b []byte) (n int, err error) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	p.eventChan <- &model.DrainNodeEvent{
		EvictedPods: p.evictedPods,
		TotalPods:   p.totalPods,
		Stderr:      string(b),
		Timestamp:   time.Now(),
	}
	return len(b), nil
}

type drainNodeLogEventWriter struct {
	patcher  *drainNodeLogEventDispatcher
	isStderr bool
}

func (w drainNodeLogEventWriter) Write(b []byte) (n int, err error) {
	if w.isStderr {
		return w.patcher.WriteStderr(b)
	}
	return w.patcher.WriteStdout(b)
}

type nodePatcher struct {
	Metadata nodeMetadataPatcher `json:"metadata"`
	Spec     nodeSpecPatcher     `json:"spec"`
}

type nodeMetadataPatcher struct {
	Labels map[string]*string `json:"labels"`
}

type nodeSpecPatcher struct {
	Unschedulable bool           `json:"unschedulable"`
	Taints        []corev1.Taint `json:"taints"`
}
