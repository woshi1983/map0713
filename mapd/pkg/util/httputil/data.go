// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package httputil

import (
	"bytes"
	"encoding/base64"
)

// DataURL indicates the data scheme URI.
//
// Layout: data:[mediatype;][base64],<data>
//
// See also:
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URIs
//   - https://docs.microsoft.com/en-us/previous-versions//cc848897(v=vs.85)?redirectedfrom=MSDN
type DataURL struct {
	MediaType string
	Base64    bool
	RawData   []byte
}

// Bytes returns the DataURL as bytes.
func (d DataURL) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString("data:")
	if d.MediaType != "" {
		buf.WriteString(d.MediaType)
		if d.MediaType[len(d.MediaType)-1] != ';' && d.Base64 {
			buf.WriteRune(';')
		}
	}
	if d.Base64 {
		buf.WriteString("base64,")
		buf.WriteString(base64.StdEncoding.EncodeToString(d.RawData))
	} else {
		buf.WriteRune(',')
		buf.Write(d.RawData)
	}
	return buf.Bytes()
}

func (d DataURL) String() string {
	return string(d.Bytes())
}
