package slogger

import (
	"context"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

// Logger is a logging utility that handles real-time and aggregated logging for application events.
// It processes log entries with optional configuration for headers, paths, and bot detection.
type Logger struct {
	queue chan logEntry

	conf *conf
}

// New initializes a new Logger instance with the specified application name, version, and optional configuration options.
func New(ctx context.Context, opts ...Option) *Logger {

	logConf := configure(opts...)

	a := &Logger{

		conf: logConf,
	}

	if logConf.isAggregationEnabled {
		a.queue = make(chan logEntry, logConf.aggregationQueueSize)
		go a.initLoggerAggregator(ctx)
	}
	return a
}

// Middleware returns a Gin middleware handler function for request logging with optional path skipping and isAggregationEnabled.
func (a *Logger) Middleware() gin.HandlerFunc {
	skipPaths := make(map[string]struct{})
	for _, v := range a.conf.excludedPaths {
		skipPaths[v] = struct{}{}
	}

	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		if _, ok := skipPaths[path]; ok {
			c.Next()
			return
		}

		c.Next()
		end := time.Now()

		r := c.Request
		routerPath := c.FullPath()
		statusCode := c.Writer.Status()
		responseBodySize := c.Writer.Size()
		if responseBodySize < 0 {
			responseBodySize = 0
		}
		ip := c.ClientIP()
		var logItem = a.buildLogEntry(start, end, r, ip, statusCode, routerPath, responseBodySize)
		if a.conf.isAggregationEnabled {
			logItem.isAggregate = true
			a.send(logItem)
			return
		}
		printLog("api_logger v1", logItem, a.conf)

	}
}

// buildLogEntry constructs a logEntry object using request details, start-end timestamps, status code, and response attributes.
func (a *Logger) buildLogEntry(start time.Time, end time.Time, r *http.Request, ip string, statusCode int, routerPath string, responseBodySize int) logEntry {
	logConf := a.conf

	query := r.URL.RawQuery
	path := r.URL.Path
	pathAggregated := a.conf.pathMappingFunction(routerPath, path, statusCode)

	remoteAddress, _, _ := net.SplitHostPort(r.RemoteAddr)

	//Override ip
	if logConf.clientIPHeaders != nil && len(logConf.clientIPHeaders) > 0 {

		if ipFromHeader := GetClientIPFromHeaders(r, logConf.clientIPHeaders); ipFromHeader != "" {
			ip = ipFromHeader
		}
	}
	if ip == "" {
		ip = remoteAddress
	}

	latency := end.Sub(start)
	method := r.Method
	userAgent := r.UserAgent()
	if logConf.userAgentHeaders != nil && len(logConf.userAgentHeaders) > 0 {
		userAgent, _ = getHeaderValue(r.Header, logConf.userAgentHeaders)
	}
	referer := r.Referer()
	proto := r.Proto

	statsD := logEntry{
		created:       start,
		ip:            ip,
		remoteIp:      remoteAddress,
		ua:            userAgent,
		method:        method,
		aggregatePath: pathAggregated,
		statusCode:    statusCode,
		count:         1,
		proto:         proto,

		isAggregate: false,
		realtimeDetails: realtimeDetails{
			queryString:      query,
			path:             path,
			headers:          nil,
			referer:          referer,
			latency:          latency,
			responseBodySize: responseBodySize,
		},
		extraFields: make(map[string]extraFields),
	}

	if logConf.botDetectionService != nil {
		isBot := 0
		if logConf.botDetectionService.IsBot(userAgent) {
			isBot = 1
		}
		statsD.isBot = isBot

	}
	if logConf.logHeaders && len(r.Header) > 0 {
		headers := make(map[string]string, len(r.Header))
		for key, val := range r.Header {
			k := strings.ToLower(key)
			if k == "cdn-loop" || k == "user-agent" || k == "x-real-ip" || strings.HasPrefix(k, "cf-") || strings.HasPrefix(k, "x-forwarded-") {
				continue
			}
			headers[k] = strings.Join(val, " | ")
		}
		statsD.headers = headers
	}
	if logConf.logHeadersWithName != nil {
		for k, v := range logConf.logHeadersWithName {

			value, found := getHeaderValue(r.Header, v)

			statsD.extraFields[k] = extraFields{
				found: found,
				value: value,
			}
		}
	}

	return statsD
}

// send attempts to send a logEntry to the loggingHandler's queue channel, emitting a warning if the queue is full.
func (a *Logger) send(l logEntry) {

	select {
	case a.queue <- l:
	default:
		slog.Warn("stats loggingHandler queue channel is full")
	}
}
