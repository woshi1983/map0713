// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package httputil

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"golang.org/x/net/publicsuffix"
)

// Client is a generic purpose HTTP client.
type Client struct {
	*http.Client

	opt []HTTPRequestOption
}

// Options injects opts into client as default request options.
func (c *Client) Options(opts ...HTTPRequestOption) *Client {
	c.opt = append(c.opt, opts...)
	return c
}

// New factory a new HTTPRequestFunc for advanced callback handling.
func (c *Client) New(ctx context.Context, url string, opts ...HTTPRequestOption) (HTTPRequestFunc, error) {
	cfg := &HTTPRequestConfig{Method: "GET", URL: url}
	opts = append(c.opt, opts...)
	for i := range opts {
		opts[i](cfg)
	}
	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, cfg.Body)
	if err != nil {
		return nil, err
	}
	return func(fn HTTPResponseCallbackFunc) error {
		for i := range cfg.fns {
			cfg.fns[i](c.Client, req)
		}
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if fn != nil {
			if err := fn(resp); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

// NewClient returns a new Client.
func NewClient(opts ...HTTPClientOption) *Client {
	cfg := &HTTPClientConfig{}
	for i := range opts {
		opts[i](cfg)
	}
	return &Client{Client: NewHTTPClient(cfg)}
}

// NewHTTPClient is an insecured HTTP client for generic purpose of use.
func NewHTTPClient(cfg *HTTPClientConfig) *http.Client {
	if cfg == nil {
		cfg = DefaultClientConfig()
	}

	cli := &http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           cfg.DialContext,
		TLSClientConfig:       cfg.TLSClientConfig,
		DialTLSContext:        cfg.DialTLSContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}}
	if cfg.EnableCookie {
		jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		cli.Jar = jar
	}
	return cli
}

// HEAD marks the request as a HEAD request.
func HEAD(cfg *HTTPRequestConfig) {
	cfg.Method = "HEAD"
}

// OPTIONS marks the request as a OPTIONS request.
func OPTIONS(cfg *HTTPRequestConfig) {
	cfg.Method = "OPTIONS"
}

// GET marks the request as a GET request.
func GET(cfg *HTTPRequestConfig) {
	cfg.Method = "GET"
}

// POST marks the request as a POST request.
func POST(cfg *HTTPRequestConfig) {
	cfg.Method = "POST"
}

// PUT marks the request as a PUT request.
func PUT(cfg *HTTPRequestConfig) {
	cfg.Method = "PUT"
}

// PATCH marks the request as a PATCH request.
func PATCH(cfg *HTTPRequestConfig) {
	cfg.Method = "PATCH"
}

// DELETE marks the request as a DELETE request.
func DELETE(cfg *HTTPRequestConfig) {
	cfg.Method = "DELETE"
}

// AddHeader adds value to the given header key.
func AddHeader(key, value string) HTTPRequestOption {
	return func(cfg *HTTPRequestConfig) {
		cfg.fns = append(cfg.fns, func(cli *http.Client, req *http.Request) {
			req.Header.Add(key, value)
		})
	}
}

// SetHeader set value for the given header key.
func SetHeader(key, value string) HTTPRequestOption {
	return func(cfg *HTTPRequestConfig) {
		cfg.fns = append(cfg.fns, func(cli *http.Client, req *http.Request) {
			req.Header.Set(key, value)
		})
	}
}

// DelHeader deletes value from the given header key.
func DelHeader(key string) HTTPRequestOption {
	return func(cfg *HTTPRequestConfig) {
		cfg.fns = append(cfg.fns, func(cli *http.Client, req *http.Request) {
			req.Header.Del(key)
		})
	}
}

// Body sets request body with given reader r.
func Body(r io.Reader) HTTPRequestOption {
	return func(cfg *HTTPRequestConfig) {
		cfg.Body = r
	}
}

// JSONBody is revelant to Body but also set header.
func JSONBody(data interface{}) (HTTPRequestOption, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)
	return func(cfg *HTTPRequestConfig) {
		Body(r)(cfg)
		SetHeader("Content-Type", "application/json")(cfg)
	}, nil
}

