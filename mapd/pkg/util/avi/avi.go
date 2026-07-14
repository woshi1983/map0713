// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package avi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
)

type AVI struct {
	Endpoint      *url.URL
	LoginUser     string
	LoginPassword []byte

	mux sync.Mutex
	cli *httputil.Client
}

func (a *AVI) login(ctx context.Context) error {
	body, err := httputil.JSONBody(&Login{Username: a.LoginUser, Password: string(a.LoginPassword)})
	if err != nil {
		return err
	}

	u := &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: "login"}
	do, err := a.cli.New(ctx, u.String(), httputil.POST, body)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *AVI) logout(ctx context.Context) error {
	u := &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: "logout"}
	do, err := a.cli.New(ctx, u.String(), httputil.POST)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *AVI) new(ctx context.Context, url string, opts ...httputil.HTTPRequestOption) (httputil.HTTPRequestFunc, error) {
	do, err := a.cli.New(ctx, url, opts...)
	if err != nil {
		return nil, err
	}
	return func(fn httputil.HTTPResponseCallbackFunc) error {
		a.mux.Lock()
		defer a.mux.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := a.login(ctx); err != nil {
			return err
		}

		err := do(fn)

		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			a.logout(ctx)
		}()

		return err
	}, nil
}

func (a *AVI) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.PingContext(ctx)
}

func (a *AVI) PingContext(ctx context.Context) error {
	a.mux.Lock()
	defer a.mux.Unlock()

	if err := a.login(ctx); err != nil {
		return err
	}

	func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		a.logout(ctx)
	}()

	return nil
}

func (a *AVI) CreatePool(obj *Pool) (*Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.CreatePoolContext(ctx, obj)
}

func (a *AVI) CreatePoolContext(ctx context.Context, obj *Pool) (*Pool, error) {
	for _, each := range obj.Servers {
		if each.IP == nil {
			continue
		}
		ip := net.ParseIP(each.IP.Addr)
		if ip.To4() != nil {
			each.IP.Type = "V4"
		} else {
			each.IP.Type = "V6"
		}
	}
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.PoolRef("").String(), httputil.POST, body)
	if err != nil {
		return nil, err
	}
	d := &Pool{}
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

func (a *AVI) UpdatePool(obj *Pool) (*Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.UpdatePoolContext(ctx, obj)
}

func (a *AVI) UpdatePoolContext(ctx context.Context, obj *Pool) (*Pool, error) {
	for _, each := range obj.Servers {
		if each.IP == nil {
			continue
		}
		ip := net.ParseIP(each.IP.Addr)
		if ip.To4() != nil {
			each.IP.Type = "V4"
		} else {
			each.IP.Type = "V6"
		}
	}
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.PoolRef(obj.UUID).String(), httputil.PUT, body)
	if err != nil {
		return nil, err
	}
	d := &Pool{}
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

func (a *AVI) GetPool(uuid string) (*Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.GetPoolContext(ctx, uuid)
}

func (a *AVI) GetPoolContext(ctx context.Context, uuid string) (*Pool, error) {
	do, err := a.new(ctx, a.PoolRef(uuid).String())
	if err != nil {
		return nil, err
	}
	d := &Pool{}
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

func (a *AVI) DeletePool(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.DeletePoolContext(ctx, uuid)
}

func (a *AVI) DeletePoolContext(ctx context.Context, uuid string) error {
	do, err := a.new(ctx, a.PoolRef(uuid).String(), httputil.DELETE)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *AVI) PoolRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/pool", uuid)}
}

func (a *AVI) CreatePoolGroup(obj *PoolGroup) (*PoolGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.CreatePoolGroupContext(ctx, obj)
}

func (a *AVI) CreatePoolGroupContext(ctx context.Context, obj *PoolGroup) (*PoolGroup, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.PoolGroupRef("").String(), httputil.POST, body)
	if err != nil {
		return nil, err
	}
	d := &PoolGroup{}
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

func (a *AVI) UpdatePoolGroup(obj *PoolGroup) (*PoolGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.UpdatePoolGroupContext(ctx, obj)
}

func (a *AVI) UpdatePoolGroupContext(ctx context.Context, obj *PoolGroup) (*PoolGroup, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.PoolGroupRef(obj.UUID).String(), httputil.PUT, body)
	if err != nil {
		return nil, err
	}
	d := &PoolGroup{}
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

func (a *AVI) DeletePoolGroup(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.DeletePoolGroupContext(ctx, uuid)
}

