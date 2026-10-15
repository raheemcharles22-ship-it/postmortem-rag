package graph

import (
	"github.com/raheemcharles/postmortem-rag/internal/rag"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Answerer *rag.Answerer
}
