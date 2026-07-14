// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/cmd/helm/search"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
)

// DownloadChartCmd represents the download command.
var DownloadChartCmd = &cobra.Command{
	Use:   "download [url] [keyword]...",
	Short: "Download Helm charts from a repository.",
	Long: `Download Helm charts from a repository.
	
By default, it download all charts from VMware Bitnami repository.`,
	Example: `
$ mapkit chart download

$ mapkit chart download -o my_save_dir

$ mapkit chart download https://charts.bitnami.com/bitnami nginx`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			args = append(args, "https://charts.bitnami.com/bitnami")
		}
		downloadChart.URL = args[0]
		if err := downloadChart.Download(downloadChartPath, args[1:]...); err != nil {
			exitCode := 2
			fmt.Fprintf(os.Stderr, "FATAL(%d): %v.\n", exitCode, err)
			os.Exit(exitCode)
		}
	},
}

var (
	downloadChart     = newChartDownloader()
	downloadChartPath string
)

type chartDownloader struct {
	URL                   string
	Username              string
	Password              string
	CertFile              string
	KeyFile               string
	CAFile                string
	InsecureSkipTLSverify bool
	Fuzzy                 bool
	Regexp                bool
	AllVersion            bool
	OnlyVersion           string
	Threads               uint
	Persistent            bool
}

func (d *chartDownloader) repo() (*repo.ChartRepository, *repo.Entry, error) {
	entry := &repo.Entry{
		Name:                  "source",
		URL:                   d.URL,
		Username:              d.Username,
		Password:              d.Password,
		CertFile:              d.CertFile,
		KeyFile:               d.KeyFile,
		CAFile:                d.CAFile,
		InsecureSkipTLSverify: d.InsecureSkipTLSverify,
	}
	repo, err := repo.NewChartRepository(entry, getter.Providers{
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
		getter.Provider{
			Schemes: []string{"oci"},
			New:     getter.NewOCIGetter,
		},
	})
	return repo, entry, err
}