func (a *AVI) DeletePoolGroupContext(ctx context.Context, uuid string) error {
	do, err := a.new(ctx, a.PoolGroupRef(uuid).String(), httputil.DELETE)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *AVI) PoolGroupRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/poolgroup", uuid)}
}

func (a *AVI) CreateVSVIP(obj *VSVIP) (*VSVIP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.CreateVSVIPContext(ctx, obj)
}

func (a *AVI) CreateVSVIPContext(ctx context.Context, obj *VSVIP) (*VSVIP, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.VSVIPRef("").String(), httputil.POST, body)
	if err != nil {
		return nil, err
	}
	d := &VSVIP{}
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

func (a *AVI) GetVSVIP(uuid string) (*VSVIP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.GetVSVIPContext(ctx, uuid)
}

func (a *AVI) GetVSVIPContext(ctx context.Context, uuid string) (*VSVIP, error) {
	do, err := a.new(ctx, a.VSVIPRef(uuid).String())
	if err != nil {
		return nil, err
	}
	d := &VSVIP{}
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

func (a *AVI) ListVSVIP(options *ListVSVIPOptions) (*VSVIPList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.ListVSVIPContext(ctx, options)
}

func (a *AVI) ListVSVIPContext(ctx context.Context, options *ListVSVIPOptions) (*VSVIPList, error) {
	u := a.VSVIPRef("")
	if options != nil {
		q := u.Query()
		if options.Name != "" {
			q.Set("name", options.Name)
		}
		u.RawQuery = q.Encode()
	}

	do, err := a.new(ctx, u.String())
	if err != nil {
		return nil, err
	}
	d := &VSVIPList{}
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

func (a *AVI) DeleteVSVIP(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.DeleteVSVIPContext(ctx, uuid)
}

func (a *AVI) DeleteVSVIPContext(ctx context.Context, uuid string) error {
	do, err := a.new(ctx, a.VSVIPRef(uuid).String(), httputil.DELETE)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *AVI) VSVIPRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/vsvip", uuid)}
}

func (a *AVI) CreateVirtualService(obj *VirtualService) (*VirtualService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.CreateVirtualServiceContext(ctx, obj)
}

func (a *AVI) CreateVirtualServiceContext(ctx context.Context, obj *VirtualService) (*VirtualService, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.VirtualServiceRef("").String(), httputil.POST, body)
	if err != nil {
		return nil, err
	}
	d := &VirtualService{}
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

func (a *AVI) UpdateVirtualService(obj *VirtualService) (*VirtualService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.UpdateVirtualServiceContext(ctx, obj)
}

func (a *AVI) UpdateVirtualServiceContext(ctx context.Context, obj *VirtualService) (*VirtualService, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.VirtualServiceRef(obj.UUID).String(), httputil.PUT, body)
	if err != nil {
		return nil, err
	}
	d := &VirtualService{}
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

func (a *AVI) DeleteVirtualService(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.DeleteVirtualServiceContext(ctx, uuid)
}

func (a *AVI) DeleteVirtualServiceContext(ctx context.Context, uuid string) error {
	do, err := a.new(ctx, a.VirtualServiceRef(uuid).String(), httputil.DELETE)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *AVI) VirtualServiceRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/virtualservice", uuid)}
}

func (a *AVI) CreateApplicationProfile(obj *ApplicationProfile) (*ApplicationProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.CreateApplicationProfileContext(ctx, obj)
}

func (a *AVI) CreateApplicationProfileContext(ctx context.Context, obj *ApplicationProfile) (*ApplicationProfile, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.ApplicationProfileRef("").String(), httputil.POST, body)
	if err != nil {
		return nil, err
	}
	d := &ApplicationProfile{}
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

func (a *AVI) UpdateApplicationProfile(obj *ApplicationProfile) (*ApplicationProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.UpdateApplicationProfileContext(ctx, obj)
}

func (a *AVI) UpdateApplicationProfileContext(ctx context.Context, obj *ApplicationProfile) (*ApplicationProfile, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.ApplicationProfileRef(obj.UUID).String(), httputil.PUT, body)
	if err != nil {
		return nil, err
	}
	d := &ApplicationProfile{}
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

func (a *AVI) ListApplicationProfile(options *ListApplicationProfileOptions) (*ApplicationProfileList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.ListApplicationProfileContext(ctx, options)
}

