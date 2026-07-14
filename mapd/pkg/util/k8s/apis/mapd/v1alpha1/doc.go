// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the map.tanzu v1alpha1 API group
// +k8s:deepcopy-gen=package
// +groupName=map.tanzu.vmware.com
package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var CRDBinaryImage = &apiextensionsv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: Resource("binaryimages").String(),
	},
	Spec: apiextensionsv1.CustomResourceDefinitionSpec{
		Group: SchemeGroupVersion.Group,
		Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
			Name:    "v1alpha1",
			Served:  true,
			Storage: true,
			Schema: &apiextensionsv1.CustomResourceValidation{
				OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"spec": {
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"baseImage": {
									Type: "string",
								},
								"distImage": {
									Type: "string",
								},
								"binaryDir": {
									Type: "string",
								},
								"url": {
									Type: "string",
								},
								"processUID": {
									Type: "integer",
								},
							},
						},
					},
				},
			},
		}},
		Scope: apiextensionsv1.NamespaceScoped,
		Names: apiextensionsv1.CustomResourceDefinitionNames{
			Plural:   "binaryimages",
			Singular: "binaryimage",
			Kind:     "BinaryImage",
			ListKind: "BinaryImageList",
		},
	},
}

var CRDDockerfileImage = &apiextensionsv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: Resource("dockerfileimages").String(),
	},
	Spec: apiextensionsv1.CustomResourceDefinitionSpec{
		Group: SchemeGroupVersion.Group,
		Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
			Name:    "v1alpha1",
			Served:  true,
			Storage: true,
			Schema: &apiextensionsv1.CustomResourceValidation{
				OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"spec": {
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"source": {
									Type: "string",
								},
								"distImage": {
									Type: "string",
								},
								"branch": {
									Type: "string",
								},
								"subDir": {
									Type: "string",
								},
								"fileName": {
									Type: "string",
								},
								"url": {
									Type: "string",
								},
							},
						},
					},
				},
			},
		}},
		Scope: apiextensionsv1.NamespaceScoped,
		Names: apiextensionsv1.CustomResourceDefinitionNames{
			Plural:   "dockerfileimages",
			Singular: "dockerfileimage",
			Kind:     "DockerfileImage",
			ListKind: "DockerfileImageList",
		},
	},
}
