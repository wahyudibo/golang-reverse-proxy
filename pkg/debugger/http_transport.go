package debugger

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type DebugTransport struct{}

func (DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return http.DefaultTransport.RoundTrip(req)
}
