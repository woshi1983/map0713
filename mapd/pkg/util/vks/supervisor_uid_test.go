package vks

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSupervisorUID(t *testing.T) {
	client := &Client{kube: fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system", UID: types.UID("supervisor-123")}})}
	uid, err := client.SupervisorUID(context.Background())
	if err != nil || uid != "supervisor-123" {
		t.Fatalf("got uid %q, err %v", uid, err)
	}
	missing := &Client{kube: fake.NewSimpleClientset()}
	if _, err := missing.SupervisorUID(context.Background()); err == nil {
		t.Fatal("missing kube-system must fail")
	}
}
