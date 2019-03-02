package log

import (
	"sync"

	"github.com/bloom42/rz-go"
)

var (
	logger   rz.Logger
	initOnce sync.Once
)

type Level string

const (
	DebugLevel   = Level("debug")
	InfoLevel    = Level("info")
	WarningLevel = Level("warning")
	ErrorLevel   = Level("error")
	FatalLevel   = Level("fatal")
	PanicLevel   = Level("panic")
)

var (
	strLevelMap = map[Level]rz.LogLevel{
		DebugLevel:   rz.DebugLevel,
		InfoLevel:    rz.InfoLevel,
		WarningLevel: rz.WarnLevel,
		ErrorLevel:   rz.ErrorLevel,
		FatalLevel:   rz.FatalLevel,
		PanicLevel:   rz.PanicLevel,
	}
)

func Setup(appName string, level Level) {
	initOnce.Do(func() {
		logger = rz.New(
			rz.Level(strLevelMap[level]),
			rz.LevelFieldName("level"),
			rz.Formatter(rz.FormatterConsole()),
			rz.Timestamp(true),
			rz.TimestampFieldName("timestamp"),
			// caller setup
			rz.Caller(true),
			rz.CallerFieldName("file"),
			rz.CallerSkipFrameCount(4),
			rz.ErrorFieldName("error"),
			rz.With(func(e *rz.Event) {
				e.String("app", appName)
			}),
		)
	})
}

func logFields(fields []Field, event *rz.Event) {
	for _, f := range fields {
		f(event)
	}
}

func D(msg string, fields ...Field) {
	logger.Debug(msg, func(event *rz.Event) { logFields(fields, event) })
}
func I(msg string, fields ...Field) {
	logger.Info(msg, func(event *rz.Event) { logFields(fields, event) })
}
func W(msg string, fields ...Field) {
	logger.Warn(msg, func(event *rz.Event) { logFields(fields, event) })
}
func E(msg string, fields ...Field) {
	logger.Error(msg, func(event *rz.Event) { logFields(fields, event) })
}
func F(msg string, fields ...Field) {
	logger.Fatal(msg, func(event *rz.Event) { logFields(fields, event) })
}
func P(msg string, fields ...Field) {
	logger.Panic(msg, func(event *rz.Event) { logFields(fields, event) })
}
