// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package harbor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
)

// V2Client is a Harbor v2.0 API client.
type V2Client struct {
	Endpoint      *url.URL
	LoginUser     string
	LoginPassword []byte

	cli *httputil.Client
}

func (c *V2Client) wrapRequestOptionsOnDemand(opts ...httputil.HTTPRequestOption) []httputil.HTTPRequestOption {
	if c.LoginUser != "" {
		return append(opts, httputil.BasicAuth(c.LoginUser, string(c.LoginPassword)))
	}
	return opts
}

func (c *V2Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.PingContext(ctx)
}

// PingContext Harbor to check if it's alive.
func (c *V2Client) PingContext(ctx context.Context) error {
	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: "api/v2.0/ping"}
	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		return err
	})
}

func (c *V2Client) GetSystemInfo() (*SystemInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetSystemInfoContext(ctx)
}

func (c *V2Client) GetSystemInfoContext(ctx context.Context) (*SystemInfo, error) {
	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: "api/v2.0/systeminfo"}
	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
	if err != nil {
		return nil, err
	}
	d := &SystemInfo{}
	if err := do(func(resp *http.Response) error {
		b, err := validateResponse(resp)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, d); err != nil {
			return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return d, nil
}

func (c *V2Client) GetCurrentUser() (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetCurrentUserContext(ctx)
}

func (c *V2Client) GetCurrentUserContext(ctx context.Context) (*User, error) {
	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: "api/v2.0/users/current"}
	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
	if err != nil {
		return nil, err
	}
	d := &User{}
	if err := do(func(resp *http.Response) error {
		b, err := validateResponse(resp)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, d); err != nil {
			return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return d, nil
}

func (c *V2Client) PostProject(obj *CreateProject) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.PostProjectContext(ctx, obj)
}

func (c *V2Client) PostProjectContext(ctx context.Context, obj *CreateProject) error {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return err
	}

	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: "api/v2.0/projects"}
	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand(httputil.POST, body)...)
	if err != nil {
		return err
	}
	if err := do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (c *V2Client) HeadProject(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.HeadProjectContext(ctx, name)
}

func (c *V2Client) HeadProjectContext(ctx context.Context, name string) error {
	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: "api/v2.0/projects"}

	q := u.Query()
	q.Set("project_name", name)
	u.RawQuery = q.Encode()

	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand(httputil.HEAD)...)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		return err
	})
}

func (c *V2Client) GetProject(id int64) (*Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetProjectContext(ctx, id)
}

