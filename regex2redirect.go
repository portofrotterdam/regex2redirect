// Package regex2redirect is a plugin
package regex2redirect

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
)

// Config the plugin configuration.
type Config struct {
	Regex string `json:"regex"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// RegexRedirect a Traefik plugin.
type RegexRedirect struct {
	next  http.Handler
	regex *regexp.Regexp
}

// HTTPClient is a reduced interface for http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// New creates a new Regex2Redirect plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &RegexRedirect{regex: regexp.MustCompile(config.Regex), next: next}, nil
}

func (c *RegexRedirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.Header.Set("Accept-Encoding", "identity")
	wrappedWriter := &responseBuffer{
		ResponseWriter: rw,
	}

	c.next.ServeHTTP(wrappedWriter, req)

	bodyBytes := wrappedWriter.bodyBuffer.Bytes()

	contentEncoding := wrappedWriter.Header().Get("Content-Encoding")
	if contentEncoding != "" && contentEncoding != "identity" {
		if _, err := rw.Write(bodyBytes); err != nil {
			log.Printf("Content encoding not supported by : %v", err)
		}
		rw.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = rw.Write([]byte("Content encoding not supported"))
		return
	}
	result := c.regex.Find(bodyBytes)

	if result == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	redirectURL, err := url.Parse(string(result))
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}

	rw.Header().Set("Location", redirectURL.String())
	rw.WriteHeader(http.StatusTemporaryRedirect)
}

type responseBuffer struct {
	bodyBuffer bytes.Buffer
	statusCode int

	http.ResponseWriter
}

func (r *responseBuffer) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *responseBuffer) Write(p []byte) (int, error) {
	if r.statusCode == 0 {
		r.WriteHeader(http.StatusOK)
	}

	return r.bodyBuffer.Write(p)
}

func (r *responseBuffer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.ResponseWriter)
	}

	return hijacker.Hijack()
}

func (r *responseBuffer) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
