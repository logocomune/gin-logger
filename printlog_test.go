package slogger

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestPrint(t *testing.T) {

	tests := []struct {
		name    string
		msg     string
		log     logEntry
		conf    conf
		wantLog string // Expected log output
	}{
		{
			name: "basic log entry",
			msg:  "test",
			log: logEntry{
				created:    time.Date(2025, time.September, 11, 3, 34, 22, 0, time.UTC),
				ip:         "127.0.0.1",
				remoteIp:   "192.168.1.1",
				ua:         "Go-http-client",
				method:     "GET",
				proto:      "HTTP/1.1",
				statusCode: 200,
				count:      1,
				realtimeDetails: realtimeDetails{
					latency: time.Millisecond * 50,
					path:    "/home",
				},
			},
			conf: conf{
				loggingHandler: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
			},
			wantLog: "level=INFO msg=test created=2025-09-11T03:34:22Z ip=127.0.0.1 remoteIp=192.168.1.1 ua=Go-http-client method=GET proto=HTTP/1.1 statusCode=200 counter=1 path=/home latency=50ms responseSize=0",
		},
		{
			name: "log with referer",
			msg:  "request with referer",
			log: logEntry{

				ip:         "192.168.1.100",
				ua:         "Mozilla",
				method:     "GET",
				statusCode: 404,
				created:    time.Date(2025, time.September, 11, 3, 34, 22, 0, time.UTC),

				count: 2,
				realtimeDetails: realtimeDetails{
					referer: "https://example.com",
					latency: time.Millisecond * 20,
				},
			},
			conf: conf{
				loggingHandler: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
			},
			wantLog: "level=INFO msg=\"request with referer\" created=2025-09-11T03:34:22Z ip=192.168.1.100 remoteIp=\"\" ua=Mozilla method=GET proto=\"\" statusCode=404 counter=2 referer=https://example.com latency=20ms responseSize=0", // Set expected output for the log
		},
		{
			name: "log with bot detection",
			msg:  "bot detected",
			log: logEntry{
				created:              time.Date(2025, time.September, 11, 3, 34, 22, 0, time.UTC),
				isBotDetectorEnabled: true,
				isBot:                1,
				statusCode:           403,
			},
			conf: conf{
				loggingHandler: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
			},
			wantLog: "level=INFO msg=\"bot detected\" created=2025-09-11T03:34:22Z ip=\"\" remoteIp=\"\" ua=\"\" method=\"\" proto=\"\" statusCode=403 counter=0 isBot=1 latency=0s responseSize=0", // Expected log
		},
		{
			name: "aggregate log entry",
			msg:  "aggregated data",
			log: logEntry{
				created:     time.Date(2025, time.September, 11, 3, 34, 22, 0, time.UTC),
				isAggregate: true,
				count:       5,
				aggregateDetails: aggregateDetails{
					sumLatency:       time.Millisecond * 250,
					minLatency:       time.Millisecond * 30,
					maxLatency:       time.Millisecond * 70,
					sumSizeRespoBody: 5000,
				},
			},
			conf: conf{
				loggingHandler: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
			},
			wantLog: " level=INFO msg=\"aggregated data\" created=2025-09-11T03:34:22Z ip=\"\" remoteIp=\"\" ua=\"\" method=\"\" proto=\"\" statusCode=0 counter=5 meanLatency=50ms minLatency=30ms maxLatency=70ms meanSizeRespBody=1000 sumSizeRespBody=5000", // Expected log
		},
		{
			name: "log with headers and query string",
			msg:  "log with headers",
			log: logEntry{
				created: time.Date(2025, time.September, 11, 3, 34, 22, 0, time.UTC),
				realtimeDetails: realtimeDetails{
					headers:     map[string]string{"Content-Type": "application/json"},
					queryString: "?id=123",
					latency:     time.Millisecond * 10,
					path:        "/api/data",
				},
				count: 1,
			},
			conf: conf{
				loggingHandler: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
				logQueryString: true,
				logHeaders:     true,
			},
			wantLog: " level=INFO msg=\"log with headers\" created=2025-09-11T03:34:22Z ip=\"\" remoteIp=\"\" ua=\"\" method=\"\" proto=\"\" statusCode=0 counter=1 path=/api/data queryString=\"?id=123\" fullHeaders=map[Content-Type:application/json] latency=10ms responseSize=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.conf.loggingHandler = slog.New(slog.NewTextHandler(&buf, nil))

			printLog(tt.msg, tt.log, &tt.conf)
			fmt.Println(buf.String())
			// Compare logs (when `wantLog` is set)
			if !strings.Contains(buf.String(), tt.wantLog) {
				t.Errorf("unexpected log output:\ngot: %s\nwant: %s", buf.String(), tt.wantLog)
			}
		})
	}
}
