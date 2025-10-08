package container

import (
	"fmt"

	"thaily/src/config"
	"thaily/src/server/client"
)

// Container chứa tất cả dependencies
type Container struct {
	Config   *config.Config
	Clients  *Clients
}

// Clients chứa tất cả gRPC clients
type Clients struct {
	Academic *client.GRPCAcadamicClient
	Council  *client.GRPCCouncil
	File     *client.GRPCfile
	Role     *client.GRPCRole
	Thesis   *client.GRPCthesis
	User     *client.GRPCUser
}

// New tạo container mới và khởi tạo tất cả dependencies
func New(cfg *config.Config) (*Container, error) {
	clients, err := initClients(cfg.Services)
	if err != nil {
		return nil, fmt.Errorf("failed to init clients: %w", err)
	}

	return &Container{
		Config:  cfg,
		Clients: clients,
	}, nil
}

func initClients(cfg config.ServiceConfig) (*Clients, error) {
	academic, err := client.NewGRPCAcadamicClient(cfg.Academic)
	if err != nil {
		return nil, fmt.Errorf("academic client: %w", err)
	}

	council, err := client.NewGRPCCouncil(cfg.Council)
	if err != nil {
		return nil, fmt.Errorf("council client: %w", err)
	}

	file, err := client.NewGRPCfile(cfg.File)
	if err != nil {
		return nil, fmt.Errorf("file client: %w", err)
	}

	role, err := client.NewGRPCRole(cfg.Role)
	if err != nil {
		return nil, fmt.Errorf("role client: %w", err)
	}

	thesis, err := client.NewGRPCthesis(cfg.Thesis)
	if err != nil {
		return nil, fmt.Errorf("thesis client: %w", err)
	}

	user, err := client.NewGRPCUser(cfg.User)
	if err != nil {
		return nil, fmt.Errorf("user client: %w", err)
	}

	return &Clients{
		Academic: academic,
		Council:  council,
		File:     file,
		Role:     role,
		Thesis:   thesis,
		User:     user,
	}, nil
}
