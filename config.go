package slogger

import (
	"log/slog"
	"os"
	"time"
)

// Option defines a function type used to modify or configure a conf instance.
type Option func(c *conf)

// TimeSource represents a function that returns the current time as a time.Time value.
type TimeSource func() time.Time

// BotDetector is an interface used to identify whether a given user agent string corresponds to a bot or not.
type BotDetector interface {
	IsBot(userAgent string) bool
}

// conf represents the configuration options for the application, including logging, bot detection, and path handling.
type conf struct {
	botDetectionService  BotDetector
	logQueryString       bool
	excludedPaths        []string
	logHeaders           bool
	aggregationQueueSize int
	defaultLogMessage    string
	loggingHandler       *slog.Logger
	clientIPHeaders      []string
	logHeadersWithName   map[string][]string
	pathMappingFunction  func(route string, path string, statusCode int) string
	isAggregationEnabled bool
	aggregationInterval  time.Duration
	userAgentHeaders     []string
	staticLogEntries     map[string]string
}

// WithTimeAggregation sets the time duration for aggregation and enables the aggregation feature in the configuration.
func WithTimeAggregation(every time.Duration) Option {
	return func(c *conf) {
		c.isAggregationEnabled = true
		c.aggregationInterval = every
	}
}

// WithAggregation sets the aggregation flag in the configuration to the specified boolean value.
func WithAggregation(b bool) Option {
	return func(c *conf) {
		c.isAggregationEnabled = b
	}
}

// WithLogMessage returns an Option to set a custom log message in the configuration.
func WithLogMessage(msg string) Option {
	return func(c *conf) {
		c.defaultLogMessage = msg
	}
}

// WithBotDetector sets the given BotDetector instance to the configuration.
func WithBotDetector(detector BotDetector) Option {
	return func(c *conf) {
		c.botDetectionService = detector
	}
}

// WithAggregatePath sets a custom function for path aggregation based on router, path, and status code.
func WithAggregatePath(pathFunc func(router, path string, statusCode int) string) Option {
	return func(c *conf) {
		c.pathMappingFunction = pathFunc
	}
}

// WithLogQueryString sets the logQueryString configuration to the provided boolean value.
func WithLogQueryString(log bool) Option {
	return func(c *conf) {
		c.logQueryString = log
	}
}

// WithSkipPaths sets the excludedPaths field in the configuration to exclude specific paths from processing.
func WithSkipPaths(paths []string) Option {
	return func(c *conf) {
		c.excludedPaths = paths
	}
}

// WithLogHeaders configures whether HTTP headers should be logged in the application logs.
func WithLogHeaders(log bool) Option {
	return func(c *conf) {
		c.logHeaders = log
	}
}

// WithQueueSize sets the queue size in the configuration. Accepts an integer `size` as the queue size.
func WithQueueSize(size int) Option {
	return func(c *conf) {
		c.aggregationQueueSize = size
	}
}

// WithLogger sets a custom slog.Logger for configuration and returns an Option to modify the conf instance.
func WithLogger(logger *slog.Logger) Option {
	if logger == nil {
		return func(c *conf) {}
	}
	return func(c *conf) {
		c.loggingHandler = logger
	}
}

// WithIpHeaders sets the list of IP headers to be used for extracting client IP information into the configuration.
func WithIpHeaders(headers []string) Option {
	return func(c *conf) {
		c.clientIPHeaders = headers
	}
}

// WithUaHeaders sets the list of user-agent headers to include during configuration and assigns it to the conf instance.
func WithUaHeaders(headers []string) Option {
	return func(c *conf) {
		c.userAgentHeaders = headers
	}
}

// WithHeaderToLogs configures the headers to be logged by associating them with specific names in the logHeadersWithName map.
func WithHeaderToLogs(headerToLogs map[string][]string) Option {
	return func(c *conf) {
		c.logHeadersWithName = headerToLogs
	}
}

// WithPathAggregator sets a custom path aggregation function to modify how route, path, and status codes are aggregated.
func WithPathAggregator(pathAggregator func(route string, path string, statusCode int) string) Option {
	return func(c *conf) {
		c.pathMappingFunction = pathAggregator
	}
}

// WithStaticLogEntries sets static log entries to be included in all generated log messages. It modifies the configuration.
func WithStaticLogEntries(entries map[string]string) Option {
	return func(c *conf) {
		c.staticLogEntries = entries
	}
}

// configure sets up the configuration for the application logger with the provided name, version, and optional settings.
func configure(opts ...Option) *conf {
	c := &conf{
		loggingHandler:      slog.New(slog.NewTextHandler(os.Stdout, nil)),
		defaultLogMessage:   "logger v1",
		botDetectionService: nil,
		pathMappingFunction: func(route string, path string, statusCode int) string {
			if route == "" {
				if statusCode >= 400 && statusCode < 500 {
					return "error_4xx"
				} else if statusCode >= 500 {
					return "error_5xx"
				}
				return "missing_route"
			}
			return route
		},
		excludedPaths:        []string{},            // eg: []string{"/health", "/api/health", "/metrics", "/api/metrics", "/static"},
		clientIPHeaders:      []string{},            //[]string{"x-CF-Connecting-IP", "X-CF-Connecting-IP", "X-Forwarded-For", "X-Real-IP"},
		userAgentHeaders:     []string{},            //[]string{"x-user-agent", "user-agent"},
		logHeadersWithName:   map[string][]string{}, //map[string][]string{"country": {"x-cf-ipcountry", "cf-ipcountry"},"referer": {"x-referer", "referer"},},
		staticLogEntries:     map[string]string{},
		isAggregationEnabled: false,
		aggregationQueueSize: 100,
		aggregationInterval:  10 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
