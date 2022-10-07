package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v9"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/alias"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/debugger"
	enc "github.com/wahyudibo/golang-reverse-proxy/pkg/encoding"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/proxy"
)

type Service struct {
	Config *config.Config
	Cache  *redis.Client
	RP     *proxy.ReverseProxy
}

func New(cfg *config.Config, cache *redis.Client) (*Service, error) {
	url, err := url.Parse(alias.AppDomain)
	if err != nil {
		return nil, err
	}

	rp, err := proxy.New(cfg.ProxyServerAddress, url.String())
	if err != nil {
		return nil, err
	}

	if cfg.ProxyDebugMode {
		rp.Proxy.Transport = debugger.DebugTransport{}
	}

	director := rp.Proxy.Director
	rp.Proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = rp.OriginURL.Host
		req.URL.Host = rp.OriginURL.Host
		req.URL.Scheme = "https"
	}

	s := &Service{Config: cfg, Cache: cache, RP: rp}

	rp.Proxy.ModifyResponse = s.TransformResponse

	return s, nil
}

// Handler handles the http request using proxy.
func (s *Service) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for key := range r.Header {
			r.Header.Set(key, r.Header.Get(key))
		}

		r.Header.Set("user-agent", s.Config.ProxyUserAgent)
		r.Header.Set("origin", s.RP.OriginURL.Host)
		r.Header.Set("referer", strings.Replace(r.Header.Get("referer"), s.RP.ProxyHost, s.RP.OriginURL.String(), 1))

		s.RP.Proxy.ServeHTTP(w, r)
	}
}

func (s *Service) TransformResponse(resp *http.Response) (err error) {
	contentEncoding := resp.Header.Get("Content-Encoding")
	reader, err := enc.Reader(contentEncoding, resp.Body)
	if err != nil {
		return err
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	body = replaceAppSubdomain(body, s.RP.ProxyHost)
	body = replaceStaticSubdomain(body, s.RP.ProxyHost)

	buf := new(bytes.Buffer)
	closeWriter, err := enc.Writer(contentEncoding, body, buf)
	if err != nil {
		closeWriter()
		return err
	}
	closeWriter()

	resp.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
	resp.ContentLength = int64(len(buf.Bytes()))
	resp.Header.Set("Content-Length", strconv.Itoa(len(buf.Bytes())))

	return nil
}

func replaceAppSubdomain(body []byte, proxyHost string) []byte {
	re, _ := regexp.Compile(`https:\/\/app\.ahrefs\.com`)
	return re.ReplaceAll(body, []byte(fmt.Sprintf("%s%s", proxyHost, alias.AppDomainAlias)))
}

func replaceStaticSubdomain(htmlBody []byte, proxyHost string) []byte {
	re, _ := regexp.Compile(`https:\/\/static\.ahrefs\.com`)
	return re.ReplaceAll(htmlBody, []byte(fmt.Sprintf("%s%s", proxyHost, alias.StaticDomainAlias)))
}
