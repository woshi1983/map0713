// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterBootstrap is the Schema for the ClusterBootstraps API
type ClusterBootstrap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *ClusterBootstrapTemplateSpec `json:"spec"`
	Status ClusterBootstrapStatus        `json:"status,omitempty"`
}

// ClusterBootstrapStatus defines the observed state of ClusterBootstrap
type ClusterBootstrapStatus struct {
	ResolvedTKR string `json:"resolvedTKR,omitempty"`

	Conditions clusterapiv1beta1.Conditions `json:"conditions,omitempty"`
}

// GetConditions returns the set of conditions for this object. implements Setter interface
func (c *ClusterBootstrap) GetConditions() clusterapiv1beta1.Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object. implements Setter interface
func (c *ClusterBootstrap) SetConditions(conditions clusterapiv1beta1.Conditions) {
	c.Status.Conditions = conditions
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterBootstrapList contains a list of ClusterBootstrap
type ClusterBootstrapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBootstrap `json:"items"`
}
