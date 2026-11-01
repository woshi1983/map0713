// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"strings"
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestJobFailureMessageIncludesObservedLoadBalancerState(t *testing.T) {
	job := &batchv1.Job{}
	job.Spec.ActiveDeadlineSeconds = int64Pointer(900)
	message := jobFailureMessage("Grafana", "tkgm-test02", "vmware-system-map", "cluster-extension-grafana", job, "DeadlineExceeded", []string{"tanzu-system-dashboards/grafana"})
	for _, want := range []string{"Grafana", "tkgm-test02", "vmware-system-map/cluster-extension-grafana", "15m", "tanzu-system-dashboards/grafana", "LoadBalancer", "AVI IPAM"} {
		if !strings.Contains(message, want) {
			t.Errorf("message %q missing %q", message, want)
		}
	}
}

func TestPendingLoadBalancerServiceName(t *testing.T) {
	svc := &corev1.Service{}
	svc.Spec.Type = corev1.ServiceTypeLoadBalancer
	if !isPendingLoadBalancerService(svc) {
		t.Fatal("missing ingress should be pending")
	}
	svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: "172.16.84.1"}}
	if isPendingLoadBalancerService(svc) {
		t.Fatal("assigned ingress should be ready")
	}
	svc.Spec.Type = corev1.ServiceTypeNodePort
	svc.Status.LoadBalancer.Ingress = nil
	if isPendingLoadBalancerService(svc) {
		t.Fatal("NodePort should not be pending LB")
	}
}

func int64Pointer(value int64) *int64 { return &value }

func TestUnmarshalEnvFromExtensionTreatsEmptySpecAsEmptyEnv(t *testing.T) {
	tests := []struct {
		name string
		spec []byte
	}{
		{name: "nil", spec: nil},
		{name: "empty", spec: []byte("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{}
			ext := &model.ClusterExtension{Spec: tt.spec}

			if err := unmarshalEnvFromExtension(ext, env); err != nil {
				t.Fatalf("unmarshalEnvFromExtension() error = %v", err)
			}
			if len(env) != 0 {
				t.Fatalf("unmarshalEnvFromExtension() env length = %d, want 0", len(env))
			}
		})
	}
}
