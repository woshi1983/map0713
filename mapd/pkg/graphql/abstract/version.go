// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	semver "github.com/Masterminds/semver/v3"
)

func MarshalVersion(v *semver.Version) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(v.String()))
	})
}

func UnmarshalVersion(v interface{}) (*semver.Version, error) {
	switch v := v.(type) {
	case *semver.Version:
		return v, nil
	case string:
		if str, err := strconv.Unquote(v); err == nil {
			v = str
		}
		return semver.NewVersion(v)
	default:
		return nil, fmt.Errorf("%T is not a Version", v)
	}
}

func MustParseVersion(s string) *semver.Version {
	v, err := semver.NewVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}