func (a *AVI) ListApplicationProfileContext(ctx context.Context, options *ListApplicationProfileOptions) (*ApplicationProfileList, error) {
	u := a.ApplicationProfileRef("")
	if options != nil {
		q := u.Query()
		if options.Name != "" {
			q.Set("name", options.Name)
		}
		u.RawQuery = q.Encode()
	}

	do, err := a.new(ctx, u.String())
	if err != nil {
		return nil, err
	}
	d := &ApplicationProfileList{}
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

func (a *AVI) DeleteApplicationProfile(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.DeleteApplicationProfileContext(ctx, uuid)
}

func (a *AVI) DeleteApplicationProfileContext(ctx context.Context, uuid string) error {
	do, err := a.new(ctx, a.ApplicationProfileRef(uuid).String(), httputil.DELETE)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *AVI) ApplicationProfileRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/applicationprofile", uuid)}
}

func (a *AVI) CreateSEGroup(obj *ServiceEngineGroup) (*ServiceEngineGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.CreateSEGroupContext(ctx, obj)
}

func (a *AVI) CreateSEGroupContext(ctx context.Context, obj *ServiceEngineGroup) (*ServiceEngineGroup, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.SEGroupRef("").String(), httputil.POST, body)
	if err != nil {
		return nil, err
	}
	d := &ServiceEngineGroup{}
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

func (a *AVI) UpdateSEGroup(obj *ServiceEngineGroup) (*ServiceEngineGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.UpdateSEGroupContext(ctx, obj)
}

