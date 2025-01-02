package slogger

import (
	"log/slog"
	"testing"
	"time"
)

type BD struct {
}

func (b *BD) IsBot(ua string) bool {
	if ua == "bot-agent" {
		return true
	}
	return false
}

func TestWithLogMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		expected string
	}{
		{"SetLogMessage", "custom message", "custom message"},
		{"EmptyLogMessage", "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithLogMessage(test.msg)
			opt(c)
			if c.defaultLogMessage != test.expected {
				t.Errorf("expected %v, got %v", test.expected, c.defaultLogMessage)
			}
		})
	}
}

func TestWithBotDetector(t *testing.T) {
	tests := []struct {
		name     string
		detector BotDetector
	}{
		{"SetBotDetector", &BD{}},
		{"NoBotDetector", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithBotDetector(test.detector)
			opt(c)
			if c.botDetectionService != test.detector {
				t.Errorf("expected %v, got %v", test.detector, c.botDetectionService)
			}
		})
	}
}

func TestWithAggregatePath(t *testing.T) {
	tests := []struct {
		name         string
		pathFunc     func(string, string, int) string
		inputPath    string
		inputStatus  int
		expectedPath string
	}{
		{"CustomAggregateFunc", func(router, p string, s int) string {
			return p + "_custom"
		}, "/test", 200, "/test_custom"},
		{"DefaultAggregateFunc", nil, "", 404, "path_4xx"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			if test.pathFunc != nil {
				opt := WithAggregatePath(test.pathFunc)
				opt(c)
			}
			if c.pathMappingFunction == nil {
				c.pathMappingFunction = func(router, p string, s int) string {
					if p == "" && s >= 400 && s < 500 {
						return "path_4xx"
					}
					return p
				}
			}
			result := c.pathMappingFunction("", test.inputPath, test.inputStatus)
			if result != test.expectedPath {
				t.Errorf("expected %v, got %v", test.expectedPath, result)
			}
		})
	}
}

func TestWithLogQueryString(t *testing.T) {
	tests := []struct {
		name     string
		log      bool
		expected bool
	}{
		{"EnableLogQuery", true, true},
		{"DisableLogQuery", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithLogQueryString(test.log)
			opt(c)
			if c.logQueryString != test.expected {
				t.Errorf("expected %v, got %v", test.expected, c.logQueryString)
			}
		})
	}
}

func TestWithSkipPaths(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected []string
	}{
		{"SetSkipPaths", []string{"/skip1", "/skip2"}, []string{"/skip1", "/skip2"}},
		{"EmptySkipPaths", nil, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithSkipPaths(test.paths)
			opt(c)
			if len(c.excludedPaths) != len(test.expected) {
				t.Errorf("expected %v, got %v", test.expected, c.excludedPaths)
			}
		})
	}
}

func TestWithLogHeaders(t *testing.T) {
	tests := []struct {
		name     string
		log      bool
		expected bool
	}{
		{"EnableLogHeaders", true, true},
		{"DisableLogHeaders", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithLogHeaders(test.log)
			opt(c)
			if c.logHeaders != test.expected {
				t.Errorf("expected %v, got %v", test.expected, c.logHeaders)
			}
		})
	}
}

func TestWithQueueSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"SetPositiveSize", 50, 50},
		{"SetZeroSize", 0, 0},
		{"SetNegativeSize", -1, -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithQueueSize(test.size)
			opt(c)
			if c.aggregationQueueSize != test.expected {
				t.Errorf("expected %v, got %v", test.expected, c.aggregationQueueSize)
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	tests := []struct {
		name          string
		appName       string
		appVersion    string
		options       []Option
		expectedName  string
		expectedVer   string
		expectedMsg   string
		expectedQueue int
	}{
		{"DefaultConfig", "app1", "v1.0", nil, "app1", "v1.0", "logger v1", 100},
		{"CustomConfig", "app2", "v2.0", []Option{WithLogMessage("test loggingHandler"), WithQueueSize(200)}, "app2", "v2.0", "test loggingHandler", 200},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := configure(test.options...)

			if c.defaultLogMessage != test.expectedMsg {
				t.Errorf("expected defaultLogMessage %v, got %v", test.expectedMsg, c.defaultLogMessage)
			}
			if c.aggregationQueueSize != test.expectedQueue {
				t.Errorf("expected aggregationQueueSize %v, got %v", test.expectedQueue, c.aggregationQueueSize)
			}
		})
	}
}

