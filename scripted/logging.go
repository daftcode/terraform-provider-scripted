package scripted

import "github.com/hashicorp/go-hclog"

type Logger struct {
	hcloggers []hclog.Logger
	logFn     func(msg string, args ...interface{})
}

type Loggers struct {
	stack []*Logger
}

func initLoggers(s *Scripted, args ...interface{}) *Loggers {
	logger := &Logger{}
	logger.append(s.pc.Logger)
	logger.append(s.pc.FileLogger)
	loggers := &Loggers{
		stack: []*Logger{
			logger,
		},
	}
	if len(args) > 0 {
		loggers.Push(args...)
	}
	return loggers
}

func (ls *Loggers) Push(args ...interface{}) *Logger {
	l := len(ls.stack)
	logger := ls.stack[l-1]
	ret := &Logger{}
	for _, hl := range logger.hcloggers {
		ret.append(hl.With(args...))
	}
	ls.stack = append(ls.stack, ret)
	return ret
}

func (ls *Loggers) PopIf(expected *Logger) *Logger {
	s := ls.stack
	l := len(s)
	logger := s[l-1]
	if logger != expected {
		return nil
	}
	ls.stack = s[:l-1]
	return logger
}

func (ls *Loggers) Log(level hclog.Level, msg string, args ...interface{}) {
	ls.stack[len(ls.stack)-1].Log(level, msg, args...)
}

func (l *Logger) With(args ...interface{}) *Logger {
	ret := &Logger{}
	for _, l := range l.hcloggers {
		ret.append(l.With(args...))
	}
	return ret
}

func (l *Logger) append(logger hclog.Logger) {
	if logger != nil {
		l.hcloggers = append(l.hcloggers, logger)
	}
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
