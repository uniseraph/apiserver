package logging

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
)

var (
	loggers     loggerPool
	ProjectName string
)

type ContextLogger struct {
	l *logrus.Logger
	f logrus.Fields
}

type loggerPool struct {
	m *sync.Mutex
	l logrus.Level  // default level
	w io.Writer     // default writer
	f logrus.Fields // default fields
	p map[string]*ContextLogger
}

func init() {
	loggers = loggerPool{
		m: &sync.Mutex{},
		p: map[string]*ContextLogger{},
		l: logrus.InfoLevel,
		w: os.Stderr,
		f: logrus.Fields{},
	}
}

func addContextLogger(name string) {
	l := logrus.New()
	l.Out = loggers.w
	l.Level = loggers.l
	l.Formatter = &logrus.TextFormatter{FullTimestamp: true}

	// clone fields from default
	fields := logrus.Fields{"module": name}
	for k, v := range loggers.f {
		fields[k] = v
	}

	loggers.p[name] = &ContextLogger{l, fields}
}

func GetLogger(name string) *ContextLogger {
	loggers.m.Lock()
	defer loggers.m.Unlock()

	if _, exists := loggers.p[name]; !exists {
		addContextLogger(name)
	}
	return loggers.p[name]
}

func SetLevel(level logrus.Level) {
	loggers.m.Lock()
	defer loggers.m.Unlock()

	loggers.l = level

	for _, cl := range loggers.p {
		cl.l.Level = level
	}
}

func SetOutput(out io.Writer) {
	loggers.m.Lock()
	defer loggers.m.Unlock()

	loggers.w = out

	for _, cl := range loggers.p {
		cl.l.Out = out
	}
}

func AddField(key string, value interface{}) {
	loggers.m.Lock()
	defer loggers.m.Unlock()

	loggers.f[key] = value

	for _, cl := range loggers.p {
		cl.f[key] = value
	}
}

func getCaller() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 1
	} else {
		if strings.Contains(file, ProjectName) {
			file = strings.Split(file, ProjectName+"/")[1]
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func (cl *ContextLogger) WithField(key string, value interface{}) *logrus.Entry {
	return logrus.NewEntry(cl.l).WithFields(cl.f).WithField(key, value).WithField("file", getCaller())
}

func (cl *ContextLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	return logrus.NewEntry(cl.l).WithFields(cl.f).WithFields(fields).WithField("file", getCaller())
}

func (cl *ContextLogger) Debugf(format string, args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Debugf(format, args...)
}

func (cl *ContextLogger) Infof(format string, args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Infof(format, args...)
}

func (cl *ContextLogger) Printf(format string, args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Printf(format, args...)
}

func (cl *ContextLogger) Warnf(format string, args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Warnf(format, args...)
}

func (cl *ContextLogger) Warningf(format string, args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Warningf(format, args...)
}

func (cl *ContextLogger) Errorf(format string, args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Errorf(format, args...)
}

func (cl *ContextLogger) Fatalf(format string, args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Fatalf(format, args...)
	os.Exit(1)
}

func (cl *ContextLogger) Panicf(format string, args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Panicf(format, args...)
}

func (cl *ContextLogger) Debug(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Debug(args...)
}

func (cl *ContextLogger) Info(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Info(args...)
}

func (cl *ContextLogger) Print(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Print(args...)
}

func (cl *ContextLogger) Warn(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Warn(args...)
}

func (cl *ContextLogger) Warning(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Warning(args...)
}

func (cl *ContextLogger) Error(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Error(args...)
}

func (cl *ContextLogger) Fatal(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Fatal(args...)
}

func (cl *ContextLogger) Panic(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Panic(args...)
}

func (cl *ContextLogger) Debugln(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Debugln(args...)
}

func (cl *ContextLogger) Infoln(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Infoln(args...)
}

func (cl *ContextLogger) Println(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Println(args...)
}

func (cl *ContextLogger) Warnln(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Warnln(args...)
}

func (cl *ContextLogger) Warningln(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Warningln(args...)
}

func (cl *ContextLogger) Errorln(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Errorln(args...)
}

func (cl *ContextLogger) Fatalln(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Fatalln(args...)
}

func (cl *ContextLogger) Panicln(args ...interface{}) {
	cl.l.WithFields(cl.f).WithField("file", getCaller()).Panicln(args...)
}
