package slogger

import (
	"context"
	"strconv"
	"time"
)

// initLoggerAggregator initializes the logging aggregator, periodically processing and emitting aggregated log statistics.
func (a *Logger) initLoggerAggregator(ctx context.Context) {
	c := a.queue
	duration := a.conf.aggregationInterval
	t := time.NewTicker(duration)
	logEntries := make(map[string]logEntry)
	go func() {
		for {
			select {
			case <-ctx.Done():
				t.Stop()
				a.printLogs(logEntries, duration)
				logEntries = make(map[string]logEntry)

			case <-t.C:
				a.printLogs(logEntries, duration)

				logEntries = make(map[string]logEntry)
			case st := <-c:

				//Update stats
				var v logEntry
				var ok bool

				key := st.ip + "_" + strconv.Itoa(st.statusCode) + "_" +
					st.ua + "_" + st.method + "_" + st.proto + "_" + "_" + st.aggregatePath
				if v, ok = logEntries[key]; !ok {
					hasBotDetector, isBot := a.conf.botDetectorInfo(st.ua)
					v = logEntry{
						created:     time.Now().UTC(),
						ip:          st.ip,
						remoteIp:    st.remoteIp,
						ua:          st.ua,
						method:      st.method,
						statusCode:  st.statusCode,
						isAggregate: true,

						aggregateDetails: aggregateDetails{
							lastMod:          time.Now().UTC(),
							sumLatency:       0,
							maxLatency:       st.latency,
							minLatency:       st.latency,
							sumSizeRespoBody: st.responseBodySize,
						},

						isBotDetectorEnabled: hasBotDetector,
						isBot:                isBot,
						proto:                st.proto,
						aggregatePath:        st.aggregatePath,
					}
				}
				v.count++
				if v.maxLatency < st.latency {
					v.maxLatency = st.latency
				}
				if v.minLatency > st.latency {
					v.minLatency = st.latency
				}

				v.sumLatency += st.latency
				v.sumSizeRespoBody += st.responseBodySize
				logEntries[key] = v

			}
		}
	}()
}

// botDetectorInfo determines if a bot detector instance is enabled and checks if the provided user agent represents a bot.
func (c *conf) botDetectorInfo(userAgent string) (hasBotDetector bool, isBot int) {
	hasBotDetector = false
	isBot = 0
	if c.botDetectionService != nil {
		hasBotDetector = true
		isBot = 0
		if c.botDetectionService.IsBot(userAgent) {
			isBot = 1
		}

	}
	return hasBotDetector, isBot
}

// printLogs processes and prints aggregated log entries for a specified duration, utilizing the provided statistics map.
func (a *Logger) printLogs(stats map[string]logEntry, duration time.Duration) {
	if len(stats) == 0 {
		return
	}
	for _, v := range stats {

		printLog("api_logger v1", v, a.conf)

	}
}
