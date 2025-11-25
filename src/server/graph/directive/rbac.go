package directive

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

// rbacKey is unexported - cannot be injected by client
type rbacKey struct{}

// RbacState holds the role for a query subtree
type RbacState struct {
	Role string
}

// RbacRoleDirective implements @rbacRole directive (optional - for schema-level role)
// Sets role for the field and all its children
func RbacRoleDirective(ctx context.Context, obj interface{}, next graphql.Resolver, role string) (interface{}, error) {
	state := &RbacState{Role: role}
	ctx = context.WithValue(ctx, rbacKey{}, state)
	return next(ctx)
}

// RbacFieldMiddleware creates isolated context for each root field
// This prevents race conditions when multiple root fields run in parallel
func RbacFieldMiddleware(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)

	// Nếu là root field (Query/Mutation) -> tạo state mới
	if fc != nil && fc.Parent != nil && fc.Parent.Parent == nil {
		// Parent là Query/Mutation, không có grandparent -> đây là root field
		state := &RbacState{}
		ctx = context.WithValue(ctx, rbacKey{}, state)
	}

	return next(ctx)
}

// SetRole sets role in context (called by resolver)
// Mutates the RbacState pointer so children can read it
func SetRole(ctx context.Context, role string) {
	if state, ok := ctx.Value(rbacKey{}).(*RbacState); ok && state != nil {
		state.Role = role
	}
}

// GetRole returns the current role from context
func GetRole(ctx context.Context) string {
	if state, ok := ctx.Value(rbacKey{}).(*RbacState); ok && state != nil {
		return state.Role
	}
	return ""
}

// HasRole checks if current context has one of the specified roles
func HasRole(ctx context.Context, roles ...string) bool {
	role := GetRole(ctx)
	if role == "" {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
