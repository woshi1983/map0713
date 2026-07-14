// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"encoding/json"
	"math"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	runv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/run.tanzu/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateTanzuKubernetesCluster(cluster *model.Cluster, namespace string) *runv1alpha1.TanzuKubernetesCluster {
	spec := &model.ClusterTKGSSpec{}
	if err := json.Unmarshal(cluster.Spec, spec); err != nil {
		panic(err)
	}

	sm := map[string]string{}
	sm[spec.CtlPlaneStorage] = spec.CtlPlaneStorage
	sm[spec.DataPlaneStorage] = spec.DataPlaneStorage
	sm[spec.DefStorage] = spec.DefStorage
	ss := []string{}
	for s := range sm {
		if s == "" {
			continue
		}
		ss = append(ss, s)
	}

	network := &runv1alpha1.Network{
		CNI: &runv1alpha1.CNIConfiguration{
			Name: strings.ToLower(spec.CNI),
		},
		Pods: &runv1alpha1.NetworkRanges{
			CIDRBlocks: []string{spec.PodCIDR},
		},
		ServiceDomain: spec.ServiceDomain,
		Services: &runv1alpha1.NetworkRanges{
			CIDRBlocks: []string{spec.ServiceCIDR},
		},
	}
	if network.CNI.Name == "antrea-nsx-routed" {
		network.Pods.CIDRBlocks = network.Pods.CIDRBlocks[:0]
	}

	return &runv1alpha1.TanzuKubernetesCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TanzuKubernetesCluster",
			APIVersion: runv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: namespace,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
		},
		Spec: runv1alpha1.TanzuKubernetesClusterSpec{
			Distribution: runv1alpha1.Distribution{
				VersionHint: "v" + strings.TrimPrefix(cluster.Version, "v"),
			},
			Settings: &runv1alpha1.Settings{
				Network: network,
				Storage: &runv1alpha1.Storage{
					Classes:      ss,
					DefaultClass: spec.DefStorage,
				},
			},
			Topology: runv1alpha1.Topology{
				ControlPlane: runv1alpha1.TopologySettings{
					Class:        spec.CtlPlaneClass,
					Count:        int32(spec.CtlPlaneCount),
					StorageClass: spec.CtlPlaneStorage,
					Volumes: []runv1alpha1.Volume{
						{
							Name:      "etcd",
							MountPath: "/var/lib/etcd",
							Capacity: corev1.ResourceList{
								"storage": *resource.NewQuantity(spec.CtlPlaneStorageCapGi*int64(math.Pow(2, 30)), resource.BinarySI),
							},
						},
					},
				},
				Workers: runv1alpha1.TopologySettings{
					Class:        spec.DataPlaneClass,
					Count:        int32(spec.DataPlaneCount),
					StorageClass: spec.DataPlaneStorage,
					Volumes: []runv1alpha1.Volume{
						{
							Name:      "containerd",
							MountPath: "/var/lib/containerd",
							Capacity: corev1.ResourceList{
								"storage": *resource.NewQuantity(spec.DataPlaneStorageCapGi*int64(math.Pow(2, 30)), resource.BinarySI),
							},
						},
					},
				},
			},
		},
	}
}

func createNamespace(client *k8s.Kubernetes, name string) (*corev1.Namespace, error) {
	obj := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
		},
	}
	return client.CreateNamespace(obj)
}

func createSSHPrivateKeySecret(client *k8s.Kubernetes, namespace, name string, privateKey []byte) (*corev1.Secret, error) {
	obj := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
		},
		Type: corev1.SecretTypeSSHAuth,
		Data: map[string][]byte{"ssh-privatekey": privateKey},
	}
	return client.CreateSecret(obj)
}

func createServiceAccount(client *k8s.Kubernetes, namespace, name string) (*corev1.ServiceAccount, error) {
	obj := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
		},
	}
	return client.CreateServiceAccount(obj)
}

func createServiceAccountToken(client *k8s.Kubernetes, namespace, name, saName string) (*corev1.Secret, error) {
	obj := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": saName,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
	return client.CreateSecret(obj)
}

func createClusterRoleBindingClusterAdmin(client *k8s.Kubernetes, name string, saNamespace, saName string) (*rbacv1.ClusterRoleBinding, error) {
	obj := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      saName,
			Namespace: saNamespace,
		}},
	}
	return client.CreateClusterRoleBinding(obj)
}

func createClusterRoleBindingPSPPrivileged(client *k8s.Kubernetes, name string) (*rbacv1.ClusterRoleBinding, error) {
	obj := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				defLabelInternal: "true",
				defLabelManaged:  "true",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     "psp:vmware-system-privileged",
		},
		Subjects: []rbacv1.Subject{{
			Kind:     "Group",
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Name:     "system:authenticated",
		}},
	}
	return client.CreateClusterRoleBinding(obj)
}