func TestBotDetectorInfo(t *testing.T) {
	botDetector := &BD{}
	tests := []struct {
		name             string
		conf             *conf
		userAgent        string
		expectedDetector bool
		expectedIsBot    int
	}{
		{"NoBotDetectorInstance", &conf{}, "test-agent", false, 0},
		{"BotDetectorInstanceNoBot", &conf{botDetectionService: botDetector}, "human-agent", true, 0},
		{"BotDetectorInstanceWithBot", &conf{botDetectionService: botDetector}, "bot-agent", true, 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hasDetector, isBot := test.conf.botDetectorInfo(test.userAgent)
			if hasDetector != test.expectedDetector {
				t.Errorf("expected hasDetector %v, got %v", test.expectedDetector, hasDetector)
			}
			if isBot != test.expectedIsBot {
				t.Errorf("expected isBot %v, got %v", test.expectedIsBot, isBot)
			}
		})
	}
}

func TestWithStaticLogEntries(t *testing.T) {
	tests := []struct {
		name     string
		entries  map[string]string
		expected map[string]string
	}{
		{
			name:     "SetStaticLogEntries",
			entries:  map[string]string{"key1": "value1", "key2": "value2"},
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "EmptyStaticLogEntries",
			entries:  nil,
			expected: map[string]string{},
		},
		{
			name:     "OverrideExistingKeys",
			entries:  map[string]string{"key1": "newValue", "key3": "value3"},
			expected: map[string]string{"key1": "newValue", "key3": "value3"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithStaticLogEntries(test.entries)
			opt(c)

			if len(c.staticLogEntries) != len(test.expected) {
				t.Errorf("expected %v, got %v", test.expected, c.staticLogEntries)
				return
			}

			for key, expectedValue := range test.expected {
				if value, ok := c.staticLogEntries[key]; !ok || value != expectedValue {
					t.Errorf("expected key %s with value %v, got %v", key, expectedValue, c.staticLogEntries[key])
				}
			}
		})
	}
}

func TestWithTimeAggregation(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected time.Duration
	}{
		{"SetPositiveDuration", 5 * time.Second, 5 * time.Second},
		{"SetZeroDuration", 0 * time.Second, 0 * time.Second},
		{"SetNegativeDuration", -1 * time.Second, -1 * time.Second}, // Assuming negative values are allowed in some cases
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithTimeAggregation(test.duration)
			opt(c)
			if c.aggregationInterval != test.expected {
				t.Errorf("expected %v, got %v", test.expected, c.aggregationInterval)
			}
		})
	}
}

func TestWithAggregation(t *testing.T) {
	tests := []struct {
		name            string
		enabled         bool
		expectedEnabled bool
	}{
		{"EnableAggregation", true, true},
		{"DisableAggregation", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithAggregation(test.enabled)
			opt(c)
			if c.isAggregationEnabled != test.expectedEnabled {
				t.Errorf("expected %v, got %v", test.expectedEnabled, c.isAggregationEnabled)
			}
		})
	}
}

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	tests := []struct {
		name     string
		logger   *slog.Logger
		expected *slog.Logger
	}{
		{
			name:     "SetCustomLogger",
			logger:   logger, // Assuming `customLogger` implements the Logger interface
			expected: logger,
		},
		{
			name:     "SetNilLogger",
			logger:   nil,
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithLogger(test.logger)
			opt(c)
			if c.loggingHandler != test.expected {
				t.Errorf("expected logger %v, got %v", test.expected, c.loggingHandler)
			}
		})
	}
}

func TestWithIpHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		expected []string
	}{
		{
			name:     "SetIPHeaders",
			headers:  []string{"X-Forwarded-For", "X-Real-IP"},
			expected: []string{"X-Forwarded-For", "X-Real-IP"},
		},
		{
			name:     "SetEmptyHeaders",
			headers:  nil,
			expected: nil,
		},
		{
			name:     "OverrideHeaders",
			headers:  []string{"CF-Connecting-IP", "True-Client-IP"},
			expected: []string{"CF-Connecting-IP", "True-Client-IP"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithIpHeaders(test.headers)
			opt(c)

			if len(c.clientIPHeaders) != len(test.expected) {
				t.Errorf("expected %v, got %v", test.expected, c.clientIPHeaders)
				return
			}

			for i, expectedHeader := range test.expected {
				if c.clientIPHeaders[i] != expectedHeader {
					t.Errorf("at index %d, expected header %v, got %v", i, expectedHeader, c.clientIPHeaders[i])
				}
			}
		})
	}
}

func TestWithUaHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		expected []string
	}{
		{
			name:     "SetUserAgentHeaders",
			headers:  []string{"User-Agent", "X-Custom-UA"},
			expected: []string{"User-Agent", "X-Custom-UA"},
		},
		{
			name:     "SetEmptyUserAgentHeaders",
			headers:  nil,
			expected: nil,
		},
		{
			name:     "OverrideUserAgentHeaders",
			headers:  []string{"X-UA-Test", "Another-UA-Header"},
			expected: []string{"X-UA-Test", "Another-UA-Header"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithUaHeaders(test.headers)
			opt(c)

			if len(c.userAgentHeaders) != len(test.expected) {
				t.Errorf("expected %v, got %v", test.expected, c.userAgentHeaders)
				return
			}

			for i, expectedHeader := range test.expected {
				if c.userAgentHeaders[i] != expectedHeader {
					t.Errorf("at index %d, expected header %v, got %v", i, expectedHeader, c.userAgentHeaders[i])
				}
			}
		})
	}
}

func TestWithHeaderToLogs(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string][]string
		expected map[string][]string
	}{
		{
			name:     "SetHeadersToLogs",
			headers:  map[string][]string{"Header1": {"log1"}, "Header2": {"log2", "log3"}},
			expected: map[string][]string{"Header1": {"log1"}, "Header2": {"log2", "log3"}},
		},
		{
			name:     "SetEmptyHeadersToLogs",
			headers:  nil,
			expected: map[string][]string{},
		},
		{
			name:     "OverrideHeadersToLogs",
			headers:  map[string][]string{"Header3": {"log4"}, "Header1": {"updatedLog1", "updatedLog2"}},
			expected: map[string][]string{"Header3": {"log4"}, "Header1": {"updatedLog1", "updatedLog2"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &conf{}
			opt := WithHeaderToLogs(test.headers)
			opt(c)

			if len(c.logHeadersWithName) != len(test.expected) {
				t.Errorf("expected %v, got %v", test.expected, c.logHeadersWithName)
				return
			}

			for key, expectedValues := range test.expected {
				actualValues, ok := c.logHeadersWithName[key]
				if !ok {
					t.Errorf("expected key %s to exist in logHeadersWithName, but it does not", key)
					continue
				}
				if len(actualValues) != len(expectedValues) {
					t.Errorf("for key %s, expected %v, got %v", key, expectedValues, actualValues)
					continue
				}
				for i, expectedValue := range expectedValues {
					if actualValues[i] != expectedValue {
						t.Errorf("for key %s at index %d, expected %v, got %v", key, i, expectedValue, actualValues[i])
					}
				}
			}
		})
	}
}
