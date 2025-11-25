package resolver

import (
	"context"
	"thaily/src/server/graph/controller"
	"thaily/src/server/graph/directive"
)

// This file will not be regenerated automatically.
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Ctrl *controller.Controller
}

// RbacRoleDynamic set role vào context (isolated cho mỗi field subtree)
func (r *Resolver) RbacRoleDynamic(ctx context.Context, role string) {
	directive.SetRole(ctx, role)
}

// RbacAccess kiểm tra role từ context
// Trả về true nếu current role nằm trong danh sách roles cho phép
func (r *Resolver) RbacAccess(ctx context.Context, roles []string) (bool, error) {
	return directive.HasRole(ctx, roles...), nil
}
