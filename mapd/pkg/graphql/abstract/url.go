// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"fmt"
	"io"
	"net/url"
	"strconv"

	graphql "github.com/99designs/gqlgen/graphql"
)

func MarshalURL(u *url.URL) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(u.String()))
	})
}

func UnmarshalURL(v interface{}) (*url.URL, error) {
	switch v := v.(type) {
	case *url.URL:
		return v, nil
	case string:
		if str, err := strconv.Unquote(v); err == nil {
			v = str
		}
		u, err := url.Parse(v)
		if err != nil {
			return u, fmt.Errorf("%s is not a valid URL", v)
		}
		return u, nil
	default:
		return nil, fmt.Errorf("%T is not a URL", v)
	}
}

func MustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
