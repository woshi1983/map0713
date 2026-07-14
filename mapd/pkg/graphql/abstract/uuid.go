// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package abstract

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

func MarshalUUID(id uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(id.String()))
	})
}

func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	switch v := v.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		if str, err := strconv.Unquote(v); err == nil {
			v = str
		}
		return uuid.Parse(v)
	default:
		return uuid.UUID{}, fmt.Errorf("%T is not a UUID", v)
	}
}
