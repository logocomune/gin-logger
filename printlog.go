package slogger

import (
	"log/slog"
	"time"
)

type logEntry struct {
	created       time.Time
	remoteIp      string
	ua            string
	method        string
	aggregatePath string

	statusCode int
	count      int
	proto      string

	isBotDetectorEnabled bool
	isBot                int

	isAggregate bool
	aggregateDetails
	realtimeDetails
	extraFields map[string]extraFields
	ip          string
}

type extraFields struct {
	value string
	found bool
}

type aggregateDetails struct {
	lastMod          time.Time
	sumLatency       time.Duration
	maxLatency       time.Duration
	minLatency       time.Duration
	sumSizeRespoBody int
	int
}
type realtimeDetails struct {
	errorMessage     string
	queryString      string
	path             string
	headers          map[string]string
	referer          string
	latency          time.Duration
	responseBodySize int
}

// printLog processes and emits structured logging for HTTP requests, including metadata, request details, and metrics.
func printLog(msg string, v logEntry, c *conf) {
	args := []any{
		slog.String("created", v.created.Format(time.RFC3339)),
		slog.String("ip", v.ip),
		slog.String("remoteIp", v.remoteIp),
		slog.String("ua", v.ua),
		slog.String("method", v.method),
		slog.String("proto", v.proto),
		slog.Int("statusCode", v.statusCode),
		slog.Int("counter", v.count),
	}
	if v.referer != "" {
		args = append(args, slog.String("referer", v.referer))
	}

	if v.isBotDetectorEnabled {
		args = append(args, slog.Int("isBot", v.isBot))
	}
	if v.aggregatePath != "" {
		args = append(args, slog.String("aggregatePath", v.aggregatePath))
	}
	if v.isAggregate {
		meanSizeRespBody := 0.0
		if v.sumSizeRespoBody > 0 && v.count != 0 {
			meanSizeRespBody = float64(v.sumSizeRespoBody) / float64(v.count)
		}

		args = append(args,
			slog.Duration("meanLatency", v.sumLatency/time.Duration(v.count)),
			slog.Duration("minLatency", v.minLatency),
			slog.Duration("maxLatency", v.maxLatency),
			slog.Float64("meanSizeRespBody", meanSizeRespBody),
			slog.Int("sumSizeRespBody", v.sumSizeRespoBody))

	} else {
		//Only realtime
		if v.path != "" {
			args = append(args, slog.String("path", v.path))
		}
		if v.errorMessage != "" {
			args = append(args, slog.String("errorMessage", v.errorMessage))
		}
		if c.logQueryString && v.queryString != "" {
			args = append(args, slog.String("queryString", v.queryString))
		}
		if c.logHeaders && v.headers != nil && len(v.headers) > 0 {
			args = append(args, slog.Any("fullHeaders", v.headers))
		}
		args = append(args, slog.Duration("latency", v.latency))

		args = append(args, slog.Int("responseSize", v.responseBodySize))

	}
	if c.logHeadersWithName != nil && len(c.logHeadersWithName) > 0 {
		for key, value := range v.extraFields {
			if value.found {
				args = append(args, slog.String(key, value.value))
			}
		}
	}
	if c.staticLogEntries != nil && len(c.staticLogEntries) > 0 {
		for key, value := range c.staticLogEntries {
			args = append(args, slog.String(key, value))
		}
	}

	c.loggingHandler.Info(
		msg,
		args...,
	)
}
