package launcher

import (
	"fmt"
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

var stringLevels []string = []string{"Error", "Warning", "Info", "Debug", "Any"}

func (l LogLevel) String() string {
	return stringLevels[l-1]
}

func parseLogLevel(level string) (LogLevel, error) {
	for index, stringLevel := range stringLevels {
		if level == stringLevel {
			return LogLevel(index + 1), nil
		}
	}
	return 0, fmt.Errorf("unknown log level: %s", level)
}

var loglevelRe *regexp.Regexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(\.\d{3})* <[^>]+> \w+ (\w+)`)

func getLogLevel(line string) LogLevel {
	matches := loglevelRe.FindStringSubmatch(line)
	if matches == nil || len(matches) < 3 {
		return LogLevelAny
	}
	switch matches[2] {
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
