// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package compressutil

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func writeTar(f io.Writer, from ...string) error {
	w := &TarWriter{w: tar.NewWriter(f)}
	defer w.Close()

	for _, each := range from {
		if err := w.WriteFile(each); err != nil {
			return err
		}
	}
	return nil
}

func CompressTAR(fpath string, from ...string) error {
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return err
	}

	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	return writeTar(f, from...)
}

func extractTar(fpath string, f io.Reader) error {
	r := tar.NewReader(f)

	for {
		header, err := r.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		// validate name against path traversal
		if header.Name == "" || strings.Contains(header.Name, `\`) || strings.HasPrefix(header.Name, "/") || strings.Contains(header.Name, "../") {
			return fmt.Errorf("tar contained invalid entry: %q", header.Name)
		}

		// add dst + re-format slashes according to system
		target := filepath.Join(fpath, header.Name)
		// if no join is needed, replace with ToSlash:
		// target = filepath.ToSlash(header.Name)

		// check the type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it (with 0755 permission)
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		// if it's a file create it (with same permission)
		case tar.TypeReg:
			fw, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer fw.Close() // make sure the file is closed when we leave

			if _, err := io.Copy(fw, r); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			fw.Close()
		}
	}
	return nil
}

func DecompressTAR(fpath, from string) error {
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	defer f.Close()

	return extractTar(fpath, f)
}
