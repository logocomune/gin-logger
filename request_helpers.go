package slogger

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
)

// extractRequestDetails extracts details from an HTTP request such as path, query, IP country, client IP, referer, and user agent.
// It uses helper functions to parse and derive specific header and request information.
// param: LogFormatterParams object containing HTTP request and response data.
// pAggregatorFunc: Function for aggregating paths and status codes for metrics or logging purposes.
// Returns aggregated path, original path, query string, country of the client IP, client IP address, referer URL, and user agent.
func extractRequestDetails(param gin.LogFormatterParams, pAggregatorFunc func(path string, statusCode int) string) (pathAggregated, path, query, ipCountry, clientIp, referer, userAgent string) {
	header := param.Request.Header

	pathAggregated = pathAggregator(param.Path, param.StatusCode, pAggregatorFunc)

	// Get path and query
	path, query = getPathAndQuery(param.Path)

	// Get IP country
	ipCountry = getCFCountry(header)

	// Get client IP
	clientIp = getClientIp(param.ClientIP, header)

	// Get referer
	referer = getReferer(header)

	// Get user agent
	userAgent = getUserAgent(header)

	return pathAggregated, path, query, ipCountry, clientIp, referer, userAgent
}

// pathAggregator applies a given aggregator function to a path and statusCode and returns the result as a string.
// If the aggregator function is nil, it returns an empty string.
func pathAggregator(path string, statusCode int, aggregator func(path string, statusCode int) string) string {
	if aggregator == nil {
		return ""
	}
	return aggregator(path, statusCode)
}

// getCFCountry extracts the Cloudflare country code from the HTTP headers.
// It prioritizes the "x-cf-ipcountry" header over the "cf-ipcountry" header.
// Returns an empty string if neither header is present.
func getCFCountry(header http.Header) string {
	if ipCountry := header.Get("x-cf-ipcountry"); ipCountry != "" {
		return ipCountry
	}
	return header.Get("cf-ipcountry")

}

// getClientIp retrieves the client's IP address from the request headers or falls back to the provided ginClientIp parameter.
func getClientIp(ginClientIp string, header http.Header) string {
	if clientIp := header.Get("x-cf-connecting-ip"); clientIp != "" {
		return clientIp
	}
	if clientIp := header.Get("cf-connecting-ip"); clientIp != "" {
		return clientIp
	}
	return ginClientIp
}

// getReferer extracts the referer URL from the provided HTTP headers, checking "x-referer" first, then defaulting to "referer".
func getReferer(header http.Header) string {
	if referer := header.Get("x-referer"); referer != "" {
		return referer
	}
	return header.Get("referer")
}

// getUserAgent retrieves the user agent string from the HTTP header, prioritizing the "x-user-agent" key over "user-agent".
func getUserAgent(header http.Header) string {
	if userAgent := header.Get("x-user-agent"); userAgent != "" {
		return userAgent
	}
	return header.Get("user-agent")
}

// GetClientIPFromHeaders extracts the client's IP address from the specified headers or the request's remote address as fallback.
func GetClientIPFromHeaders(r *http.Request, headerNames []string) string {
	for _, headerName := range headerNames {

		headerValue := r.Header.Get(headerName)
		if headerValue != "" {

			ips := strings.Split(headerValue, ",")

			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if isValidIP(ip) {
					return ip
				}
			}
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && isValidIP(ip) {
		return ip
	}

	return ""
}

// isValidIP determines whether the given string is a valid IP address.
// It returns true if the string represents a valid IPv4 or IPv6 address, otherwise false.
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// getPathAndQuery splits a given ginPath into the path and query components, returning them as separate strings.
func getPathAndQuery(ginPath string) (path, query string) {
	pathComp := strings.Split(ginPath, "?")
	path = pathComp[0]
	query = ""
	if len(pathComp) > 0 {
		query = strings.Join(pathComp[1:], "?")
	}
	return path, query
}

// getHeaderValue retrieves the first non-empty value of the specified keys from the given HTTP header.
// Returns the value and a boolean indicating if a non-empty value was found.
func getHeaderValue(header http.Header, keys []string) (string, bool) {
	if len(keys) == 0 {
		return "", false
	}
	for _, key := range keys {
		if val := header.Get(key); val != "" {
			return val, true
		}
	}
	return "", false
}

// getPathFromContext extracts the URL path from an HTTP request's context, returning an empty string if not available.
func getPathFromContext(r *http.Request) string {
	if r != nil && r.URL != nil {
		return r.URL.Path
	}
	return ""
}
