package controller

import (
	"thaily/src/server/client"
)

type Controller struct {
	common *client.GRPCCommonClient
}

// Constructor function
func NewController(common *client.GRPCCommonClient) *Controller {
	return &Controller{common: common}
}
