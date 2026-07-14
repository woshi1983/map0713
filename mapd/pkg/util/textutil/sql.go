// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package textutil

// EscapeSQLString escapes '%' operators in a SQL expression for fuzzy query.
func EscapeSQLString(expr string) string {
	var n int
	for i := range expr {
		switch expr[i] {
		case '%', '_':
			n++
		}
	}
	if n == 0 {
		return expr // avoid allocation
	}
	buf := make([]byte, len(expr)+n)
	var j int
	for i := range expr {
		switch expr[i] {
		case '%', '_':
			j += copy(buf[j:], "\\")
			j += copy(buf[j:], expr[i:i+1])
		default:
			j += copy(buf[j:], expr[i:i+1])
		}
	}
	return string(buf)
}
