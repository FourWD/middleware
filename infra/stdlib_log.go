package infra

import (
	stdlog "log"
	"strings"
)

// stdlibLogWriter adapts an io.Writer so that stdlib "log" output is routed
// through the structured Logger instead of being printed raw to stdout.
type stdlibLogWriter struct {
	logger *Logger
}

func (w *stdlibLogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	if msg == "" {
		return len(p), nil
	}
	w.logger.Info(M(msg),
		WithComponent("stdlib_log"),
		WithOperation("write"),
		WithLogKind("fallback"),
		WithoutSource(),
	)
	return len(p), nil
}

// redirectStdlibLog routes the stdlib "log" package output through the given
// structured Logger. NewApp calls this once so that log.Print/Println/Printf
// from this project and any third-party library appear in the same JSON
// format as the other infra log entries.
//
// log.SetFlags(0) strips the default "2006/01/02 15:04:05" prefix that stdlib
// prepends, since the structured Logger already adds its own timestamp.
func redirectStdlibLog(logger *Logger) {
	stdlog.SetFlags(0)
	stdlog.SetOutput(&stdlibLogWriter{logger: logger})
}
