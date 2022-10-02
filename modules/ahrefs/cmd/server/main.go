package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	// "github.com/PuerkitoBio/goquery"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

const proxyHost = "http://localhost:8080"

type DebugTransport struct{}

func (DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return http.DefaultTransport.RoundTrip(req)
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy, host string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for key := range r.Header {
			r.Header.Set(key, r.Header.Get(key))
		}
		r.Header.Set("referer", strings.Replace(r.Header.Get("referer"), proxyHost, host, 1))

		proxy.ServeHTTP(w, r)
	}
}

func TransformResponse(resp *http.Response) (err error) {
	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// doc, err := goquery.NewDocumentFromReader(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// html, err := goquery.OuterHtml(doc.First())
	// if err != nil {
	// 	log.Fatalf("Bad html %v", err)
	// 	return err
	// }
	// fmt.Printf("Body %v", html)

	// resp.Body = html
	// resp.ContentLength = int64(len(b))
	// resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

	// return nil

	// b, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// err = resp.Body.Close()
	// if err != nil {
	// 	return err
	// }

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
	default:
		reader = resp.Body
	}
	defer reader.Close()

	// reader, err := gzip.NewReader(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// defer reader.Close()
	b, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	re, _ := regexp.Compile(`https:\/\/static\.ahrefs\.com`)
	b = re.ReplaceAll(b, []byte(fmt.Sprintf("%s/stah", proxyHost)))

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(b); err != nil {
		gz.Close()
		panic(err)
	}
	gz.Close()

	// fmt.Printf("HTML: %s\n", string(b))

	// resp.Body = io.NopCloser(bytes.NewReader(b))
	resp.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

	return nil
}

func main() {
	url, err := url.Parse("https://ahrefs.com")
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.Transport = DebugTransport{}
	d := proxy.Director
	proxy.Director = func(req *http.Request) {
		d(req)
		req.Host = url.Host
		req.URL.Host = url.Host
		req.URL.Scheme = "https"
	}
	proxy.ModifyResponse = TransformResponse

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))

	r.HandleFunc("/*", ProxyRequestHandler(proxy, url.String()))

	log.Fatal(http.ListenAndServe(":8080", r))
}
