package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor creates a gRPC interceptor that adds tracing to all requests
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Add request ID and trace stack to context
		ctx = WithRequestID(ctx)
		ctx = WithTraceStack(ctx)

		// Record start time
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Get trace information
		stack := GetTraceStack(ctx)
		var trace *FunctionTrace
		if stack != nil {
			trace = stack.GetRoot()
		}

		// Extract request arguments (only actual values, not full object structure)
		var requestArgs map[string]interface{}
		if req != nil {
			// Try to marshal request to JSON for logging
			reqJSON, marshalErr := json.Marshal(req)
			if marshalErr == nil {
				var reqMap map[string]interface{}
				if json.Unmarshal(reqJSON, &reqMap) == nil {
					// Filter out empty/nil values and keep only actual arguments
					requestArgs = extractRequestArgs(reqMap)
				}
			}
		}

		// Print trace as JSON
		traceData := map[string]interface{}{
			"request_id":  GetRequestID(ctx),
			"method":      info.FullMethod,
			"duration_ms": duration.Milliseconds(),
			"success":     err == nil,
		}

		// Add request arguments if available
		if len(requestArgs) > 0 {
			traceData["request_args"] = requestArgs
		}

		if trace != nil {
			traceData["trace"] = trace
			traceData["queries"] = FlattenQueries(trace)
		}

		if err != nil {
			traceData["error"] = err.Error()
		}

		// Write trace to file logger
		fileLogger := GetFileLogger()
		if fileLogger != nil {
			fileLogger.WriteTrace(traceData)
		} else {
			// Fallback to console if file logger not initialized
			jsonData, _ := json.MarshalIndent(traceData, "", "  ")
			fmt.Printf("\n=== Request Trace ===\n%s\n====================\n\n", string(jsonData))
		}

		return resp, err
	}
}

// extractRequestArgs extracts only non-empty, non-nil values from request map
func extractRequestArgs(reqMap map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range reqMap {
		if value == nil {
			continue
		}

		// Check if value is empty based on type
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}

		// Skip empty values
		if isEmptyValue(val) {
			continue
		}

		// Handle nested objects - only include if they have actual values
		if val.Kind() == reflect.Map || val.Kind() == reflect.Struct {
			if nestedMap, ok := value.(map[string]interface{}); ok {
				nestedResult := extractRequestArgs(nestedMap)
				if len(nestedResult) > 0 {
					result[key] = nestedResult
				}
			} else {
				// For struct types, try to convert to map
				if jsonBytes, err := json.Marshal(value); err == nil {
					var nestedMap map[string]interface{}
					if json.Unmarshal(jsonBytes, &nestedMap) == nil {
						nestedResult := extractRequestArgs(nestedMap)
						if len(nestedResult) > 0 {
							result[key] = nestedResult
						}
					}
				}
			}
		} else {
			// Include primitive values and arrays
			result[key] = value
		}
	}

	return result
}

// isEmptyValue checks if a reflect.Value is empty
// Note: We only filter out truly empty values (nil, empty strings, empty arrays/maps)
// We keep 0, false as they might be valid arguments from client
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	// Don't filter out 0, false as they might be valid values
	return false
}
