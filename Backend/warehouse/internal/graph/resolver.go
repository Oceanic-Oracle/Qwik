package graph

import (
	"log/slog"
	"warehouse/internal/repo"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	Repo *repo.Repo
	Log *slog.Logger
}
