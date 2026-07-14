// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BinaryImage is the schema for the binary build image.
type BinaryImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BinaryImageSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BinaryImageList contains a list of BinaryImage
type BinaryImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BinaryImage `json:"items"`
}

type BinaryImageSpec struct {
	BaseImage  string `json:"baseImage,omitempty"`
	DistImage  string `json:"distImage,omitempty"`
	BinaryDir  string `json:"binaryDir,omitempty"`
	URL        string `json:"url,omitempty"`
	ProcessUID int64  `json:"processUID,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Dockerfile is the schema for the dockerfile build image.
type DockerfileImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DockerfileImageSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DockerfileImageList contains a list of DockerfileImage
type DockerfileImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerfileImage `json:"items"`
}

type DockerfileImageSpec struct {
	Source    string `json:"source,omitempty"`
	DistImage string `json:"distImage,omitempty"`
	Branch    string `json:"branch,omitempty"`
	SubDir    string `json:"subDir,omitempty"`
	FileName  string `json:"fileName,omitempty"`
	URL       string `json:"url,omitempty"`
}
