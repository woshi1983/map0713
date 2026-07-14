// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"encoding/json"
	"fmt"
	"io"

	graphql "github.com/99designs/gqlgen/graphql"
)

func MarshalAny(msg json.RawMessage) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		buf, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		w.Write(buf)
	})
}

func UnmarshalAny(v interface{}) (json.RawMessage, error) {
	switch v := v.(type) {
	case int:
		return json.Marshal(v)
	case int64:
		return json.Marshal(v)
	case float32:
		return json.Marshal(v)
	case float64:
		return json.Marshal(v)
	case string:
		return json.Marshal(v)
	case json.RawMessage:
		return v, nil
	case map[string]interface{}:
		return json.Marshal(v)
	default:
		return json.RawMessage{}, fmt.Errorf("%T is not a JSON object", v)
	}
}
