// Package rz provides a lightweight logging library dedicated to JSON logging.
//
// A global Logger can be use for simple logging:
//
//     import "github.com/bloom42/rz-go/log"
//
//     log.Info("hello world", nil)
//
//     // Output: {"timestamp":"2019-02-07T09:30:07Z","level":"info","message":"hello world"}
//
// NOTE: To import the global logger, import the "log" subpackage "github.com/bloom42/rz-go/log".
//
// Fields can be added to log messages:
//
//     log.Info("hello world", func(e *rz.Event) { e.String("foo", "bar") })
//
//     // Output: {"timestamp":"2019-02-07T09:30:07Z","level":"info","message":"hello world","foo":"bar"}
//
// Create logger instance to manage different outputs:
//
//     logger := rz.New()
//     log.Info("hello world", func(e *rz.Event) {
//          e.String("foo", "bar") })
//     })
//
//     // Output: {"timestamp":"2019-02-07T09:30:07Z","level":"info","message":"hello world","foo":"bar"}
//
// Sub-loggers let you chain loggers with additional context:
//
//     sublogger := log.Config(rz.With(func(e *Event) {
//         e.String("component": "foo")
//     }))
//     sublogger.Info("hello world", nil)
//
//     // Output: {"timestamp":"2019-02-07T09:30:07Z","level":"info","message":"hello world","component":"foo"}
package rz