// BasicAuth enables Basic authorization.
func BasicAuth(username, password string) HTTPRequestOption {
	return func(cfg *HTTPRequestConfig) {
		cfg.fns = append(cfg.fns, func(cli *http.Client, req *http.Request) {
			req.SetBasicAuth(username, password)
		})
	}
}

// CSRFToken reads from cookie and set them into header.
func CSRFToken(header, cookie string) HTTPRequestOption {
	return func(cfg *HTTPRequestConfig) {
		cfg.fns = append(cfg.fns, func(cli *http.Client, req *http.Request) {
			if cli.Jar == nil {
				return
			}

			cookies := cli.Jar.Cookies(req.URL)
			for _, each := range cookies {
				if each.Name == cookie {
					req.Header.Set(header, each.Value)
				}
			}
		})
	}
}

// HTTPRequestConfig is a config for constructing HTTP request.
type HTTPRequestConfig struct {
	Method string
	URL    string
	Body   io.Reader

	fns []func(*http.Client, *http.Request)
}

// HTTPRequestOption is a middleware of constructing HTTP request.
type HTTPRequestOption func(*HTTPRequestConfig)

// HTTPRequestFunc is used for performing actual HTTP request, and interact
// the given callback function to parse the response.
type HTTPRequestFunc func(HTTPResponseCallbackFunc) error

// HTTPResponseCallbackFunc will be called within HTTPRequestFunc to parse
// the HTTP response.
type HTTPResponseCallbackFunc func(*http.Response) error

// DialContext overrides the default DialContext function of HTTP client.
func DialContext(fn func(ctx context.Context, network string, addr string) (net.Conn, error)) HTTPClientOption {
	return func(cfg *HTTPClientConfig) {
		cfg.DialContext = fn
	}
}

// TLSClientConfig overrides the default TLSClientConfig function of HTTP client.
func TLSClientConfig(config *tls.Config) HTTPClientOption {
	return func(cfg *HTTPClientConfig) {
		cfg.TLSClientConfig = config
	}
}

// InsecureSkipVerify determines whether to skip TLS verification. This is a
// shorthand of TLSClientConfig().
func InsecureSkipVerify(yes bool) HTTPClientOption {
	return func(cfg *HTTPClientConfig) {
		if cfg.TLSClientConfig == nil {
			cfg.TLSClientConfig = defaultTLSClientConfig()
		}
		cfg.TLSClientConfig.InsecureSkipVerify = yes
	}
}

// TrustRootCACerts append additional root CA certificates to trusted CA
// certificate pool. This is a shorthand of TLSClientConfig().
func TrustRootCACerts(certs ...[]byte) HTTPClientOption {
	return func(cfg *HTTPClientConfig) {
		if cfg.TLSClientConfig == nil {
			cfg.TLSClientConfig = defaultTLSClientConfig()
		}
		for i := range certs {
			cfg.TLSClientConfig.RootCAs.AppendCertsFromPEM(certs[i])
		}
	}
}

// DialTLSContext overrides the default DialTLSContext function of HTTP client.
func DialTLSContext(fn func(ctx context.Context, network, addr string) (net.Conn, error)) HTTPClientOption {
	return func(cfg *HTTPClientConfig) {
		cfg.DialTLSContext = fn
	}
}

// EnableCookie determines whether to enable cookie for HTTP client.
func EnableCookie(yes bool) HTTPClientOption {
	return func(cfg *HTTPClientConfig) {
		cfg.EnableCookie = yes
	}
}

// HTTPClientConfig is a config for constructing HTTP client.
type HTTPClientConfig struct {
	DialContext     func(ctx context.Context, network string, addr string) (net.Conn, error)
	TLSClientConfig *tls.Config
	DialTLSContext  func(ctx context.Context, network, addr string) (net.Conn, error)
	EnableCookie    bool
}

type HTTPClientOption func(*HTTPClientConfig)

// DefaultClientConfig generates a default configuration for creating HTTP Client.
func DefaultClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		TLSClientConfig: defaultTLSClientConfig(),
	}
}

func defaultTLSClientConfig() *tls.Config {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if mapd.GlobalConfig.HTTPClientCACert != "" {
		certs, err := os.ReadFile(mapd.GlobalConfig.HTTPClientCACert)
		if err == nil {
			rootCAs.AppendCertsFromPEM(certs)
		}
	}

	return &tls.Config{
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}
}
