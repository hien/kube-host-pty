<p align="center">
  <h3 align="center">rz</h3>
  <p align="center">ripzap - The fastest structured, leveled JSON logger for Go 📖. Dependency free.</p>
</p>

--------

[Make logging great again](https://kerkour.com/post/logging/)

[![GoDoc](https://godoc.org/github.com/bloom42/rz-go?status.svg)](https://godoc.org/github.com/bloom42/rz-go)
[![Build Status](https://travis-ci.org/bloom42/rz-go.svg?branch=master)](https://travis-ci.org/bloom42/rz-go)
[![GitHub release](https://img.shields.io/github/release/bloom42/rz-go.svg)](https://github.com/bloom42/rz-go/releases)
<!-- [![Coverage](http://gocover.io/_badge/github.com/bloom42/rz-go)](http://gocover.io/github.com/bloom42/rz-go) -->

![Console logging](docs/example_screenshot.png)

The rz package provides a fast and simple logger dedicated to JSON output avoiding allocations and reflection..

Uber's [zap](https://godoc.org/go.uber.org/zap) and rs's [zerolog](https://godoc.org/github.com/rs/zerolog)
libraries pioneered this approach.

ripzap is an update of zerolog taking this concept to the next level with a **simpler** to use and **safer**
API and even better [performance](#benchmarks).

To keep the code base and the API simple, ripzap focuses on efficient structured logging only.
Pretty logging on the console is made possible using the provided (but inefficient)
[`Formatter`s](https://godoc.org/github.com/bloom42/rz-go#LogFormatter).


1. [Quickstart](#quickstart)
2. [Configuration](#configuration)
3. [Field types](#field-types)
3. [HTTP Handler](#http-handler)
4. [Examples](#examples)
5. [Benchmarks](#benchmarks)
6. [Contributing](#contributing)
7. [License](#license)

-------------------

## Quickstart

```go
package main

import (
	"os"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
)

func main() {

	env := os.Getenv("GO_ENV")
	hostname, _ := os.Hostname()

	// update global logger's context fields
	log.Logger = log.Config(rz.With(func(e *rz.Event) {
		e.String("hostname", hostname).
			String("environment", env)
	}))

	if env == "production" {
		log.Logger = log.Config(rz.Level(rz.InfoLevel))
	}

	log.Info("info from logger", func(e *rz.Event) {
		e.String("hello", "world")
	})
	// {"level":"info","hostname":"","environment":"","hello":"world","timestamp":"2019-02-07T09:30:07Z","message":"info from logger"}
}
```


## Configuration

### Logger
```go
// Writer update logger's writer.
func Writer(writer io.Writer) LoggerOption {}
// Level update logger's level.
func Level(lvl LogLevel) LoggerOption {}
// Sampler update logger's sampler.
func Sampler(sampler LogSampler) LoggerOption {}
// AddHook appends hook to logger's hook
func AddHook(hook LogHook) LoggerOption {}
// Hooks replaces logger's hooks
func Hooks(hooks ...LogHook) LoggerOption {}
// With replaces logger's context fields
func With(fields func(*Event)) LoggerOption {}
// Stack enable/disable stack in error messages.
func Stack(enableStack bool) LoggerOption {}
// Timestamp enable/disable timestamp logging in error messages.
func Timestamp(enableTimestamp bool) LoggerOption {}
// Caller enable/disable caller field in message messages.
func Caller(enableCaller bool) LoggerOption {}
// Formatter update logger's formatter.
func Formatter(formatter LogFormatter) LoggerOption {}
// TimestampFieldName update logger's timestampFieldName.
func TimestampFieldName(timestampFieldName string) LoggerOption {}
// LevelFieldName update logger's levelFieldName.
func LevelFieldName(levelFieldName string) LoggerOption {}
// MessageFieldName update logger's messageFieldName.
func MessageFieldName(messageFieldName string) LoggerOption {}
// ErrorFieldName update logger's errorFieldName.
func ErrorFieldName(errorFieldName string) LoggerOption {}
// CallerFieldName update logger's callerFieldName.
func CallerFieldName(callerFieldName string) LoggerOption {}
// CallerSkipFrameCount update logger's callerSkipFrameCount.
func CallerSkipFrameCount(callerSkipFrameCount int) LoggerOption {}
// ErrorStackFieldName update logger's errorStackFieldName.
func ErrorStackFieldName(errorStackFieldName string) LoggerOption {}
// TimeFieldFormat update logger's timeFieldFormat.
func TimeFieldFormat(timeFieldFormat string) LoggerOption {}
// TimestampFunc update logger's timestampFunc.
func TimestampFunc(timestampFunc func() time.Time) LoggerOption {}
```

### Global
```go
var (
	// DurationFieldUnit defines the unit for time.Duration type fields added
	// using the Duration method.
	DurationFieldUnit = time.Millisecond

	// DurationFieldInteger renders Duration fields as integer instead of float if
	// set to true.
	DurationFieldInteger = false

	// ErrorHandler is called whenever rz fails to write an event on its
	// output. If not set, an error is printed on the stderr. This handler must
	// be thread safe and non-blocking.
	ErrorHandler func(err error)
)
```


## Field Types

### Standard Types

* `String`
* `Bool`
* `Int`, `Int8`, `Int16`, `Int32`, `Int64`
* `Uint`, `Uint8`, `Uint16`, `Uint32`, `Uint64`
* `Float32`, `Float64`

### Advanced Fields

* `Err`: Takes an `error` and render it as a string using the `logger.errorFieldName` field name.
* `Error`: Adds a field with a `error`.
* `Timestamp`: Insert a timestamp field with `logger.timestampFieldName` field name and formatted using `logger.timeFieldFormat`.
* `Time`: Adds a field with the time formated with the `logger.timeFieldFormat`.
* `Duration`: Adds a field with a `time.Duration`.
* `Dict`: Adds a sub-key/value as a field of the event.
* `Interface`: Uses reflection to marshal the type.


## HTTP Handler

See the [bloom42/rz-go/rzhttp](https://godoc.org/github.com/bloom42/rz-go/rzhttp) package or the
[example here](https://github.com/bloom42/rz-go/tree/master/examples/http).


## Examples

See the [examples](https://github.com/bloom42/rz-go/tree/master/examples) folder.


## Benchmarks

```
$ make benchmarks
cd benchmarks && ./run.sh
goos: linux
goarch: amd64
pkg: github.com/bloom42/rz-go/benchmarks
BenchmarkDisabledWithoutFields/sirupsen/logrus-4         	100000000	        16.8 ns/op	      16 B/op	       1 allocs/op
BenchmarkDisabledWithoutFields/uber-go/zap-4             	30000000	        37.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/rs/zerolog-4              	500000000	         3.69 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/bloom42/rz-go-4           	500000000	         3.32 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/sirupsen/logrus-4                 	  300000	      4495 ns/op	    1137 B/op	      24 allocs/op
BenchmarkWithoutFields/uber-go/zap-4                     	 5000000	       325 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/rs/zerolog-4                      	10000000	       120 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/bloom42/rz-go-4                   	10000000	       118 ns/op	       0 B/op	       0 allocs/op
Benchmark10Context/sirupsen/logrus-4                     	  100000	     19503 ns/op	    3261 B/op	      50 allocs/op
Benchmark10Context/uber-go/zap-4                         	 5000000	       328 ns/op	       0 B/op	       0 allocs/op
Benchmark10Context/rs/zerolog-4                          	10000000	       130 ns/op	       0 B/op	       0 allocs/op
Benchmark10Context/bloom42/rz-go-4                       	10000000	       129 ns/op	       0 B/op	       0 allocs/op
Benchmark10Fields/sirupsen/logrus-4                      	   50000	     24542 ns/op	    4043 B/op	      54 allocs/op
Benchmark10Fields/uber-go/zap-4                          	  500000	      3189 ns/op	     946 B/op	       8 allocs/op
Benchmark10Fields/rs/zerolog-4                           	  500000	      2449 ns/op	     640 B/op	       6 allocs/op
Benchmark10Fields/bloom42/rz-go-4                        	  500000	      2319 ns/op	     640 B/op	       6 allocs/op
Benchmark10Fields10Context/sirupsen/logrus-4             	   50000	     25675 ns/op	    4566 B/op	      53 allocs/op
Benchmark10Fields10Context/uber-go/zap-4                 	  500000	      3254 ns/op	     948 B/op	       8 allocs/op
Benchmark10Fields10Context/rs/zerolog-4                  	  500000	      2392 ns/op	     640 B/op	       6 allocs/op
Benchmark10Fields10Context/bloom42/rz-go-4               	  500000	      2369 ns/op	     640 B/op	       6 allocs/op
PASS
ok  	github.com/bloom42/rz-go/benchmarks	31.306s
```


## Contributing

See [https://opensource.bloom.sh/contributing](https://opensource.bloom.sh/contributing)


## License

See `LICENSE.txt` and [https://opensource.bloom.sh/licensing](https://opensource.bloom.sh/licensing)

From an original work by [rs](https://github.com/rs): [zerolog](https://github.com/rs/zerolog) - commit aa55558e4cb2e8f05cd079342d430f77e946d00a
