package route

import (
	"net/url"
	"strings"
	"testing"
)

func TestTarget_BuildRedirectURL(t *testing.T) {
	type routeTest struct {
		req  string
		want string
	}
	tests := []struct {
		route string
		tests []routeTest
	}{
		{ // simple absolute redirect
			route: "route add svc / http://bar.com/",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/"},
				{req: "/a/b/c", want: "http://bar.com/"},
				{req: "/?aaa=1", want: "http://bar.com/"},
			},
		},
		{ // simple absolute redirect with encoded character
			route: "route add svc / http://bar.com/%2f/",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/%2f/"},
				{req: "/abc", want: "http://bar.com/%2f/"},
				{req: "/a/b/c", want: "http://bar.com/%2f/"},
				{req: "/?aaa=1", want: "http://bar.com/%2f/"},
			},
		},
		{ // absolute redirect to deep path with query
			route: "route add svc / http://bar.com/a/b/c?foo=bar",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/abc", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/?aaa=1", want: "http://bar.com/a/b/c?foo=bar"},
			},
		},
		{ // simple redirect to corresponding path
			route: "route add svc / http://bar.com/$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/abc"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/abc/?aaa=1"},
			},
		},
		{ // same as above but without / before $path
			route: "route add svc / http://bar.com$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/abc"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/abc/?aaa=1"},
			},
		},
		{ // arbitrary subdir on target with $path at end
			route: "route add svc / http://bar.com/bbb/$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/bbb/"},
				{req: "/abc", want: "http://bar.com/bbb/abc"},
				{req: "/a/b/c", want: "http://bar.com/bbb/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/bbb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/bbb/abc/?aaa=1"},
			},
		},
		{ // same as above but without / before $path
			route: "route add svc / http://bar.com/bbb$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/bbb/"},
				{req: "/abc", want: "http://bar.com/bbb/abc"},
				{req: "/a/b/c", want: "http://bar.com/bbb/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/bbb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/bbb/abc/?aaa=1"},
			},
		},
		{ // subdir with encoded char on target with $path at end
			route: "route add svc / http://bar.com/b%2fb/$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/b%2fb/"},
				{req: "/abc", want: "http://bar.com/b%2fb/abc"},
				{req: "/a/b/c", want: "http://bar.com/b%2fb/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/b%2fb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/b%2fb/abc/?aaa=1"},
			},
		},
		{ // simple redirect to corresponding path with encoded char in path
			route: "route add svc / http://bar.com/$path",
			tests: []routeTest{
				{req: "/%20", want: "http://bar.com/%20"},
				{req: "/a%2fbc", want: "http://bar.com/a%2fbc"},
				{req: "/a/b%22/c", want: "http://bar.com/a/b%22/c"},
				{req: "/%2f/?aaa=1", want: "http://bar.com/%2f/?aaa=1"},
			},
		},
		{ // subdir + path with encoded char on target and encoded char in path
			route: "route add svc / http://monkey.com/b%2fb/$path",
			tests: []routeTest{
				{req: "/%22", want: "http://monkey.com/b%2fb/%22"},
				//{req: "/a%2fbc", want: "http://monkey.com/b%2fb/a%2fbc"},
				//{req: "/a/b%22/c", want: "http://monkey.com/b%2fb/a/b%22/c"},
				//{req: "/%2f/?aaa=1", want: "http://monkey.com/b%2fb/%2f/?aaa=1"},
			},
		},
		{ // strip prefix
			route: "route add svc /stripme http://bar.com/$path opts \"strip=/stripme\"",
			tests: []routeTest{
				{req: "/stripme/", want: "http://bar.com/"},
				{req: "/stripme/abc", want: "http://bar.com/abc"},
				{req: "/stripme/a/b/c", want: "http://bar.com/a/b/c"},
				{req: "/stripme/?aaa=1", want: "http://bar.com/?aaa=1"},
				{req: "/stripme/abc/?aaa=1", want: "http://bar.com/abc/?aaa=1"},
			},
		},
		{ // strip prefix containing encoded char
			route: "route add svc /strip%2fme http://bar.com/$path opts \"strip=/strip%2fme\"",
			tests: []routeTest{
				{req: "/strip%2fme/abc", want: "http://bar.com/abc"},
				{req: "/strip%2fme/ab%22c", want: "http://bar.com/ab%22c"},
				{req: "/strip%2fme/ab%2fc", want: "http://bar.com/ab%2fc"},
			},
		},
	}
	firstRoute := func(tbl Table) *Route {
		for _, routes := range tbl {
			return routes[0]
		}
		return nil
	}
	for _, tt := range tests {
		tbl, _ := NewTable(tt.route)
		route := firstRoute(tbl)
		target := route.Targets[0]
		for _, rt := range tt.tests {
			reqURL, _ := url.Parse("http://foo.com" + rt.req)
			target.BuildRedirectURL(reqURL)
			if strings.Contains(rt.want, "monkey") {
				t.Logf("%#v", reqURL)
				t.Logf("%#v", target.RedirectURL)
			}
			if got := target.RedirectURL.String(); got != rt.want {
				t.Errorf("Got %s, wanted %s", got, rt.want)
			}
		}
	}
}
