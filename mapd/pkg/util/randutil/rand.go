// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package randutil

import (
	"crypto/rand"
	"math/big"

	"github.com/pkg/errors"
)

const (
	NUMERIC           = "0123456789"
	LOWERCASE_LETTERS = "abcdefghijklmnopqrstuvwxyz"
	UPPERCASE_LETTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LOWERCASE         = NUMERIC + LOWERCASE_LETTERS
	UPPERCASE         = NUMERIC + UPPERCASE_LETTERS
	LETTERS           = LOWERCASE_LETTERS + UPPERCASE_LETTERS
	DEFAULT           = NUMERIC + LOWERCASE_LETTERS + UPPERCASE_LETTERS
)

// BytesN generates a random bytes that has length n.
func BytesN(n int, charset ...string) []byte {
	var chars string
	if len(charset) == 0 {
		chars = DEFAULT
	} else {
		for _, s := range charset {
			chars += s
		}
	}

	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(errors.Wrap(err, "error generating random bytes"))
		}
		ret[i] = chars[num.Int64()]
	}
	return ret
}

// StringN generates a random string that has length n.
func StringN(n int, charset ...string) string {
	return string(BytesN(n, charset...))
}
