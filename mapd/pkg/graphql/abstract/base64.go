// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func MarshalBase64String(val []byte) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		b, err := json.Marshal(base64.RawStdEncoding.EncodeToString(val))
		if err != nil {
			panic(err)
		}
		w.Write(b)
	})
}

func UnmarshalBase64String(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case []byte:
		dbuf := make([]byte, base64.RawStdEncoding.DecodedLen(len(v)))
		n, err := base64.RawStdEncoding.Decode(dbuf, []byte(v))
		if err != nil {
			return nil, errors.Wrap(err, "invalid base64 string")
		}
		return dbuf[:n], nil
	case string:
		p, err := base64.RawStdEncoding.DecodeString(v)
		if err != nil {
			return nil, errors.Wrap(err, "invalid base64 string")
		}
		return p, nil
	default:
		return nil, fmt.Errorf("%T is not base64 string", v)
	}
}
