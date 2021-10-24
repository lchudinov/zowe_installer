package launcher

import (
	"regexp"
)

type LogLevel int

const (
	LogLevelError LogLevel = iota + 1
	LogLevelWarning
	LogLevelInfo
	LogLevelDebug
	LogLevelAny
)

func (l LogLevel) String() string {
	return [...]string{"Error", "Warning", "Info", "Debug", "Any"}[l-1]
}

var loglevelRe *regexp.Regexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} <[^>]+> \w+ (\w+)`)

func getLogLevel(line string) LogLevel {
	matches := loglevelRe.FindStringSubmatch(line)
	if matches == nil || len(matches) == 1 {
		return LogLevelAny
	}
	switch matches[1] {
	case "INFO":
		return LogLevelInfo
	case "ERROR":
		return LogLevelError
	case "WARN":
		return LogLevelWarning
	case "DEBUG":
		return LogLevelDebug
	default:
		return LogLevelAny
	}
}
