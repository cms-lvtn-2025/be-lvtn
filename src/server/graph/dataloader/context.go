package dataloader

import (
	"context"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const loadersKey contextKey = "dataloaders"

// WithLoaders adds dataloaders to the context
func WithLoaders(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey, loaders)
}

// GetLoaders retrieves dataloaders from the context
func GetLoaders(ctx context.Context) *Loaders {
	loaders, ok := ctx.Value(loadersKey).(*Loaders)
	if !ok {
		return nil
	}
	return loaders
}

// MustGetLoaders retrieves dataloaders from the context or panics
func MustGetLoaders(ctx context.Context) *Loaders {
	loaders := GetLoaders(ctx)
	if loaders == nil {
		panic("dataloaders not found in context")
	}
	return loaders
}
