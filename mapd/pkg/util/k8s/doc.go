// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func IsErrUnauthorized(err error) bool {
	if e, ok := err.(*manifestAPIError); ok {
		return apierrors.IsUnauthorized(e.Err)
	}
	return apierrors.IsUnauthorized(err)
}

func IsErrObjectNotExists(err error) bool {
	if e, ok := err.(*manifestAPIError); ok {
		return apierrors.IsNotFound(e.Err)
	}
	return apierrors.IsNotFound(err)
}

func IsErrObjectAlreadyExists(err error) bool {
	if e, ok := err.(*manifestAPIError); ok {
		return apierrors.IsAlreadyExists(e.Err) || apierrors.IsConflict(e.Err)
	}
	return apierrors.IsAlreadyExists(err) || apierrors.IsConflict(err)
}

func IsErrInvalid(err error) bool {
	if e, ok := err.(*manifestAPIError); ok {
		return apierrors.IsInvalid(e.Err)
	}
	return apierrors.IsInvalid(err)
}

func IsErrUnexpected(err error) bool {
	if e, ok := err.(*manifestAPIError); ok {
		if status := apierrors.APIStatus(nil); !errors.As(e.Err, &status) {
			return true
		}
		return false
	}
	if status := apierrors.APIStatus(nil); !errors.As(err, &status) {
		return true
	}
	return false
}

func IsErrInvalidManifestSyntax(err error) bool {
	if _, ok := err.(*yamlutil.YAMLSyntaxError); ok {
		return true
	}
	if _, ok := err.(yamlutil.YAMLSyntaxError); ok {
		return true
	}
	if _, ok := err.(*yamlutil.JSONSyntaxError); ok {
		return true
	}
	if _, ok := err.(yamlutil.JSONSyntaxError); ok {
		return true
	}
	return false
}

type DockerConfigJSON struct {
	Auths map[string]*DockerConfigJSONAuth `json:"auths"`
}

type DockerConfigJSONAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     []byte `json:"auth,omitempty"`
}

const FieldManager = "vmware-tanzu-map"
