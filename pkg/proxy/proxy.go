package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ReverseProxy struct {
	Proxy     *httputil.ReverseProxy
	OriginURL *url.URL
	ProxyHost string
}

func New(proxyHost, originURL string) (*ReverseProxy, error) {
	url, err := url.Parse(originURL)
	if err != nil {
		return nil, err
	}

	return &ReverseProxy{
		Proxy:     httputil.NewSingleHostReverseProxy(url),
		OriginURL: url,
		ProxyHost: proxyHost,
	}, nil
}

// Handler handles the http request using proxy.
func (rp *ReverseProxy) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for key := range r.Header {
			r.Header.Set(key, r.Header.Get(key))
		}
		r.Header.Set("referer", strings.Replace(r.Header.Get("referer"), rp.ProxyHost, rp.OriginURL.String(), 1))

		rp.Proxy.ServeHTTP(w, r)
	}
}
