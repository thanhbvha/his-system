package utils

import (
	"runtime/debug"
	"time"

	"github.com/thanhbvha/go-common/logger"
)

// SafeGo runs a function in a new goroutine and safely recovers from any panic.
// It logs the panic error and stack trace asynchronously using the global default logger,
// preventing the main application from crashing.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorAsync("Goroutine panic recovered",
					"error", r,
					"stack", string(debug.Stack()),
					"dispatch_time", time.Now().Format(time.RFC3339Nano),
				)
			}
		}()

		fn()
	}()
}
