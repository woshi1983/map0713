// Copyright (c) 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ContentLibraryProviderSpec defines the desired state of ContentLibraryProvider
type ContentLibraryProviderSpec struct {
	// UUID describes the UUID of a vSphere content library. It is the unique identifier for a
	// vSphere content library.
	UUID string `json:"uuid,omitempty"`
}

// ContentLibraryProviderStatus defines the observed state of ContentLibraryProvider
// Can include fields indicating when was the last time VM images were updated from a library
type ContentLibraryProviderStatus struct {
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ContentLibraryProvider is the Schema for the contentlibraryproviders API
type ContentLibraryProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContentLibraryProviderSpec   `json:"spec,omitempty"`
	Status ContentLibraryProviderStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ContentLibraryProviderList contains a list of ContentLibraryProvider
type ContentLibraryProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContentLibraryProvider `json:"items"`
}
