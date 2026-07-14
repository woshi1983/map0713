// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"fmt"
	"io"
	"net"
	"strconv"

	graphql "github.com/99designs/gqlgen/graphql"
)

func MarshalCIDR(ipNet *net.IPNet) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(ipNet.String()))
	})
}

func UnmarshalCIDR(v interface{}) (*net.IPNet, error) {
	switch v := v.(type) {
	case *net.IPNet:
		return v, nil
	case string:
		if str, err := strconv.Unquote(v); err == nil {
			v = str
		}
		_, ipNet, err := net.ParseCIDR(v)
		if err != nil {
			return nil, fmt.Errorf("%s is not a valid CIDR", v)
		}
		return ipNet, nil
	default:
		return nil, fmt.Errorf("%T is not a CIDR", v)
	}
}

func MustParseCIDR(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipNet
}