func (d *chartDownloader) Download(to string, keywords ...string) error {
	tempDir, err := os.MkdirTemp("", "mapkit-download-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	r, c, err := d.repo()
	if err != nil {
		return err
	}
	r.CachePath = filepath.Join(tempDir, ".cache")

	doneWg := new(sync.WaitGroup)
	p := mpb.New(mpb.WithWaitGroup(doneWg))

	chanSearchChart := make(chan struct{})
	go func() {
		spinner := p.New(-1,
			mpb.SpinnerStyle("|", "/", "-", "\\").PositionLeft(),
			mpb.PrependDecorators(
				decor.Name("Fetching metadata..."),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Spinner([]string{""}), "[DONE]"),
			),
			mpb.BarFillerClearOnComplete(),
		)
		defer spinner.SetTotal(-1, true)

		ticker := time.NewTicker(time.Microsecond * 250)
		defer ticker.Stop()

	LOOP:
		for {
			select {
			case <-chanSearchChart:
				break LOOP
			case <-ticker.C:
				spinner.Increment()
			}
		}
	}()

	var f repo.File
	if _, err := r.DownloadIndexFile(); err != nil {
		return errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", d.URL)
	}
	f.Update(c)

	repoFile := filepath.Join(tempDir, "repository.yaml")
	if err := f.WriteFile(repoFile, 0644); err != nil {
		return err
	}

	index := search.NewIndex()
	ind, err := repo.LoadIndexFile(filepath.Join(r.CachePath, helmpath.CacheIndexFile(r.Config.Name)))
	if err != nil {
		return err
	}
	index.AddRepo(r.Config.Name, ind, true)

	var res []*search.Result
	if len(keywords) == 0 {
		res = index.All()
	} else {
		term, isRegex := strings.Join(keywords, " "), d.Regexp
		if !d.Fuzzy {
			term = fmt.Sprintf("(/%s\v){1}", regexp.QuoteMeta(term))
			isRegex = true
		}
		res, err = index.Search(term, 25, isRegex)
		if err != nil {
			return err
		}
	}
	search.SortScore(res)
	res, err = d.applyConstraint(res)
	if err != nil {
		return err
	}
	sort.SliceStable(res, func(i, j int) bool {
		v1, err := semver.NewVersion(res[i].Chart.Version)
		if err != nil {
			return false
		}
		v2, err := semver.NewVersion(res[j].Chart.Version)
		if err != nil {
			return false
		}
		return v1.GreaterThan(v2)
	})
	sort.SliceStable(res, func(i, j int) bool { return res[i].Name < res[j].Name })

	dlTo := filepath.Join(tempDir, "download")
	if err := os.MkdirAll(dlTo, 0755); err != nil {
		return err
	}

	close(chanSearchChart)

	taskName := "Downloading"
	bar := p.AddBar(int64(len(res)),
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

	pendingDownload := make(chan *search.Result, len(res))
	for _, each := range res {
		pendingDownload <- each
	}
	close(pendingDownload)
	if d.Threads == 0 {
		d.Threads = 16
	}
	resultChan := make(chan *downloadResult, d.Threads)

	opts := action.ChartPathOptions{
		CaFile:                d.CAFile,
		CertFile:              d.CertFile,
		KeyFile:               d.KeyFile,
		InsecureSkipTLSverify: d.InsecureSkipTLSverify,
		Password:              d.Password,
		Username:              d.Username,
	}
	var mux sync.Mutex
	for i := 0; i < int(d.Threads); i++ {
		go download(bar, &mux, opts, r.CachePath, repoFile, dlTo, pendingDownload, resultChan)
	}

	for i := 0; i < len(res); i++ {
		result := <-resultChan
		if result == nil {
			continue
		}
		if result.Error != nil {
			e := errors.Wrap(result.Error, fmt.Sprintf("error downloading %s", strconv.Quote(result.Search.Name)))
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n\n", e)
			err = multierr.Append(err, e)
		}
	}

	httpClient := httputil.NewClient()
	downloaded, e := os.ReadDir(dlTo)
	if e != nil {
		return e
	}

	if len(downloaded) == 0 {
		p.Wait()
		return err
	}
	taskName = "Saving"
	bar = p.AddBar(int64(len(downloaded)),
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
	for _, each := range downloaded {
		since := time.Now()
		if d.Persistent {
			if e := modifyChart(httpClient, to, filepath.Join(dlTo, each.Name())); e != nil {
				e := errors.Wrap(e, fmt.Sprintf("error saving %s", strconv.Quote(each.Name())))
				fmt.Fprintf(os.Stderr, "[ERROR] %v\n\n", e)
				err = multierr.Append(err, e)
			}
		} else {
			if e := os.Rename(filepath.Join(dlTo, each.Name()), filepath.Join(to, each.Name())); e != nil {
				e := errors.Wrap(e, fmt.Sprintf("error saving %s", strconv.Quote(each.Name())))
				fmt.Fprintf(os.Stderr, "[ERROR] %v\n\n", e)
				err = multierr.Append(err, e)
			}
		}
		bar.Increment()
		safeDecoratorEwmaUpdate(bar, time.Since(since))
	}
	p.Wait()

	return err
}

func (d *chartDownloader) applyConstraint(res []*search.Result) ([]*search.Result, error) {
	if d.OnlyVersion == "" {
		if d.AllVersion {
			return res, nil
		}
		return d.filterLatest(res)
	}

	constraint, err := semver.NewConstraint(d.OnlyVersion)
	if err != nil {
		return res, errors.Wrap(err, "an invalid version/constraint format")
	}

	data := res[:0]
	for _, r := range res {
		v, err := semver.NewVersion(r.Chart.Version)
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			data = append(data, r)
		}
	}
	if d.AllVersion {
		return data, nil
	}
	return d.filterLatest(data)
}

func (d *chartDownloader) filterLatest(res []*search.Result) ([]*search.Result, error) {
	resMap := make(map[string]*search.Result)
	for _, re := range res {
		reo, ok := resMap[re.Name]
		if !ok {
			resMap[re.Name] = re
			continue
		}
		v1, _ := semver.NewVersion(reo.Chart.Version)
		v2, _ := semver.NewVersion(re.Chart.Version)
		if v2.GreaterThan(v1) {
			resMap[re.Name] = re
		}
	}
	data := res[:0]
	for _, r := range resMap {
		data = append(data, r)
	}
	return data, nil
}

func (d *chartDownloader) SetFlags(f *pflag.FlagSet) {
	f.StringVar(&d.Username, "username", "", "Chart repository username")
	f.StringVar(&d.Password, "password", "", "Chart repository password")
	f.StringVar(&d.CertFile, "cert-file", "", "Identify HTTPS client using this SSL certificate file")
	f.StringVar(&d.KeyFile, "key-file", "", "Identify HTTPS client using this SSL key file")
	f.StringVar(&d.CAFile, "ca-file", "", "Verify certificates of HTTPS-enabled servers using this CA bundle")
	f.BoolVar(&d.InsecureSkipTLSverify, "insecure-skip-tls-verify", false, "Skip TLS certificate checks for the repository")
	f.BoolVar(&d.Fuzzy, "fuzzy", true, "Use fuzzy search")
	f.BoolVar(&d.Regexp, "regexp", false, "Use regular expression for fuzzy search")
	f.BoolVar(&d.AllVersion, "all-version", false, "Download all version rather than the latest")
	f.StringVar(&d.OnlyVersion, "only-version", "", "Only download versions that match a given contraint rather than the latest")
	f.UintVar(&d.Threads, "threads", 16, "Threads to use for downloading contents")
	f.BoolVar(&d.Persistent, "persistent", true, "Attempt to convert avatar of charts to data url form")
}

func newChartDownloader() *chartDownloader {
	return &chartDownloader{}
}

func download(p *mpb.Bar, mux *sync.Mutex, opts action.ChartPathOptions, repoCache, repoCfg, destDir string, chartChan <-chan *search.Result, resultChan chan<- *downloadResult) {
	for c := range chartChan {
		since := time.Now()
		if c == nil {
			continue
		}

		pull := action.NewPullWithOpts(
			action.WithConfig(&action.Configuration{}),
			func(p *action.Pull) {
				p.ChartPathOptions = opts
				p.ChartPathOptions.Version = c.Chart.Version
				p.Settings = &cli.EnvSettings{
					RepositoryConfig: repoCfg,
					RepositoryCache:  repoCache,
				}
				p.DestDir = destDir
			},
		)
		_, err := pull.Run(c.Name)
		resultChan <- &downloadResult{
			Search: c,
			Error:  err,
		}
		mux.Lock()
		p.Increment()
		safeDecoratorEwmaUpdate(p, time.Since(since))
		mux.Unlock()
	}
}

func safeDecoratorEwmaUpdate(p *mpb.Bar, d time.Duration) {
	defer recover()
	p.DecoratorEwmaUpdate(d)
}

func modifyChart(httpClient *httputil.Client, saveDir, fp string) error {
	files, err := loadArchive(fp)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name == "Chart.yaml" {
			meta := new(chart.Metadata)
			if err := yaml.Unmarshal(file.Data, meta); err != nil {
				return errors.Wrap(err, "cannot load Chart.yaml")
			}
			u, err := url.Parse(meta.Icon)
			if err == nil {
				switch u.Scheme {
				case "http", "https":
					if err := util.Retry(func() error {
						dataUrl, err := downloadBlob(httpClient, meta.Icon)
						if err != nil {
							return err
						}
						meta.Icon = dataUrl.String()
						return nil
					}, 3, 0); err != nil {
						return errors.Wrap(err, "cannot download chart icon")
					}
				}
			}
			b, err := yaml.Marshal(meta)
			if err != nil {
				return errors.Wrap(err, "cannot save Chart.yaml")
			}
			file.Data = b
			break
		}
	}

	chrt, err := loader.LoadFiles(files)
	if err != nil {
		return err
	}
	_, err = chartutil.Save(chrt, saveDir)
	return err
}

func downloadBlob(httpClient *httputil.Client, url string) (*httputil.DataURL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	do, err := httpClient.New(ctx, url, httputil.GET)
	if err != nil {
		return nil, err
	}
	u := httputil.DataURL{Base64: true}
	err = do(func(resp *http.Response) error {
		u.RawData, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		u.MediaType = resp.Header.Get("Content-Type")
		return nil
	})
	return &u, err
}

func loadArchive(fp string) ([]*loader.BufferedFile, error) {
	fi, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	return loader.LoadArchiveFiles(fi)
}

type downloadResult struct {
	Search *search.Result
	Error  error
}
