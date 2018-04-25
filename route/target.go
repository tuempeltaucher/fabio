package route

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/fabiolb/fabio/metrics"
)

type Target struct {
	// Service is the name of the service the targetURL points to
	Service string

	// Tags are the list of tags for this target
	Tags []string

	// Opts is the raw options for the target.
	Opts map[string]string

	// StripPath will be removed from the front of the outgoing
	// request path
	StripPath string

	// TLSSkipVerify disables certificate validation for upstream
	// TLS connections.
	TLSSkipVerify bool

	// Host signifies what the proxy will set the Host header to.
	// The proxy does not modify the Host header by default.
	// When Host is set to 'dst' the proxy will use the host name
	// of the target host for the outgoing request.
	Host string

	// URL is the endpoint the service instance listens on
	URL *url.URL

	// RedirectCode is the HTTP status code used for redirects.
	// When set to a value > 0 the client is redirected to the target url.
	RedirectCode int

	// RedirectURL is the redirect target based on the request.
	// This is cached here to prevent multiple generations per request.
	RedirectURL *url.URL

	// FixedWeight is the weight assigned to this target.
	// If the value is 0 the targets weight is dynamic.
	FixedWeight float64

	// Weight is the actual weight for this service in percent.
	Weight float64

	// Timer measures throughput and latency of this target
	Timer metrics.Timer

	// TimerName is the name of the timer in the metrics registry
	TimerName string

	// accessRules is map of access information for the target.
	accessRules map[string][]interface{}
}

func (t *Target) BuildRedirectURL(requestURL *url.URL) {
	t.RedirectURL = &url.URL{
		Scheme:   t.URL.Scheme,
		Host:     t.URL.Host,
		Path:     t.URL.Path,
		RawPath:  t.URL.RawPath,
		RawQuery: t.URL.RawQuery,
	}
	// if the target has no rawpath, but the request does, we have to set the
	// redirectURL's rawpath manually
	if t.RedirectURL.RawPath == "" && requestURL.RawPath != "" {
		t.RedirectURL.RawPath = t.RedirectURL.Path
	}
	// treat case of $path not separated with a / from host
	if strings.HasSuffix(t.RedirectURL.Host, "$path") {
		t.RedirectURL.Host = t.RedirectURL.Host[:len(t.RedirectURL.Host)-len("$path")]
		t.RedirectURL.Path = "$path"
	}

	// remove / before $path in redirect url
	if strings.Contains(t.RedirectURL.Path, "/$path") {
		t.RedirectURL.Path = strings.Replace(t.RedirectURL.Path, "/$path", "$path", 1)
	}
	if strings.Contains(t.RedirectURL.RawPath, "/$path") {
		t.RedirectURL.RawPath = strings.Replace(t.RedirectURL.RawPath, "/$path", "$path", 1)
	}
	// insert passed request path into redirect path, strip decoded strippath, set query
	if strings.Contains(t.RedirectURL.Path, "$path") {
		t.RedirectURL.Path = strings.Replace(t.RedirectURL.Path, "$path", requestURL.Path, 1)
		if t.StripPath != "" {
			// parse stripPath for not raw path
			parsedStripPath, _ := url.Parse(t.StripPath)
			decodedStripPath := parsedStripPath.Path
			if strings.HasPrefix(t.RedirectURL.Path, decodedStripPath) {
				t.RedirectURL.Path = t.RedirectURL.Path[len(decodedStripPath):]
			}
		}
		if t.RedirectURL.RawQuery == "" && requestURL.RawQuery != "" {
			t.RedirectURL.RawQuery = requestURL.RawQuery
		}
	}
	// insert passed request path into redirect rawpath
	if strings.Contains(t.RedirectURL.RawPath, "$path") {
		var replaceRawPath string
		if requestURL.RawPath == "" {
			replaceRawPath = requestURL.Path
		} else {
			replaceRawPath = requestURL.RawPath
		}
		t.RedirectURL.RawPath = strings.Replace(t.RedirectURL.RawPath, "$path", replaceRawPath, 1)
		if t.StripPath != "" && strings.HasPrefix(t.RedirectURL.RawPath, t.StripPath) {
			t.RedirectURL.RawPath = t.RedirectURL.RawPath[len(t.StripPath):]
		}
	}
	if t.RedirectURL.Path == "" {
		t.RedirectURL.Path = "/"
	}
	if t.URL.Host == "monkey.com" {
		fmt.Printf("StripPath: %s\n", t.StripPath)
		fmt.Printf("request - host: %s, path: %s, rawpath: %s\n", requestURL.Host, requestURL.Path, requestURL.RawPath)
		fmt.Printf("target - host: %s, path: %s, rawpath: %s\n", t.URL.Host, t.URL.Path, t.URL.RawPath)
		fmt.Printf("constructed - host: %s, path: %s, rawpath: %s\n", t.RedirectURL.Host, t.RedirectURL.Path, t.RedirectURL.RawPath)
		fmt.Printf("redirecturl: %s\n", t.RedirectURL.String())
		fmt.Println("###")
	}
}
