# Gin Logger

This package implements logger middleware for private projects in Golang Gin, using the slog library. It supports:

- Real-time logging: Logs each request and response in real time.

- Aggregate logging: Aggregates and logs data based on specific criteria for analysis and reporting.

Note: Aggregate logging does not log headers and referrer.

## Configuration and Usage

To configure the Gin Logger in your project, follow the steps below:

### Import the Package

Make sure to import the logger package in your application:

```go
import "github.com/logocomune/gin-logger"
```

### Real-time Logger Configuration

To initialize and use the real-time logger:

```go
logger := slogger.New(
context.Background(),
"MyAppName",
"v1.0.0",
slogger.WithLogHeaders(true),
slogger.WithSkipPaths([]string{"/health", "/metrics"}),
)

// Add the middleware to your Gin router
router := gin.Default()
router.Use(logger.Middleware())
```

### Aggregate Logger Configuration

To initialize and use the aggregate logger:

```go
ctx := context.Background()

aggregator := slogger.New(
ctx,
"MyAppName",
"v1.0.0",
slogger.WithQueueSize(200),
slogger.WithTimeAggregation(10 * time.Second),
slogger.WithAggregatePath(func(route, path string, statusCode int) string {
// route: matched route
// path: url.Path
if statusCode >= 400 {
return "error"
}
return route
}),
slogger.WithBotDetector(customBotDetector),
)

// Add the middleware to your Gin router
router := gin.Default()
router.Use(aggregator.Middleware())
```

### Optional Configuration Parameters

You can customize the logger behavior using the following options:

- `WithLogHeaders(bool)`: Enables or disables logging of HTTP headers.
- `WithSkipPaths([]string)`: Specifies paths to skip logging.
- `WithQueueSize(int)`: Sets the queue size for aggregate logging. This is valid only if aggregation is enabled.
- `WithTimeAggregation(time.Duration)`: Sets the time duration for log aggregation. This is valid only if aggregation is enabled.
- `WithAggregatePath(func(route, path string, statusCode int) string)`: Defines a custom path aggregation function.
- `WithBotDetector(slogger.BotDetector)`: Enables bot detection based on user-agent strings.
- `WithIpHeaders([]string)`: Configures headers to extract client IP information.
- `WithHeaderToLogs(map[string][]string)`: Logs specific headers with assigned names.
- `WithLogQueryString(bool)`: Enables or disables logging of the query string in requests.
- `WithPathAggregator(func(route, path string, statusCode int) string)`: Sets a custom function for path aggregation.
- `WithLogger(*slog.Logger)`: Configures a custom logger instance for the application.
- `WithLogMessage(string)`: Customizes the log message format for the application.
- `WithUaHeaders([]string)`: Configures headers to extract user-agent information.
- `WithAggregation(bool)`: Enables or disables the aggregation feature.
- `WithStaticLogEntries(map[string]string)`: Includes static entries in all log messages.

### Start the Server

Finally, start your Gin server as usual:

```go
router.GET("/example", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "Hello, World!"})
})

router.Run(":8080")
```

This example demonstrates a typical setup for both real-time and aggregate logging in a Gin-based application. Customize the configuration as per your requirements.

