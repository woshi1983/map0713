// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package liberr

func MergeAll(ech ...<-chan error) <-chan error {
	switch len(ech) {
	case 0:
		errCh := make(chan error, 0)
		close(errCh)
		return errCh
	case 1:
		return ech[0]
	}
	errCh := Merge(ech[0], ech[1])
	for i := 2; i < len(ech); i++ {
		errCh = Merge(errCh, ech[i])
	}
	return errCh
}

func Merge(ech1, ech2 <-chan error) <-chan error {
	errCh := make(chan error, defChanSize)
	go func() {
		defer close(errCh)
		i := 0
	LOOP:
		for i < 2 {
			select {
			case e, ok := <-ech1:
				if ok {
					errCh <- e
					continue LOOP
				}
				i++
			case e, ok := <-ech2:
				if ok {
					errCh <- e
					continue LOOP
				}
				i++
			}
		}
	}()
	return errCh
}
