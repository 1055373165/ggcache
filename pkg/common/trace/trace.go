package trace

import (
	"fmt"
	"runtime"
	"strings"
)

// show runtime stack information when errors occur
func Trace(errMessage string) string {
	var pcstack [32]uintptr
	/*
		Callers fills pc, the fragment of the calling goroutine's stack, with the return program counter of the calling function.
		The parameter skip is the number of stack frames to be skipped before recording them in pc, with 0 being the Callers' own frame and 1 being the Callers' caller.
		It returns the number of entries written to pc.
	*/
	n := runtime.Callers(3, pcstack[:])

	// using string  Builder optimize speed.
	var str strings.Builder
	str.WriteString(errMessage + "\nTraceback:")
	for _, pc := range pcstack[:n] {
		function := runtime.FuncForPC(pc)
		file, line := function.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}

	return str.String()
}
