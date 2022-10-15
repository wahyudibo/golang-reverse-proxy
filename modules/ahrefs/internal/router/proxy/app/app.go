package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/chromedp/cdproto/network"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog/log"
	redisClient "github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/adapter/cache/redis"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/constant"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/repository"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/debugger"
	enc "github.com/wahyudibo/golang-reverse-proxy/pkg/encoding"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/headless"
	"github.com/wahyudibo/golang-reverse-proxy/pkg/proxy"
)

type Service struct {
	Context    context.Context
	Config     *config.Config
	Repository repository.Repository
	Cache      *redis.Client
	RP         *proxy.ReverseProxy
}

func New(ctx context.Context, cfg *config.Config, repo repository.Repository, cache *redis.Client) (*Service, error) {
	url, err := url.Parse(constant.AppDomain)
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

	s := &Service{Context: ctx, Config: cfg, Repository: repo, Cache: cache, RP: rp}

	rp.Proxy.ModifyResponse = s.TransformResponse

	return s, nil
}

// Handler handles the http request using proxy.
func (s *Service) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// check if cookies already exists in cache
		cookiesKey := fmt.Sprintf("%s:cookies", redisClient.KeyPrefix)
		cookiesJSON, err := s.Cache.Get(s.Context, cookiesKey).Result()
		if err != nil {
			log.Error().Err(err).Msg("failed to get cookies from cache")

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return
		}

		networkCookies := make([]*network.Cookie, 0)
		err = json.Unmarshal([]byte(cookiesJSON), &networkCookies)
		if err != nil {
			log.Error().Err(err).Msg("failed to unmarshal cookies from cache")

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return
		}

		for key := range r.Header {
			r.Header.Set(key, r.Header.Get(key))
		}

		r.Header.Set("user-agent", s.Config.ProxyUserAgent)
		r.Header.Set("origin", s.RP.OriginURL.Host)

		if r.Header.Get("referer") == "" {
			r.Header.Set("referer", s.RP.OriginURL.String())
		} else {
			r.Header.Set("referer", strings.Replace(r.Header.Get("referer"), s.RP.ProxyHost, s.RP.OriginURL.String(), 1))
		}

		for _, c := range networkCookies {
			cookie := headless.TransformNetworkCookieToHTTPCookie(c)
			r.AddCookie(cookie)
		}

		s.RP.Proxy.ServeHTTP(w, r)
	}
}

func (s *Service) TransformResponse(resp *http.Response) (err error) {
	if resp.Header.Get("Location") != "" {
		resp.Header.Set("Location", strings.Replace(
			resp.Header.Get("Location"),
			fmt.Sprintf("//%s", s.RP.OriginURL.Host),
			s.RP.ProxyHost,
			1,
		))
	}

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
	return re.ReplaceAll(body, []byte(fmt.Sprintf("%s%s", proxyHost, constant.AppDomainAlias)))
}

func replaceStaticSubdomain(htmlBody []byte, proxyHost string) []byte {
	re, _ := regexp.Compile(`https:\/\/static\.ahrefs\.com`)
	return re.ReplaceAll(htmlBody, []byte(fmt.Sprintf("%s%s", proxyHost, constant.StaticDomainAlias)))
}
