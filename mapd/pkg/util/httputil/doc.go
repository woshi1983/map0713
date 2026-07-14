// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package httputil provides HTTP client utilities.
package httputil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
)

func EZGetTLSCertificateChain(uri string, insecureSkipTLSVerify bool, caCerts ...[]byte) ([]*x509.Certificate, error) {
	return EZGetTLSCertificateChainContext(context.Background(), uri, insecureSkipTLSVerify, caCerts...)
}

func EZGetTLSCertificateChainContext(ctx context.Context, uri string, insecureSkipTLSVerify bool, caCerts ...[]byte) ([]*x509.Certificate, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if insecureSkipTLSVerify {
		cfg.InsecureSkipVerify = true
	}
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	for _, crt := range caCerts {
		rootCAs.AppendCertsFromPEM(crt)
	}
	cfg.RootCAs = rootCAs

	return GetTLSCertificateChainContext(ctx, uri, cfg)
}

func GetTLSCertificateChain(uri string, cfg *tls.Config) ([]*x509.Certificate, error) {
	return GetTLSCertificateChainContext(context.Background(), uri, cfg)
}

// GetTLSCertificateChainContext establish a TLS connection to uri and return its certificate chain.
func GetTLSCertificateChainContext(ctx context.Context, uri string, cfg *tls.Config) ([]*x509.Certificate, error) {
	certChan := make(chan []*x509.Certificate, 1)
	client := NewClient(TLSClientConfig(cfg), DialTLSContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &tls.Dialer{NetDialer: new(net.Dialer), Config: cfg}
		conn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		state := conn.(*tls.Conn).ConnectionState()
		var chain []*x509.Certificate
		if len(state.VerifiedChains) > 0 {
			for _, crts := range state.VerifiedChains {
				chain = append(chain, crts...)
			}
		}
		if len(chain) < 1 {
			chain = state.PeerCertificates
		}
		certChan <- chain
		close(certChan)

		return conn, nil
	}))

	do, err := client.New(ctx, uri, OPTIONS)
	if err != nil {
		return nil, err
	}

	err = do(nil)
	if err != nil {
		if len(certChan) > 0 {
			return <-certChan, nil
		}
		return nil, err
	}

	// proxied or non-TLS.
	if len(certChan) > 0 {
		return <-certChan, nil
	}
	return nil, nil
}
