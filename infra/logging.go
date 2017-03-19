package infra

import (
	"fmt"
	"os"
)

var LEVEL_TRACE = 50
var LEVEL_DEBUG = 100
var LEVEL_INFO = 200
var LEVEL_WARNING = 300
var LEVEL_ERROR = 400
var LOG_LEVEL = LEVEL_DEBUG
var logHandler func(level int, levelName string, event string, kv []interface{})

func init() {
	logHandler = func(level int, levelName string, event string, kv []interface{}) {
		file := os.Stdout
		file.WriteString(fmt.Sprintf("[%s] %s\n", levelName, event))
		for i := 0; i < len(kv); i = i + 2 {
			file.WriteString("\t")
			file.WriteString(fmt.Sprintf("%v", kv[i]))
			file.WriteString(": ")
			file.WriteString(fmt.Sprintf("%v", kv[i + 1]))
			file.WriteString("\n")
		}
		file.Sync()
	}
}

func SetLogHandler(newHandler func(level int, levelName string, event string, kv []interface{})) {
	logHandler = newHandler
}

func ShouldLogDebug() bool {
	return LOG_LEVEL <= LEVEL_DEBUG
}

func ShouldLogTrace() bool {
	return LOG_LEVEL <= LEVEL_TRACE
}

func LogError(event string, kv ...interface{}) {
	if LOG_LEVEL <= LEVEL_ERROR {
		logHandler(LEVEL_ERROR, "ERROR", event, kv)
	}
}

func LogWarning(event string, kv ...interface{}) {
	if LOG_LEVEL <= LEVEL_WARNING {
		logHandler(LEVEL_WARNING, "WARNING", event, kv)
	}
}

func LogInfo(event string, kv ...interface{}) {
	if LOG_LEVEL <= LEVEL_INFO {
		logHandler(LEVEL_INFO, "INFO", event, kv)
	}
}

func LogDebug(event string, kv ...interface{}) {
	if ShouldLogDebug() {
		logHandler(LEVEL_DEBUG, "DEBUG", event, kv)
	}
}

func LogTrace(event string, kv ...interface{}) {
	if ShouldLogTrace() {
		logHandler(LEVEL_TRACE, "TRACE", event, kv)
	}
}
