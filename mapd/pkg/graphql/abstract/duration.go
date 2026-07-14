// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/peterhellberg/duration"
)

func MarshalDuration(d time.Duration) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote("PT"+strings.ToUpper(d.String())))
	})
}

func UnmarshalDuration(v interface{}) (time.Duration, error) {
	switch v := v.(type) {
	case time.Duration:
		return v, nil
	case string:
		if str, err := strconv.Unquote(v); err == nil {
			v = str
		}
		d, err := duration.Parse(v)
		if err != nil {
			return 0, fmt.Errorf("%s is not a valid ISO8601 duration", v)
		}
		return d, nil
	default:
		return 0, fmt.Errorf("%T is not a Duration", v)
	}
}