func (a *AVI) UpdateSEGroupContext(ctx context.Context, obj *ServiceEngineGroup) (*ServiceEngineGroup, error) {
	body, err := httputil.JSONBody(obj)
	if err != nil {
		return nil, err
	}

	do, err := a.new(ctx, a.SEGroupRef(obj.UUID).String(), httputil.PUT, body)
	if err != nil {
		return nil, err
	}
	d := &ServiceEngineGroup{}
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

func (a *AVI) GetSEGroup(uuid string) (*ServiceEngineGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.GetSEGroupContext(ctx, uuid)
}

func (a *AVI) GetSEGroupContext(ctx context.Context, uuid string) (*ServiceEngineGroup, error) {
	do, err := a.new(ctx, a.SEGroupRef(uuid).String())
	if err != nil {
		return nil, err
	}
	d := &ServiceEngineGroup{}
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

func (a *AVI) ListSEGroup(options *ListSEGroupOptions) (*ServiceEngineGroupList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.ListSEGroupContext(ctx, options)
}

func (a *AVI) ListSEGroupContext(ctx context.Context, options *ListSEGroupOptions) (*ServiceEngineGroupList, error) {
	u := a.SEGroupRef("")
	if options != nil {
		q := u.Query()
		if options.Name != "" {
			q.Set("name", options.Name)
		}
		if options.CloudName != "" {
			q.Set("cloud_ref.name", options.CloudName)
		}
		u.RawQuery = q.Encode()
	}
	do, err := a.new(ctx, u.String())
	if err != nil {
		return nil, err
	}
	d := &ServiceEngineGroupList{}
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

func (a *AVI) DeleteSEGroup(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.DeleteSEGroupContext(ctx, uuid)
}

func (a *AVI) DeleteSEGroupContext(ctx context.Context, uuid string) error {
	do, err := a.new(ctx, a.SEGroupRef(uuid).String(), httputil.DELETE)
	if err != nil {
		return err
	}
	return do(func(resp *http.Response) error {
		_, err := validateResponse(resp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (a *AVI) SEGroupRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/serviceenginegroup", uuid)}
}

func (a *AVI) GetNetworkProfile(uuid string) (*NetworkProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.GetNetworkProfileContext(ctx, uuid)
}

func (a *AVI) GetNetworkProfileContext(ctx context.Context, uuid string) (*NetworkProfile, error) {
	do, err := a.new(ctx, a.NetworkProfileRef(uuid).String())
	if err != nil {
		return nil, err
	}
	d := &NetworkProfile{}
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

func (a *AVI) ListNetworkProfile(options *ListNetworkProfileOptions) (*NetworkProfileList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.ListNetworkProfileContext(ctx, options)
}

func (a *AVI) ListNetworkProfileContext(ctx context.Context, options *ListNetworkProfileOptions) (*NetworkProfileList, error) {
	u := a.NetworkProfileRef("")
	if options != nil {
		q := u.Query()
		if options.Name != "" {
			q.Set("name", options.Name)
		}
		u.RawQuery = q.Encode()
	}
	do, err := a.new(ctx, u.String())
	if err != nil {
		return nil, err
	}
	d := &NetworkProfileList{}
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

func (a *AVI) NetworkProfileRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/networkprofile", uuid)}
}

func (a *AVI) GetNetwork(uuid string) (*Network, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.GetNetworkContext(ctx, uuid)
}

func (a *AVI) GetNetworkContext(ctx context.Context, uuid string) (*Network, error) {
	do, err := a.new(ctx, a.NetworkRef(uuid).String())
	if err != nil {
		return nil, err
	}
	d := &Network{}
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

func (a *AVI) ListNetwork(options *ListNetworkOptions) (*NetworkList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.ListNetworkContext(ctx, options)
}

func (a *AVI) ListNetworkContext(ctx context.Context, options *ListNetworkOptions) (*NetworkList, error) {
	u := a.NetworkRef("")
	if options != nil {
		q := u.Query()
		if options.Name != "" {
			q.Set("name", options.Name)
		}
		if options.CloudName != "" {
			q.Set("cloud_ref.name", options.CloudName)
		}
		u.RawQuery = q.Encode()
	}
	do, err := a.new(ctx, u.String())
	if err != nil {
		return nil, err
	}
	d := &NetworkList{}
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

func (a *AVI) NetworkRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/network", uuid)}
}

func (a *AVI) GetNetworkRuntime(uuid string) (*NetworkRuntime, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.GetNetworkRuntimeContext(ctx, uuid)
}

func (a *AVI) GetNetworkRuntimeContext(ctx context.Context, uuid string) (*NetworkRuntime, error) {
	do, err := a.new(ctx, a.NetworkRuntimeRef(uuid).String())
	if err != nil {
		return nil, err
	}
	d := &NetworkRuntime{}
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

func (a *AVI) NetworkRuntimeRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/networkruntime", uuid)}
}

func (a *AVI) GetCloud(uuid string) (*Cloud, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.GetCloudContext(ctx, uuid)
}

func (a *AVI) GetCloudContext(ctx context.Context, uuid string) (*Cloud, error) {
	do, err := a.new(ctx, a.CloudRef(uuid).String())
	if err != nil {
		return nil, err
	}
	d := &Cloud{}
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

func (a *AVI) ListCloud(options *ListCloudOptions) (*CloudList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.ListCloudContext(ctx, options)
}

func (a *AVI) ListCloudContext(ctx context.Context, options *ListCloudOptions) (*CloudList, error) {
	u := a.CloudRef("")
	if options != nil {
		q := u.Query()
		if options.Name != "" {
			q.Set("name", options.Name)
		}
		u.RawQuery = q.Encode()
	}
	do, err := a.new(ctx, u.String())
	if err != nil {
		return nil, err
	}
	d := &CloudList{}
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

func (a *AVI) CloudRef(uuid string) *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: filepath.Join("api/cloud", uuid)}
}

func (a *AVI) GetLicenseLedgerDetails() (*LicenseLedgerDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return a.GetLicenseLedgerDetailsContext(ctx)
}

func (a *AVI) GetLicenseLedgerDetailsContext(ctx context.Context) (*LicenseLedgerDetails, error) {
	do, err := a.new(ctx, a.LicenseLedgerDetailsRef().String())
	if err != nil {
		return nil, err
	}
	d := &LicenseLedgerDetails{}
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

func (a *AVI) LicenseLedgerDetailsRef() *url.URL {
	return &url.URL{Scheme: a.Endpoint.Scheme, Opaque: a.Endpoint.Opaque, Host: a.Endpoint.Host, Path: "api/licensing/ledger/details"}
}

// NewAVI constructs a new AVI.
func NewAVI(endpoint, user string, password []byte, insecureSkipTLSVerify bool, caCerts ...[]byte) (*AVI, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	referer := &url.URL{Scheme: u.Scheme, Opaque: u.Opaque, Host: u.Host}

	return &AVI{
		Endpoint:      u,
		LoginUser:     user,
		LoginPassword: password,

		cli: httputil.NewClient(
			httputil.EnableCookie(true),
			httputil.InsecureSkipVerify(insecureSkipTLSVerify),
			httputil.TrustRootCACerts(caCerts...),
		).Options(
			httputil.SetHeader("X-Avi-Version", "20.1.6"),
			httputil.SetHeader("Referer", referer.String()),
			httputil.CSRFToken("X-CSRFToken", "csrftoken"),
		),
	}, nil
}
