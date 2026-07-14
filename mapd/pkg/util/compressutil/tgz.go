// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package compressutil

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type TarWriter struct {
	w *tar.Writer
}

func (t *TarWriter) WriteFile(fpath string) error {
	return filepath.WalkDir(fpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		name, err := filepath.Rel(fpath, path)
		if err != nil {
			return err
		}
		return t.WriteRegFile(name, path, fi)
	})
}

func (t *TarWriter) WriteRegFile(name, fpath string, fi os.FileInfo) error {
	// generate tar header
	header, err := tar.FileInfoHeader(fi, fpath)
	if err != nil {
		return err
	}

	// must provide real name
	// (see https://golang.org/src/archive/tar/common.go?#L626)
	header.Name = filepath.ToSlash(name)

	// write header
	if err := t.w.WriteHeader(header); err != nil {
		return err
	}

	// if not a dir, write file content
	if !fi.IsDir() {
		f, err := os.Open(fpath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(t.w, f); err != nil {
			return err
		}
	}
	return nil
}

func (t *TarWriter) Close() error {
	return t.w.Close()
}

func CompressTGZ(fpath string, from ...string) error {
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return err
	}

	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := gzip.NewWriter(f)
	defer zw.Close()

	return writeTar(zw, from...)
}

func DecompressTGZ(fpath, from string) error {
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	defer f.Close()

	zr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer zr.Close()

	return extractTar(fpath, zr)
}
