// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package harbor

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ErrAPICall struct {
	Method     string
	URL        *url.URL
	StatusCode int
	Reason     string
}

func (e *ErrAPICall) Error() string {
	return fmt.Sprintf("%s %s: %d %s", e.Method, e.URL, e.StatusCode, e.Reason)
}

func IsErrUnauthorized(err error) bool {
	if e, ok := err.(*ErrAPICall); ok {
		return e.StatusCode == http.StatusUnauthorized
	}
	return false
}

func IsErrObjectNotExists(err error) bool {
	if e, ok := err.(*ErrAPICall); ok {
		return e.StatusCode == http.StatusNotFound
	}
	return false
}

func IsErrObjectAlreadyExists(err error) bool {
	if e, ok := err.(*ErrAPICall); ok {
		return e.StatusCode == http.StatusConflict
	}
	return false
}

func validateResponse(resp *http.Response) ([]byte, error) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error reading body: %v", err)}
	}
	if resp.StatusCode >= http.StatusBadRequest {
		if len(b) > 0 {
			return nil, &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: string(b)}
		}
		return nil, &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: resp.Status}
	}
	return b, nil
}
