/*
Copyright 2016 The Kubernetes Authors, VMware, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8s

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/apimachinery/pkg/util/runtime"
	restclient "k8s.io/client-go/rest"
	rc "k8s.io/client-go/tools/remotecommand"
	spdy "k8s.io/client-go/transport/spdy"
	"k8s.io/klog/v2"
)

// Executor is an interface for transporting shell-style streams.
// It is a extended version of [remotecommand.Executor](https://pkg.go.dev/k8s.io/client-go@v0.21.3/tools/remotecommand#Executor),
// which enables context-based connection control.
type Executor interface {
	// Stream initiates the transport of the standard shell streams. It will transport any
	// non-nil stream to a remote system, and return an error if a problem occurs. If tty
	// is set, the stderr stream is not used (raw TTY manages stdout and stderr over the
	// stdout stream).
	Stream(ctx context.Context, options rc.StreamOptions) (<-chan struct{}, error)
}

type streamCreator interface {
	CreateStream(headers http.Header) (httpstream.Stream, error)
}

type streamProtocolHandler interface {
	createStreams(conn streamCreator) error
	stream(conn streamCreator) error
}

// streamExecutor handles transporting standard shell streams over an httpstream connection.
type streamExecutor struct {
	upgrader  spdy.Upgrader
	transport http.RoundTripper

	method    string
	url       *url.URL
	protocols []string
}

// NewSPDYExecutor connects to the provided server and upgrades the connection to
// multiplexed bidirectional streams.
func NewSPDYExecutor(config *restclient.Config, method string, url *url.URL) (Executor, error) {
	wrapper, upgradeRoundTripper, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}
	return NewSPDYExecutorForTransports(wrapper, upgradeRoundTripper, method, url)
}

// NewSPDYExecutorForTransports connects to the provided server using the given transport,
// upgrades the response using the given upgrader to multiplexed bidirectional streams.
func NewSPDYExecutorForTransports(transport http.RoundTripper, upgrader spdy.Upgrader, method string, url *url.URL) (Executor, error) {
	return NewSPDYExecutorForProtocols(
		transport, upgrader, method, url,
		remotecommand.StreamProtocolV4Name,
		remotecommand.StreamProtocolV3Name,
		remotecommand.StreamProtocolV2Name,
		remotecommand.StreamProtocolV1Name,
	)
}

// NewSPDYExecutorForProtocols connects to the provided server and upgrades the connection to
// multiplexed bidirectional streams using only the provided protocols. Exposed for testing, most
// callers should use NewSPDYExecutor or NewSPDYExecutorForTransports.
func NewSPDYExecutorForProtocols(transport http.RoundTripper, upgrader spdy.Upgrader, method string, url *url.URL, protocols ...string) (Executor, error) {
	return &streamExecutor{
		upgrader:  upgrader,
		transport: transport,
		method:    method,
		url:       url,
		protocols: protocols,
	}, nil
}

// Stream opens a protocol streamer to the server and streams until a client closes
// the connection or the server disconnects.
func (e *streamExecutor) Stream(ctx context.Context, options rc.StreamOptions) (<-chan struct{}, error) {
	req, err := http.NewRequest(e.method, e.url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	conn, protocol, err := spdy.Negotiate(
		e.upgrader,
		&http.Client{Transport: e.transport},
		req,
		e.protocols...,
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var streamer streamProtocolHandler

	switch protocol {
	case remotecommand.StreamProtocolV4Name:
		streamer = newStreamProtocolV4(options)
	case remotecommand.StreamProtocolV3Name:
		streamer = newStreamProtocolV3(options)
	case remotecommand.StreamProtocolV2Name:
		streamer = newStreamProtocolV2(options)
	case "":
		klog.V(4).Infof("The server did not negotiate a streaming protocol version. Falling back to %s", remotecommand.StreamProtocolV1Name)
		fallthrough
	case remotecommand.StreamProtocolV1Name:
		streamer = newStreamProtocolV1(options)
	}

	if err := streamer.createStreams(conn); err != nil {
		return nil, err
	}

	doneCh := make(chan struct{}, 0)
	outCh := make(chan struct{}, 0)
	go func() {
		defer close(doneCh)
		if err := streamer.stream(conn); err != nil {
			klog.V(4).ErrorS(err, "Streaming error occured.")
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
		case <-doneCh:
		}
		conn.Close()
		close(outCh)
	}()

	return outCh, nil
}

// errorStreamDecoder interprets the data on the error channel and creates a go error object from it.
type errorStreamDecoder interface {
	decode(message []byte) error
}

// watchErrorStream watches the errorStream for remote command error data,
// decodes it with the given errorStreamDecoder, sends the decoded error (or nil if the remote
// command exited successfully) to the returned error channel, and closes it.
// This function returns immediately.
func watchErrorStream(errorStream io.Reader, d errorStreamDecoder) chan error {
	errorChan := make(chan error)

	go func() {
		defer runtime.HandleCrash()

		message, err := ioutil.ReadAll(errorStream)
		switch {
		case err != nil && err != io.EOF:
			errorChan <- fmt.Errorf("error reading from error stream: %s", err)
		case len(message) > 0:
			errorChan <- d.decode(message)
		default:
			errorChan <- nil
		}
		close(errorChan)
	}()

	return errorChan
}

// readerWrapper delegates to an io.Reader so that only the io.Reader interface is implemented,
// to keep io.Copy from doing things we don't want when copying from the reader to the data stream.
//
// If the Stdin io.Reader provided to remotecommand implements a WriteTo function (like bytes.Buffer does[1]),
// io.Copy calls that method[2] to attempt to write the entire buffer to the stream in one call.
// That results in an oversized call to spdystream.Stream#Write [3],
// which results in a single oversized data frame[4] that is too large.
//
// [1] https://golang.org/pkg/bytes/#Buffer.WriteTo
// [2] https://golang.org/pkg/io/#Copy
// [3] https://github.com/kubernetes/kubernetes/blob/90295640ef87db9daa0144c5617afe889e7992b2/vendor/github.com/docker/spdystream/stream.go#L66-L73
// [4] https://github.com/kubernetes/kubernetes/blob/90295640ef87db9daa0144c5617afe889e7992b2/vendor/github.com/docker/spdystream/spdy/write.go#L302-L304
type readerWrapper struct {
	reader io.Reader
}

func (r readerWrapper) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}
