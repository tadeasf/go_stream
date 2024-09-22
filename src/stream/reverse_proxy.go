package stream

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(targetURL string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Proxy", "go_stream")
	}

	return proxy, nil
}
