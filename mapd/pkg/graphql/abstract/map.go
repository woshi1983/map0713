// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"encoding/json"
	"fmt"
	"io"

	graphql "github.com/99designs/gqlgen/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
)

func MarshalMap(msg map[string]string) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		buf, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		w.Write(buf)
	})
}

func UnmarshalMap(v interface{}) (map[string]string, error) {
	switch v := v.(type) {
	case string:
		m := make(map[string]string)
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return nil, err
		}
		return m, nil
	case []byte:
		m := make(map[string]string)
		if err := json.Unmarshal(v, &m); err != nil {
			return nil, err
		}
		return m, nil
	case map[string]string:
		return v, nil
	case map[string]interface{}:
		m := make(map[string]string)
		for key, val := range v {
			str, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("%T cannot be used as Map value", val)
			}
			m[key] = str
		}
		return m, nil
	default:
		return nil, fmt.Errorf("%T is not a Map object", v)
	}
}

func MarshalNullableMap(msg map[string]*string) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		buf, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		w.Write(buf)
	})
}

func UnmarshalNullableMap(v interface{}) (map[string]*string, error) {
	switch v := v.(type) {
	case string:
		m := make(map[string]*string)
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return nil, err
		}
		return m, nil
	case []byte:
		m := make(map[string]*string)
		if err := json.Unmarshal(v, &m); err != nil {
			return nil, err
		}
		return m, nil
	case map[string]*string:
		return v, nil
	case map[string]interface{}:
		m := make(map[string]*string)
		for key, val := range v {
			switch val := val.(type) {
			case string:
				m[key] = util.StringPtr(val)
			case *string:
				m[key] = val
			default:
				if val != nil {
					return nil, fmt.Errorf("%T cannot be used as Map value", val)
				}
				m[key] = (*string)(nil)
			}
		}
		return m, nil
	default:
		return nil, fmt.Errorf("%T is not a Map object", v)
	}
}
