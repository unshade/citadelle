package repositories

import "github.com/unshade/citadelle/internal/models"

type ServerNodesRepo interface {
	Create(node models.ServerNode)
}

type ServerNodes struct {}
