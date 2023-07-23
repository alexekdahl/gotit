package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/AlexEkdahl/gotit/utils/colors"
)

type logger struct {
	env         string
	destination *os.File
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(err error)
	Debug(format string, v ...interface{})
	Close() error
}

type Config struct {
	Path string
	Env  string
}

type loggerPrefix struct {
	level string
	color string
}

func (cp loggerPrefix) toColor() string {
	return fmt.Sprintf(cp.color, cp.level)
}

var (
	infoPrefix  = loggerPrefix{"INFO:  ", colors.InfoColor}
	warnPrefix  = loggerPrefix{"WARN:  ", colors.WarnColor}
	errorPrefix = loggerPrefix{"ERROR: ", colors.ErrorColor}
	debugPrefix = loggerPrefix{"DEBUG: ", colors.DebugColor}
)

func NewLogger(c Config) (Logger, error) {
	var dest *os.File
	var err error

	switch c.Env {
	case "LOCAL":
		dest = os.Stdout
	case "PROD":
		dest, err = os.OpenFile(c.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
		if err != nil {
			return nil, err
		}
	default:
		dest = os.Stdout
	}

	return &logger{
		env:         c.Env,
		destination: dest,
		infoLogger:  newLog(dest, infoPrefix, c.Env),
		warnLogger:  newLog(dest, warnPrefix, c.Env),
		errorLogger: newLog(dest, errorPrefix, c.Env),
		debugLogger: newLog(dest, debugPrefix, c.Env),
	}, nil
}

func newLog(dest *os.File, prefix loggerPrefix, env string) *log.Logger {
	l := log.New(dest, prefix.level, log.Ldate|log.Ltime)
	if env == "LOCAL" {
		l.SetPrefix(prefix.toColor())
	}

	return l
}

func (l *logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

func (l *logger) Warn(format string, v ...interface{}) {
	l.warnLogger.Printf(format, v...)
}

func (l *logger) Error(err error) {
	l.errorLogger.Printf("%v", err)
}

func (l *logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

// Close the logger file
func (l *logger) Close() error {
	if l.env == "PROD" {
		return l.destination.Close()
	}
	return nil
}
