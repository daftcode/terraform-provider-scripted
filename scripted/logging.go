package scripted

import "github.com/hashicorp/go-hclog"

type Logger struct {
	hcloggers []hclog.Logger
	logFn     func(msg string, args ...interface{})
}

type Logging struct {
	stack []*Logger
	level hclog.Level
}

func newLogging(hcloggers []hclog.Logger, args ...interface{}) *Logging {
	logger := &Logger{}
	logger.append(hcloggers...)
	ret := &Logging{
		stack: []*Logger{
			logger,
		},
	}
	if len(args) > 0 {
		ret.Push(args...)
	}
	return ret
}

func (ls *Logging) PushDefer(args ...interface{}) func() {
	logger := ls.Push(args...)
	return func() {
		ls.PopIf(logger)
	}
}

func (ls *Logging) Push(args ...interface{}) *Logger {
	l := len(ls.stack)
	logger := ls.stack[l-1]
	ret := &Logger{}
	for _, hl := range logger.hcloggers {
		ret.append(hl.With(args...))
	}
	ls.stack = append(ls.stack, ret)
	// ret.Log(hclog.Trace, "[LOGGING] pushed logger", "logger", ret, "length", len(ls.stack))
	return ret
}

func (ls *Logging) PopIf(expected *Logger) *Logger {
	// expected.Log(hclog.Trace, "[LOGGING] popping logger", "logger", expected, "length", len(ls.stack))
	s := ls.stack
	l := len(s)
	logger := s[l-1]
	if logger != expected {
		return nil
	}
	ls.stack = s[:l-1]
	return logger
}

func (ls *Logging) Log(level hclog.Level, msg string, args ...interface{}) {
	ls.stack[len(ls.stack)-1].Log(level, msg, args...)
}

func (ls *Logging) Clone() *Logging {
	return &Logging{
		stack: append([]*Logger{}, ls.stack...),
		level: ls.level,
	}
}

func (l *Logger) With(args ...interface{}) *Logger {
	ret := &Logger{}
	for _, l := range l.hcloggers {
		ret.append(l.With(args...))
	}
	return ret
}

func (l *Logger) append(loggers ...hclog.Logger) {
	l.hcloggers = append(l.hcloggers, loggers...)
}

func (l *Logger) Log(level hclog.Level, msg string, args ...interface{}) {
	for _, hl := range l.hcloggers {
		selectLogFunction(hl, level)(msg, args...)
	}
}

func selectLogFunction(logger hclog.Logger, level hclog.Level) func(msg string, args ...interface{}) {
	switch level {
	case hclog.Trace:
		return logger.Trace
	case hclog.Debug:
		return logger.Debug
	case hclog.Info:
		return logger.Info
	case hclog.Warn:
		return logger.Warn
	case hclog.Error:
		return logger.Error
	default:
		return logger.Info
	}
}
