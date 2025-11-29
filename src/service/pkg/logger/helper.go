package logger

import (
	"fmt"
	"runtime"
	"strings"
)

// GetCallerFunctionName gets the actual function name of the caller (skips n frames)
func GetCallerFunctionName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	fullName := fn.Name()

	// Extract just the function name from full path
	// e.g. "github.com/user/project/service/user.(*Handler).CreateStudent" -> "CreateStudent"
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}

// BuildQueryString builds a query string with arguments for logging
func BuildQueryString(query string, args ...interface{}) string {
	if len(args) == 0 {
		return query
	}

	// Format query with args for logging
	argsStr := "["
	for i, arg := range args {
		if i > 0 {
			argsStr += ", "
		}
		argsStr += fmt.Sprintf("%v", arg)
	}
	argsStr += "]"

	return fmt.Sprintf("%s | args: %s", query, argsStr)
}
