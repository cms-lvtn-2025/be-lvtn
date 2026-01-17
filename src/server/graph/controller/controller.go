package controller

import (
	"context"
	"fmt"
	"strings"

	"thaily/src/server/client"
	"thaily/src/server/graph/directive"
	"thaily/src/server/graph/helper"
	"thaily/src/server/graph/model"

	"github.com/golang-jwt/jwt/v5"
)

// Controller handles all GraphQL resolver business logic
type Controller struct {
	academic *client.GRPCAcadamicClient
	council  *client.GRPCCouncil
	file     *client.GRPCfile
	role     *client.GRPCRole
	thesis   *client.GRPCthesis
	user     *client.GRPCUser
}

// NewController creates a new Controller instance
func NewController(
	academic *client.GRPCAcadamicClient,
	council *client.GRPCCouncil,
	file *client.GRPCfile,
	role *client.GRPCRole,
	thesis *client.GRPCthesis,
	user *client.GRPCUser,
) *Controller {
	return &Controller{
		academic: academic,
		council:  council,
		file:     file,
		role:     role,
		thesis:   thesis,
		user:     user,
	}
}

// ============================================
// PAGINATION & SEARCH HELPERS
// ============================================

// DefaultPagination returns default pagination settings
func (c *Controller) DefaultPagination() *model.PaginationInput {
	page := int32(1)
	pageSize := int32(10)
	sortBy := "created_at"
	descending := true
	return &model.PaginationInput{
		Page:       &page,
		PageSize:   &pageSize,
		SortBy:     &sortBy,
		Descending: &descending,
	}
}

// ConvertSearchRequestToPB converts GraphQL SearchRequestInput to Protobuf SearchRequest

// ============================================
// AUTH & RBAC HELPERS
// ============================================

// GetInfoRequest extracts user ID and role from JWT context
func (c *Controller) GetInfoRequest(ctx context.Context) (id *string, role *string, err error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("not authorized")
	}

	roleSystem, ok := claims["role"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("not authorized")
	}

	semester, _ := ctx.Value("semester").(string)
	idsArr := strings.Split(claims["ids"].(string), ",")

	myId := ""
	if semester == "" {
		myId = strings.Split(idsArr[0], ":")[1]
	} else {
		for _, id := range idsArr {
			if strings.HasPrefix(id, semester+":") {
				myId = strings.Split(id, ":")[1]
			}
		}
	}

	if myId == "" {
		return nil, nil, fmt.Errorf("no user found for semester %s", semester)
	}

	return &myId, &roleSystem, nil
}

// GetInfoAllRequest extracts all user IDs and semesters from JWT context
func (c *Controller) GetInfoAllRequest(ctx context.Context) (ids *[]string, semesters *[]string, role *string, err error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	if !ok {
		return nil, nil, nil, fmt.Errorf("not authorized")
	}

	roleSystem, ok := claims["role"].(string)
	if !ok {
		return nil, nil, nil, fmt.Errorf("not authorized")
	}

	idsArr := strings.Split(claims["ids"].(string), ",")
	var myIds []string
	var mySemesters []string

	for _, id := range idsArr {
		parts := strings.Split(id, ":")
		if len(parts) == 2 {
			myIds = append(myIds, parts[1])
			mySemesters = append(mySemesters, parts[0])
		}
	}

	if len(myIds) == 0 {
		return nil, nil, nil, fmt.Errorf("no id found")
	}

	return &myIds, &mySemesters, &roleSystem, nil
}

// RbacInfo checks if user has specific role and returns their ID
func (c *Controller) RbacInfo(ctx context.Context, roleCheck model.RoleSystemRole) (id *string, check bool, err error) {
	myId, role, err := c.GetInfoRequest(ctx)
	if err != nil {
		return nil, false, err
	}

	if role == nil || (*role) != "teacher" {
		return nil, false, fmt.Errorf("not a teacher")
	}

	roles, err := c.GetRole(ctx, *myId)
	if err != nil {
		return nil, false, err
	}

	if roles == nil {
		return nil, false, fmt.Errorf("no roles found")
	}

	for _, r := range roles {
		if roleCheck == r.Role {
			return myId, true, nil
		}
	}

	return nil, false, fmt.Errorf("role %s not found", roleCheck)
}

// GetRbacDynamicRole checks if context has one of the specified roles (set by directive)
func (c *Controller) GetRbacDynamicRole(ctx context.Context, roles []string) (bool, error) {
	return directive.HasRole(ctx, roles...), nil
}

// ============================================
// INTERNAL CONVERSION HELPERS
// ============================================
