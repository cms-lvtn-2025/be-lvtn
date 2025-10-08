package controller

import (
	"thaily/src/server/client"
)

type Controller struct {
	academic *client.GRPCAcadamicClient
	council  *client.GRPCCouncil
	file     *client.GRPCfile
	role     *client.GRPCRole
	thesis   *client.GRPCthesis
	user     *client.GRPCUser
}

// Constructor function
func NewController(academic *client.GRPCAcadamicClient, council *client.GRPCCouncil, file *client.GRPCfile, role *client.GRPCRole, thesis *client.GRPCthesis, user *client.GRPCUser) *Controller {
	return &Controller{
		academic: academic,
		council:  council,
		file:     file,
		role:     role,
		thesis:   thesis,
		user:     user,
	}
}
