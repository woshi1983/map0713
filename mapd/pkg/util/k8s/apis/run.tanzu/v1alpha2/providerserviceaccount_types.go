// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProviderServiceAccountSpec defines the desired state of ProviderServiceAccount.
type ProviderServiceAccountSpec struct {
	// Ref specifies the reference to the TanzuKubernetesCluster for which the ProviderServiceAccount needs to be realized.
	Ref *corev1.ObjectReference `json:"ref"`

	// Rules specifies the privileges that need to be granted to the service account.
	Rules []rbacv1.PolicyRule `json:"rules"`

	// TargetNamespace is the namespace in the target cluster where the secret containing the generated service account
	// token needs to be created.
	TargetNamespace string `json:"targetNamespace"`

	// TargetSecretName is the name of the secret in the target cluster that contains the generated service account
	// token.
	TargetSecretName string `json:"targetSecretName"`
}

// ProviderServiceAccountStatus defines the observed state of ProviderServiceAccount.
type ProviderServiceAccountStatus struct {
	Ready    bool   `json:"ready,omitempty"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProviderServiceAccount is the schema for the ProviderServiceAccount API.
type ProviderServiceAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProviderServiceAccountSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProviderServiceAccountList contains a list of ProviderServiceAccount
type ProviderServiceAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderServiceAccount `json:"items"`
}
