// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"fmt"
	"io"
	"strconv"
	"time"

	graphql "github.com/99designs/gqlgen/graphql"
)

const (
	ISO8601 = "2006-01-02T15:04:05.000Z07:00"
)

func MarshalTime(t time.Time) graphql.Marshaler {
	if t.IsZero() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(t.Format(ISO8601)))
	})
}

func UnmarshalTime(v interface{}) (time.Time, error) {
	if tmpStr, ok := v.(string); ok {
		return time.Parse(ISO8601, tmpStr)
	}
	return time.Time{}, fmt.Errorf("%T is not a Time", v)
}

func MustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}
