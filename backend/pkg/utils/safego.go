package utils

import (
	"fmt"
	"runtime/debug"

	"github.com/thanhbvha/go-common/logger"
)

// SafeGo runs a function in a new goroutine and safely recovers from any panic.
// It logs the panic error and stack trace asynchronously using the global default logger,
// preventing the main application from crashing.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				errStr := fmt.Sprintf("%v", r)
				
				logger.ErrorAsync("Goroutine panic recovered",
					"error", errStr,
					"stack", stack,
				)
			}
		}()

		fn()
	}()
}
