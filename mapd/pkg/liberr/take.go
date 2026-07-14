// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package liberr

func Take(n int, ech <-chan error) <-chan error {
	var (
		errCh   chan error
		satisfy func() bool
	)
	if n < 0 {
		errCh, satisfy = make(chan error, defChanSize), func() bool { return false }
	} else {
		i := 0
		errCh, satisfy = make(chan error, n), func() bool {
			i++
			if i < n {
				return false
			}
			return true
		}
	}
	go func() {
		defer close(errCh)
		for {
			e, ok := <-ech
			if !ok {
				break
			}
			errCh <- e
			if satisfy() {
				break
			}
		}
	}()
	return errCh
}
