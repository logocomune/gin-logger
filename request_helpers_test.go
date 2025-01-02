package slogger

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func buildHeader(pairs ...[2]string) http.Header {
	h := http.Header{}
	for _, pair := range pairs {
		h.Add(pair[0], pair[1])
	}
	return h
}

func TestGetCFCountry(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected string
	}{
		{
			name:     "XCFIPCountryPresent",
			headers:  buildHeader([2]string{"X-cf-ipcountry", "US"}),
			expected: "US",
		},
		{
			name:     "CFIPCountryPresent",
			headers:  buildHeader([2]string{"cf-ipcountry", "UK"}),
			expected: "UK",
		},
		{
			name:     "BothHeadersPresent",
			headers:  buildHeader([2]string{"x-cf-ipcountry", "CA"}, [2]string{"cf-ipcountry", "FR"}),
			expected: "CA",
		},
		{
			name:     "BothHeadersMissing",
			headers:  http.Header{},
			expected: "",
		},
		{
			name:     "EmptyHeaderValues",
			headers:  buildHeader([2]string{"x-cf-ipcountry", ""}, [2]string{"cf-ipcountry", ""}),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCFCountry(tt.headers)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPathAggregator(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
		aggregator func(path string, statusCode int) string
		expected   string
	}{
		{
			name:       "NilAggregator",
			path:       "/example/path",
			statusCode: 200,
			aggregator: nil,
			expected:   "",
		},
		{
			name:       "SimpleAggregator",
			path:       "/example/path",
			statusCode: 200,
			aggregator: func(path string, statusCode int) string {
				return path + " - " + strconv.Itoa(statusCode)
			},
			expected: "/example/path - 200",
		},
		{
			name:       "AggregatorWithPathOnly",
			path:       "/test",
			statusCode: 404,
			aggregator: func(path string, statusCode int) string {
				return path
			},
			expected: "/test",
		},
		{
			name:       "AggregatorWithStatusOnly",
			path:       "/another/test",
			statusCode: 500,
			aggregator: func(path string, statusCode int) string {
				return strconv.Itoa(statusCode)
			},
			expected: "500",
		},
		{
			name:       "EmptyPath",
			path:       "",
			statusCode: 301,
			aggregator: func(path string, statusCode int) string {
				return "Redirect - " + strconv.Itoa(statusCode)
			},
			expected: "Redirect - 301",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathAggregator(tt.path, tt.statusCode, tt.aggregator)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name        string
		headers     http.Header
		ginClientIp string
		expected    string
	}{

		{
			name:        "XCFConnectingIPPresent",
			headers:     buildHeader([2]string{"X-Cf-Connecting-Ip", "203.0.113.45"}),
			ginClientIp: "",
			expected:    "203.0.113.45",
		},
		{
			name:        "CFConnectingIPPresent",
			headers:     buildHeader([2]string{"Cf-Connecting-Ip", "198.51.100.21"}),
			ginClientIp: "",
			expected:    "198.51.100.21",
		},
		{
			name: "XCFAndCFConnectingIPPresent",
			headers: buildHeader(
				[2]string{"X-Cf-Connecting-Ip", "203.0.113.45"},
				[2]string{"Cf-Connecting-Ip", "198.51.100.21"},
			),
			ginClientIp: "",
			expected:    "203.0.113.45",
		},
		{
			name:        "OnlyGinClientIP",
			headers:     http.Header{},
			ginClientIp: "192.0.2.33",
			expected:    "192.0.2.33",
		},
		{
			name:        "GinClientIPWithEmptyHeaders",
			headers:     buildHeader([2]string{"X-Cf-Connecting-Ip", ""}),
			ginClientIp: "192.0.2.33",
			expected:    "192.0.2.33",
		},
		{
			name:        "NoHeadersOrGinClientIP",
			headers:     http.Header{},
			ginClientIp: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getClientIp(tt.ginClientIp, tt.headers)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetReferer(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected string
	}{
		{
			name:     "RefererHeaderPresent",
			headers:  buildHeader([2]string{"Referer", "https://example.com"}),
			expected: "https://example.com",
		},
		{
			name:     "RefererHeaderEmpty",
			headers:  buildHeader([2]string{"Referer", ""}),
			expected: "",
		},
		{
			name:     "NoRefererHeader",
			headers:  http.Header{},
			expected: "",
		},
		{
			name:     "MultipleHeadersWithReferer",
			headers:  buildHeader([2]string{"X-Some-Other-Header", "value"}, [2]string{"Referer", "https://test.com"}),
			expected: "https://test.com",
		},
		{
			name:     "CaseInsensitiveRefererHeader",
			headers:  buildHeader([2]string{"referer", "https://case-insensitive.com"}),
			expected: "https://case-insensitive.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getReferer(tt.headers)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetUserAgent(t *testing.T) {
	tests := []struct {
		name         string
		ginUserAgent string
		headers      http.Header
		expected     string
	}{
		{
			name:         "UserAgentHeaderOnly",
			headers:      buildHeader([2]string{"User-Agent", "Mozilla/5.0"}),
			ginUserAgent: "",
			expected:     "Mozilla/5.0",
		},
		{
			name:         "GinUserAgentOnly",
			headers:      http.Header{},
			ginUserAgent: "CustomUA/1.0",
			expected:     "CustomUA/1.0",
		},
		{
			name: "BothUserAgentsPresent",
			headers: buildHeader(
				[2]string{"x-user-agent", "Mozilla/5.0"},
			),
			ginUserAgent: "CustomUA/1.0",
			expected:     "Mozilla/5.0",
		},
		{
			name:         "EmptyUserAgentHeader",
			headers:      buildHeader([2]string{"user-agent", ""}),
			ginUserAgent: "FallbackUA/2.0",
			expected:     "FallbackUA/2.0",
		},
		{
			name:         "NoUserAgentOrGinUserAgent",
			headers:      http.Header{},
			ginUserAgent: "",
			expected:     "",
		},
		{
			name:         "CaseInsensitiveUserAgentHeader",
			headers:      buildHeader([2]string{"USER-AGENT", "Mozilla/5.0 (Linux)"}),
			ginUserAgent: "",
			expected:     "Mozilla/5.0 (Linux)",
		},
		{
			name: "MultipleHeadersWithUserAgent",
			headers: buildHeader(
				[2]string{"X-Some-Other-Header", "value"},
				[2]string{"user-agent", "Mozilla/5.0 Safari"},
			),
			ginUserAgent: "AnotherUA/3.0",
			expected:     "Mozilla/5.0 Safari",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUserAgent(tt.headers)
			if tt.ginUserAgent != "" && result == "" {
				result = tt.ginUserAgent
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetPathAndQuery(t *testing.T) {
	tests := []struct {
		name          string
		ginPath       string
		expectedPath  string
		expectedQuery string
	}{
		{
			name:          "PathWithQuery",
			ginPath:       "/example/path?key=value",
			expectedPath:  "/example/path",
			expectedQuery: "key=value",
		},
		{
			name:          "PathWithoutQuery",
			ginPath:       "/example/path",
			expectedPath:  "/example/path",
			expectedQuery: "",
		},
		{
			name:          "OnlyQuery",
			ginPath:       "/?key=value",
			expectedPath:  "/",
			expectedQuery: "key=value",
		},
		{
			name:          "EmptyPathAndQuery",
			ginPath:       "",
			expectedPath:  "",
			expectedQuery: "",
		},
		{
			name:          "ComplexQuery",
			ginPath:       "/complex/path?key1=value1&key2=value2",
			expectedPath:  "/complex/path",
			expectedQuery: "key1=value1&key2=value2",
		},
		{
			name:          "PathWithFragmentAndQuery",
			ginPath:       "/path#fragment?key=value",
			expectedPath:  "/path#fragment",
			expectedQuery: "key=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, query := getPathAndQuery(tt.ginPath)
			if path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, path)
			}
			if query != tt.expectedQuery {
				t.Errorf("Expected query %s, got %s", tt.expectedQuery, query)
			}
		})
	}
}
func TestGetHeaderValue(t *testing.T) {
	tests := []struct {
		name     string
		header   http.Header
		keys     []string
		expected string
		found    bool
	}{
		{
			name:     "SingleKeyFound",
			header:   buildHeader([2]string{"key1", "value1"}),
			keys:     []string{"key1"},
			expected: "value1",
			found:    true,
		},
		{
			name:     "MultipleKeysFirstFound",
			header:   buildHeader([2]string{"key1", "value1"}, [2]string{"key2", "value2"}),
			keys:     []string{"key1", "key2"},
			expected: "value1",
			found:    true,
		},
		{
			name:     "MultipleKeysSecondFound",
			header:   buildHeader([2]string{"key2", "value2"}),
			keys:     []string{"key1", "key2"},
			expected: "value2",
			found:    true,
		},
		{
			name:     "NoKeysFound",
			header:   buildHeader([2]string{"key3", "value3"}),
			keys:     []string{"key1", "key2"},
			expected: "",
			found:    false,
		},
		{
			name:     "EmptyKeys",
			header:   buildHeader([2]string{"key1", "value1"}),
			keys:     []string{},
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found := getHeaderValue(tt.header, tt.keys)
			if value != tt.expected || found != tt.found {
				t.Errorf("Expected value: %s, found: %v; got value: %s, found: %v", tt.expected, tt.found, value, found)
			}
		})
	}
}

func TestGetPathFromContext(t *testing.T) {
	tests := []struct {
		name     string
		request  *http.Request
		expected string
	}{
		{
			name:     "ValidURLPath",
			request:  &http.Request{URL: &url.URL{Path: "/test/path"}},
			expected: "/test/path",
		},
		{
			name:     "NilRequest",
			request:  nil,
			expected: "",
		},
		{
			name:     "NilURL",
			request:  &http.Request{URL: nil},
			expected: "",
		},
		{
			name:     "EmptyPath",
			request:  &http.Request{URL: &url.URL{Path: ""}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPathFromContext(tt.request)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetClientIPFromHeaders(t *testing.T) {
	tests := []struct {
		name        string
		request     *http.Request
		headerNames []string
		expected    string
	}{
		{
			name:        "SingleValidIPHeader",
			request:     &http.Request{Header: buildHeader([2]string{"X-Forwarded-For", "203.0.113.1"})},
			headerNames: []string{"X-Forwarded-For"},
			expected:    "203.0.113.1",
		},
		{
			name: "MultipleIPsInHeaderSingleValid",
			request: &http.Request{Header: buildHeader(
				[2]string{"X-Forwarded-For", "invalidIp, 203.0.113.5"},
			)},
			headerNames: []string{"X-Forwarded-For"},
			expected:    "203.0.113.5",
		},
		{
			name:        "InvalidIPHeaders",
			request:     &http.Request{Header: buildHeader([2]string{"X-Forwarded-For", "invalidIp"})},
			headerNames: []string{"X-Forwarded-For"},
			expected:    "",
		},
		{
			name: "FallbackToRemoteAddr",
			request: &http.Request{
				Header:     http.Header{},
				RemoteAddr: "192.0.2.4:12345",
			},
			headerNames: []string{"X-Forwarded-For"},
			expected:    "192.0.2.4",
		},
		{
			name: "MalformedRemoteAddr",
			request: &http.Request{
				Header:     http.Header{},
				RemoteAddr: "malformedAddr",
			},
			headerNames: []string{"X-Forwarded-For"},
			expected:    "",
		},
		{
			name:        "EmptyHeadersAndRemoteAddr",
			request:     &http.Request{Header: http.Header{}},
			headerNames: []string{"X-Forwarded-For"},
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetClientIPFromHeaders(tt.request, tt.headerNames)
			if result != tt.expected {
				t.Errorf("Expected IP %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{name: "ValidIPv4", ip: "192.0.2.1", expected: true},
		{name: "ValidIPv6", ip: "2001:db8::1", expected: true},
		{name: "InvalidIP", ip: "invalidIp", expected: false},
		{name: "EmptyIP", ip: "", expected: false},
		{name: "MalformedIPv4", ip: "256.256.256.256", expected: false},
		{name: "IncompleteIPv6", ip: "2001:db8:", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidIP(tt.ip)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
func TestExtractRequestDetails(t *testing.T) {
	tests := []struct {
		name            string
		param           gin.LogFormatterParams
		pAggregatorFunc func(string, int) string
		expected        struct {
			pathAggregated string
			path           string
			query          string
			ipCountry      string
			clientIp       string
			referer        string
			userAgent      string
		}
	}{
		{
			name: "InvalidPathAndStatusCode",
			param: gin.LogFormatterParams{
				Path:       "/invalid??",
				ClientIP:   "",
				Request:    &http.Request{Header: http.Header{}},
				Keys:       nil,
				StatusCode: 0,
			},
			pAggregatorFunc: func(path string, statusCode int) string {
				return "invalidRequest"
			},
			expected: struct {
				pathAggregated string
				path           string
				query          string
				ipCountry      string
				clientIp       string
				referer        string
				userAgent      string
			}{
				pathAggregated: "invalidRequest",
				path:           "/invalid",
				query:          "?",
				ipCountry:      "",
				clientIp:       "",
				referer:        "",
				userAgent:      "",
			},
		},
		{
			name: "ConflictingCaseHeaders",
			param: gin.LogFormatterParams{
				Path:       "/conflicting",
				ClientIP:   "",
				Request:    &http.Request{Header: buildHeader([2]string{"User-Agent", "UA1"}, [2]string{"user-agent", "UA2"})},
				Keys:       nil,
				StatusCode: 200,
			},
			pAggregatorFunc: func(path string, statusCode int) string {
				return path + "-ok"
			},
			expected: struct {
				pathAggregated string
				path           string
				query          string
				ipCountry      string
				clientIp       string
				referer        string
				userAgent      string
			}{
				pathAggregated: "/conflicting-ok",
				path:           "/conflicting",
				query:          "",
				ipCountry:      "",
				clientIp:       "",
				referer:        "",
				userAgent:      "UA1",
			},
		},
		{
			name: "LargeKeysInParam",
			param: gin.LogFormatterParams{
				Path:       "/largedata?data=large",
				ClientIP:   "198.20.33.4",
				Request:    &http.Request{Header: buildHeader([2]string{"Referer", "https://example.com"}, [2]string{"cf-ipcountry", "US"}, [2]string{"extra", strings.Repeat("data", 1000)})},
				StatusCode: 201,
			},
			pAggregatorFunc: func(path string, statusCode int) string {
				return "largedata_201"
			},
			expected: struct {
				pathAggregated string
				path           string
				query          string
				ipCountry      string
				clientIp       string
				referer        string
				userAgent      string
			}{
				pathAggregated: "largedata_201",
				path:           "/largedata",
				query:          "data=large",
				ipCountry:      "US",
				clientIp:       "198.20.33.4",
				referer:        "https://example.com",
				userAgent:      "",
			},
		},
		{
			name: "ComplexPathAggregator",
			param: gin.LogFormatterParams{
				Path:       "/example",
				ClientIP:   "",
				Request:    &http.Request{Header: http.Header{}},
				Keys:       nil,
				StatusCode: 400,
			},
			pAggregatorFunc: func(path string, statusCode int) string {
				return strings.ReplaceAll(path, "/", "_") + strconv.Itoa(statusCode)
			},
			expected: struct {
				pathAggregated string
				path           string
				query          string
				ipCountry      string
				clientIp       string
				referer        string
				userAgent      string
			}{
				pathAggregated: "_example400",
				path:           "/example",
				query:          "",
				ipCountry:      "",
				clientIp:       "",
				referer:        "",
				userAgent:      "",
			},
		},
		{
			name: "AllHeadersPresent",
			param: gin.LogFormatterParams{
				Path:     "/full/details?info=value",
				ClientIP: "203.0.113.7",
				Request: &http.Request{Header: buildHeader(
					[2]string{"Referer", "https://ref.com"},
					[2]string{"User-Agent", "Test-UA"},
					[2]string{"X-Cf-IpCountry", "IN"},
				)},
				Keys:       nil,
				StatusCode: 200,
			},
			pAggregatorFunc: func(path string, statusCode int) string {
				return "allHeaders-" + strconv.Itoa(statusCode)
			},
			expected: struct {
				pathAggregated string
				path           string
				query          string
				ipCountry      string
				clientIp       string
				referer        string
				userAgent      string
			}{
				pathAggregated: "allHeaders-200",
				path:           "/full/details",
				query:          "info=value",
				ipCountry:      "IN",
				clientIp:       "203.0.113.7",
				referer:        "https://ref.com",
				userAgent:      "Test-UA",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathAggregated := tt.pAggregatorFunc(tt.param.Path, tt.param.StatusCode)
			path, query := getPathAndQuery(tt.param.Path)
			ipCountry := getCFCountry(tt.param.Request.Header)
			clientIp := getClientIp(tt.param.ClientIP, tt.param.Request.Header)
			referer := getReferer(tt.param.Request.Header)
			userAgent := getUserAgent(tt.param.Request.Header)

			if pathAggregated != tt.expected.pathAggregated {
				t.Errorf("Expected pathAggregated %s, got %s", tt.expected.pathAggregated, pathAggregated)
			}
			if path != tt.expected.path {
				t.Errorf("Expected path %s, got %s", tt.expected.path, path)
			}
			if query != tt.expected.query {
				t.Errorf("Expected query %s, got %s", tt.expected.query, query)
			}
			if ipCountry != tt.expected.ipCountry {
				t.Errorf("Expected ipCountry %s, got %s", tt.expected.ipCountry, ipCountry)
			}
			if clientIp != tt.expected.clientIp {
				t.Errorf("Expected clientIp %s, got %s", tt.expected.clientIp, clientIp)
			}
			if referer != tt.expected.referer {
				t.Errorf("Expected referer %s, got %s", tt.expected.referer, referer)
			}
			if userAgent != tt.expected.userAgent {
				t.Errorf("Expected userAgent %s, got %s", tt.expected.userAgent, userAgent)
			}
		})
	}
}
