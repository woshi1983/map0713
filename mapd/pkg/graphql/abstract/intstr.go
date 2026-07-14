// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"

	graphql "github.com/99designs/gqlgen/graphql"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func MarshalIntOrString(val intstr.IntOrString) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		b, err := json.Marshal(val)
		if err != nil {
			panic(err)
		}
		w.Write(b)
	})
}

func UnmarshalIntOrString(v interface{}) (intstr.IntOrString, error) {
	switch v := v.(type) {
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return intstr.IntOrString{}, fmt.Errorf("%T is not a IntOrString", v)
		}
		if i > math.MaxInt32 {
			i = math.MaxInt32
		}
		return intstr.FromInt(int(i)), nil
	case int:
		return intstr.FromInt(v), nil
	case int64:
		return intstr.FromInt(int(v)), nil
	case string:
		if str, err := strconv.Unquote(v); err == nil {
			v = str
		}
		return intstr.FromString(v), nil
	default:
		return intstr.IntOrString{}, fmt.Errorf("%T is not a IntOrString", v)
	}
}

func MustParseIntOrString(s string) intstr.IntOrString {
	return intstr.Parse(s)
}
