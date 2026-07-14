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

func MarshalIP(ip net.IP) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(ip.String()))
	})
}

func UnmarshalIP(v interface{}) (net.IP, error) {
	switch v := v.(type) {
	case net.IP:
		return v, nil
	case string:
		if str, err := strconv.Unquote(v); err == nil {
			v = str
		}
		ip := net.ParseIP(v)
		if ip == nil {
			return net.IP{}, fmt.Errorf("%s is not a valid IP address", v)
		}
		return ip, nil
	default:
		return net.IP{}, fmt.Errorf("%T is not a IP address", v)
	}
}
