// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/chartmuseum/helm-push/pkg/chartmuseum"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"go.uber.org/multierr"
)

// UploadChartCmd represents the upload command.
var UploadChartCmd = &cobra.Command{
	Use:   "upload [url]",
	Short: "Upload Helm charts to ChartMuseum.",
	Long:  `Upload Helm charts to ChartMuseum.`,
	Args:  cobra.ExactArgs(1),
	Example: `
$ mapkit chart upload -i my_save_dir/ \
    --username myusername --password mypassword \
    --ca-file myrootca.pem \
    https://myharbor.example.org/chartrepo/myrepo	
`,
	Run: func(cmd *cobra.Command, args []string) {
		uploadChart.URL = args[0]
		if err := uploadChart.Upload(uploadChartPath); err != nil {
			exitCode := 2
			fmt.Fprintf(os.Stderr, "FATAL(%d): %v.\n", exitCode, err)
			os.Exit(exitCode)
		}
	},
}

var (
	uploadChart     = newChartUploader()
	uploadChartPath string
)

type chartUploader struct {
	URL                   string
	Username              string
	Password              string
	CertFile              string
	KeyFile               string
	CAFile                string
	InsecureSkipTLSverify bool
	Threads               uint
}

func (d *chartUploader) Upload(from string) error {
	doneWg := new(sync.WaitGroup)
	p := mpb.New(mpb.WithWaitGroup(doneWg))

	files, err := os.ReadDir(from)
	if err != nil {
		return errors.Wrap(err, "error loading charts bundle")
	}

	pendingUpload := make(chan string, len(files))
	for _, each := range files {
		pendingUpload <- filepath.Join(from, each.Name())
	}
	close(pendingUpload)
	if d.Threads == 0 {
		d.Threads = 16
	}
	resultChan := make(chan *uploadResult, d.Threads)

	taskName := "Uploading"
	bar := p.AddBar(int64(len(files)),
		mpb.PrependDecorators(
			decor.Name(taskName, decor.WC{W: len(taskName) + 1, C: decor.DidentRight}),
			decor.CountersNoUnit("%d / %d files", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Percentage(decor.WC{W: 5, C: decor.DSyncSpaceR}), ""),
			decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth), "[DONE]"),
		),
		mpb.BarFillerClearOnComplete(),
	)

	options := []chartmuseum.Option{chartmuseum.URL(d.URL), chartmuseum.Username(d.Username), chartmuseum.Password(d.Password)}
	if d.CertFile != "" {
		options = append(options, chartmuseum.CertFile(d.CertFile))
	}
	if d.KeyFile != "" {
		options = append(options, chartmuseum.KeyFile(d.KeyFile))
	}
	if d.CAFile != "" {
		options = append(options, chartmuseum.CAFile(d.CAFile))
	}
	if d.InsecureSkipTLSverify {
		options = append(options, chartmuseum.InsecureSkipVerify(true))
	}

	client, err := chartmuseum.NewClient(options...)
	if err != nil {
		return err
	}
	var mux sync.Mutex
	for i := 0; i < int(d.Threads); i++ {
		go upload(bar, &mux, client, pendingUpload, resultChan)
	}

	for i := 0; i < len(files); i++ {
		result := <-resultChan
		if result == nil {
			continue
		}
		if result.Error != nil {
			e := errors.Wrap(result.Error, fmt.Sprintf("error uploading %s", strconv.Quote(filepath.Base(result.Name))))
			err = multierr.Append(err, e)
		}
	}
	p.Wait()
	return err
}

func (d *chartUploader) SetFlags(f *pflag.FlagSet) {
	f.StringVar(&d.Username, "username", d.Username, "Chart repository username")
	f.StringVar(&d.Password, "password", d.Password, "Chart repository password")
	f.StringVar(&d.CertFile, "cert-file", d.CertFile, "Identify HTTPS client using this SSL certificate file")
	f.StringVar(&d.KeyFile, "key-file", d.KeyFile, "Identify HTTPS client using this SSL key file")
	f.StringVar(&d.CAFile, "ca-file", d.CAFile, "Verify certificates of HTTPS-enabled servers using this CA bundle")
	f.BoolVar(&d.InsecureSkipTLSverify, "insecure-skip-tls-verify", d.InsecureSkipTLSverify, "Skip TLS certificate checks for the repository")
	f.UintVar(&d.Threads, "threads", d.Threads, "Threads to use for uploading contents")
}

func newChartUploader() *chartUploader {
	return &chartUploader{
		Threads: 16,
	}
}

func upload(p *mpb.Bar, mux *sync.Mutex, client *chartmuseum.Client, chartChan <-chan string, resultChan chan<- *uploadResult) {
	for fp := range chartChan {
		since := time.Now()
		if fp == "" {
			continue
		}

		err := uploadChartFile(client, fp)
		resultChan <- &uploadResult{
			Name:  fp,
			Error: err,
		}
		mux.Lock()
		p.Increment()
		safeDecoratorEwmaUpdate(p, time.Since(since))
		mux.Unlock()
	}
}

func uploadChartFile(client *chartmuseum.Client, fp string) error {
	resp, err := client.UploadChartPackage(fp, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, e := io.ReadAll(resp.Body)
		err = fmt.Errorf("error uploading file %s: %s: %s", fp, resp.Status, b)
		err = multierr.Append(err, e)
	}
	return err
}

type uploadResult struct {
	Name  string
	Error error
}
