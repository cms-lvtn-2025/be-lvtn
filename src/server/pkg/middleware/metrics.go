package middleware

import (
	"context"
	"time"

	"thaily/src/server/pkg/metrics"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

// GraphQL Metrics Middleware
func GraphQLMetrics() graphql.HandlerExtension {
	return &metricsExtension{}
}

type metricsExtension struct{}

func (e *metricsExtension) ExtensionName() string {
	return "GraphQLMetrics"
}

func (e *metricsExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// Track GraphQL operations
func (e *metricsExtension) AroundOperations(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	start := time.Now()

	oc := graphql.GetOperationContext(ctx)
	operation := "unknown"
	operationType := "query"

	if oc.Operation != nil {
		if oc.Operation.Name != "" {
			operation = oc.Operation.Name
		}
		operationType = string(oc.Operation.Operation)
	}

	return func(ctx context.Context) *graphql.Response {
		responseHandler := next(ctx)
		response := responseHandler(ctx)
		duration := time.Since(start)

		status := "success"
		if response != nil && response.Errors != nil && len(response.Errors) > 0 {
			status = "error"
		}

		// Record metrics
		metrics.RecordGraphQLRequest(operation, status, duration)
		metrics.RecordGraphQLOperation(operationType, operation, status)

		return response
	}
}

// Track field resolution
func (e *metricsExtension) AroundFields(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return next(ctx)
	}

	start := time.Now()
	fieldName := fc.Field.Name
	typeName := fc.Object

	res, err := next(ctx)
	duration := time.Since(start)

	// Record field resolution metrics
	metrics.RecordFieldResolution(fieldName, typeName, duration)

	return res, err
}

func (e *metricsExtension) AroundResponses(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	return next(ctx)
}

// Business Logic Middleware
func BusinessMetrics() graphql.HandlerExtension {
	return &businessMetricsExtension{}
}

type businessMetricsExtension struct{}

func (e *businessMetricsExtension) ExtensionName() string {
	return "BusinessMetrics"
}

func (e *businessMetricsExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (e *businessMetricsExtension) AroundOperations(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	return func(ctx context.Context) *graphql.Response {
		oc := graphql.GetOperationContext(ctx)

		// Track business operations based on GraphQL operation
		if oc.Operation != nil {
			e.trackBusinessOperation(oc.Operation)
		}

		responseHandler := next(ctx)
		return responseHandler(ctx)
	}
}

func (e *businessMetricsExtension) trackBusinessOperation(operation *ast.OperationDefinition) {
	for _, selection := range operation.SelectionSet {
		if field, ok := selection.(*ast.Field); ok {
			switch field.Name {
			// Thesis operations
			case "createThesis", "updateThesis", "deleteThesis":
				metrics.RecordThesisOperation(field.Name, "success")
			case "submitThesis":
				metrics.RecordThesisOperation("submission", "success")

			// User operations
			case "login", "logout", "register":
				metrics.RecordUserActivity(field.Name, "user")
			case "createUser", "updateUser":
				metrics.RecordUserActivity(field.Name, "admin")

			// File operations
			case "uploadFile":
				metrics.RecordFileOperation("upload", "document", "success")
			case "downloadFile":
				metrics.RecordFileOperation("download", "document", "success")
			case "deleteFile":
				metrics.RecordFileOperation("delete", "document", "success")
			}
		}
	}
}

func (e *businessMetricsExtension) AroundFields(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	return next(ctx)
}

func (e *businessMetricsExtension) AroundResponses(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	return next(ctx)
}
