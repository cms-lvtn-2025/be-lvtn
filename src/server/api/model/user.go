package model

import pbRole "thaily/proto/role"

// UserInfo contains extracted user information from JWT claims
type UserInfo struct {
	Role     string
	Roles    []pbRole.RoleType
	Semester string
	UserID   string
	IDs      []string
}
