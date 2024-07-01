package recovery

import "main/tools/log"

// RecoverAndLog recovers from panics inside a goroutine and logs the panic information.
func RecoverAndLog() {
	if err := recover(); err != nil {
		log.GetSugger().Error("panic recovered", "err ", err)
	}
}

// GoRecovery is a utility function to recover from panics in a goroutine using defer.
func GoRecovery() {
	go func() {
		defer RecoverAndLog()
	}()
}
