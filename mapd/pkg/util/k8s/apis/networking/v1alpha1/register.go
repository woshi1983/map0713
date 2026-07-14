// Copyright 2023 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the network v1alpha1 API group
// +k8s:deepcopy-gen=package
// +groupName=networking.tkg.tanzu.vmware.com
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Define your schema name and the version
var SchemeGroupVersion = schema.GroupVersion{
	Group:   "networking.tkg.tanzu.vmware.com",
	Version: "v1alpha1",
}

var (
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Adds the list of known types to the given scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&AKODeploymentConfig{}, &AKODeploymentConfigList{},
	)

	metav1.AddToGroupVersion(
		scheme,
		SchemeGroupVersion,
	)
	return nil
}
