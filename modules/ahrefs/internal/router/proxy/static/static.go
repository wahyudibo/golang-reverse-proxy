package static

import (
	// "fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/debugger"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/proxy"
)

const baseURL = "https://static.ahrefs.com"

type Service struct {
	Config *config.Config
	RP     *proxy.ReverseProxy
}

func New(cfg *config.Config) (*Service, error) {
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	rp, err := proxy.New(cfg.ServerAddress, url.String())
	if err != nil {
		return nil, err
	}

	if cfg.DebugMode {
		rp.Proxy.Transport = debugger.DebugTransport{}
	}

	director := rp.Proxy.Director
	rp.Proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = rp.OriginURL.Host
		req.URL.Host = rp.OriginURL.Host
		req.URL.Scheme = "https"
	}

	s := &Service{Config: cfg, RP: rp}

	return s, nil
}

// Handler handles the http request using proxy.
func (s *Service) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for key := range r.Header {
			r.Header.Set(key, r.Header.Get(key))
		}
		r.Header.Set("referer", strings.Replace(r.Header.Get("referer"), s.RP.ProxyHost, s.RP.OriginURL.String(), 1))
		r.URL.Path = strings.TrimPrefix(r.URL.RequestURI(), "/ahx-static")

		// fmt.Printf("EXAMPLE: %s", strings.SplitN(r.URL.RequestURI()[1:], "/", 2))

		s.RP.Proxy.ServeHTTP(w, r)
	}
}
