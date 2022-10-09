package headless

import (
	"net/http"
	"time"

	"github.com/chromedp/cdproto/network"
)

var networkCookieToHTTPCookieSameSite = map[string]http.SameSite{
	"Strict": http.SameSiteStrictMode,
	"Lax":    http.SameSiteLaxMode,
	"None":   http.SameSiteNoneMode,
}

func TransformNetworkCookieToHTTPCookie(c *network.Cookie) *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Path:     c.Path,
		Domain:   c.Domain,
		Expires:  time.Unix(int64(c.Expires), 0),
		Secure:   c.Secure,
		HttpOnly: c.HTTPOnly,
		SameSite: networkCookieToHTTPCookieSameSite[c.SameSite.String()],
	}
}