func (c *V2Client) GetProjectContext(ctx context.Context, id int64) (*Project, error) {
	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: filepath.Join("api/v2.0/projects", strconv.FormatInt(id, 10))}
	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
	if err != nil {
		return nil, err
	}
	d := &Project{}
	if err := do(func(resp *http.Response) error {
		b, err := validateResponse(resp)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, d); err != nil {
			return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return d, nil
}

func (c *V2Client) GetProjectsByName(name string) ([]*Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetProjectsByNameContext(ctx, name)
}

func (c *V2Client) GetProjectsByNameContext(ctx context.Context, name string) ([]*Project, error) {
	var list []*Project

	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: "api/v2.0/projects"}
	page, total := 1, 1
	for len(list) < total {
		q := u.Query()
		q.Set("page", strconv.Itoa(page))
		q.Set("page_size", "10")
		q.Set("name", name)
		q.Set("with_detail", "false")
		u.RawQuery = q.Encode()

		do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
		if err != nil {
			return nil, err
		}
		if err := do(func(resp *http.Response) error {
			b, err := validateResponse(resp)
			if err != nil {
				return err
			}
			var d []*Project
			if err := json.Unmarshal(b, &d); err != nil {
				return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
			}

			list = append(list, d...)

			total, _ = strconv.Atoi(resp.Header.Get("X-Total-Count"))
			page++
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *V2Client) GetProjects() ([]*Project, error) {
	return c.GetProjectsByName("")
}

func (c *V2Client) GetProjectsContext(ctx context.Context) ([]*Project, error) {
	return c.GetProjectsByNameContext(ctx, "")
}

func (c *V2Client) GetRepositories(projectName string) ([]*Repository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetRepositoriesContext(ctx, projectName)
}

func (c *V2Client) GetRepositoriesContext(ctx context.Context, projectName string) ([]*Repository, error) {
	var list []*Repository

	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: filepath.Join("api/v2.0/projects", projectName, "repositories")}
	page, total := 1, 1
	for len(list) < total {
		q := u.Query()
		q.Set("page", strconv.Itoa(page))
		q.Set("page_size", "10")
		u.RawQuery = q.Encode()

		do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
		if err != nil {
			return nil, err
		}
		if err := do(func(resp *http.Response) error {
			b, err := validateResponse(resp)
			if err != nil {
				return err
			}
			var d []*Repository
			if err := json.Unmarshal(b, &d); err != nil {
				return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
			}

			list = append(list, d...)

			total, _ = strconv.Atoi(resp.Header.Get("X-Total-Count"))
			page++
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *V2Client) GetArtifacts(projectName, repoName string) ([]*Artifact, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetArtifactsContext(ctx, projectName, repoName)
}

func (c *V2Client) GetArtifactsContext(ctx context.Context, projectName, repoName string) ([]*Artifact, error) {
	var list []*Artifact

	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: filepath.Join("api/v2.0/projects", projectName, "repositories", repoName, "artifacts")}
	page, total := 1, 1
	for len(list) < total {
		q := u.Query()
		q.Set("page", strconv.Itoa(page))
		q.Set("page_size", "10")
		q.Set("with_tag", "true")
		u.RawQuery = q.Encode()

		do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
		if err != nil {
			return nil, err
		}
		if err := do(func(resp *http.Response) error {
			b, err := validateResponse(resp)
			if err != nil {
				return err
			}
			var d []*Artifact
			if err := json.Unmarshal(b, &d); err != nil {
				return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
			}

			list = append(list, d...)

			total, _ = strconv.Atoi(resp.Header.Get("X-Total-Count"))
			page++
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return list, nil
}

// HelmChartRepo returns the root path of helm chart repository.
func (c *V2Client) HelmChartRepo(projectName string) *url.URL {
	return &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: filepath.Join("chartrepo", projectName)}
}

func (c *V2Client) GetChartRepos(projectName string) ([]*ChartRepo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetChartReposContext(ctx, projectName)
}

func (c *V2Client) GetChartReposContext(ctx context.Context, projectName string) ([]*ChartRepo, error) {
	var list []*ChartRepo

	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: filepath.Join("api/chartrepo", projectName, "charts")}
	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
	if err != nil {
		return nil, err
	}
	if err := do(func(resp *http.Response) error {
		b, err := validateResponse(resp)
		if err != nil {
			if IsErrObjectNotExists(err) {
				return nil
			}
			return err
		}
		if err := json.Unmarshal(b, &list); err != nil {
			return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *V2Client) GetCharts(projectName, chartName string) ([]*ChartMeta, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetChartsContext(ctx, projectName, chartName)
}

func (c *V2Client) GetChartsContext(ctx context.Context, projectName, chartName string) ([]*ChartMeta, error) {
	var list []*ChartMeta

	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: filepath.Join("api/chartrepo", projectName, "charts", chartName)}
	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
	if err != nil {
		return nil, err
	}
	if err := do(func(resp *http.Response) error {
		b, err := validateResponse(resp)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &list); err != nil {
			return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *V2Client) GetChart(projectName, chartName, version string) (*Chart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return c.GetChartContext(ctx, projectName, chartName, version)
}

func (c *V2Client) GetChartContext(ctx context.Context, projectName, chartName, version string) (*Chart, error) {
	u := &url.URL{Scheme: c.Endpoint.Scheme, Opaque: c.Endpoint.Opaque, Host: c.Endpoint.Host, Path: filepath.Join("api/chartrepo", projectName, "charts", chartName, version)}
	do, err := c.cli.New(ctx, u.String(), c.wrapRequestOptionsOnDemand()...)
	if err != nil {
		return nil, err
	}
	d := &Chart{}
	if err := do(func(resp *http.Response) error {
		b, err := validateResponse(resp)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, d); err != nil {
			return &ErrAPICall{Method: resp.Request.Method, URL: resp.Request.URL, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("error parsing body: %v", err)}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return d, nil
}

// NewV2Client constructs a new V2Client.
func NewV2Client(endpoint, user string, password []byte, insecure bool, caCerts ...[]byte) (*V2Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return &V2Client{
		Endpoint:      u,
		LoginUser:     user,
		LoginPassword: password,

		cli: httputil.NewClient(
			httputil.InsecureSkipVerify(insecure),
			httputil.TrustRootCACerts(caCerts...),
		),
	}, nil
}
