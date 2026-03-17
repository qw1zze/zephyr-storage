package service

import (
	"context"
	"log/slog"
)

type ipfsClient interface {
	Add(ctx context.Context, data []byte) (string, error)
	Get(ctx context.Context, cid string) ([]byte, error)
}

type Service struct {
	ipfs     ipfsClient
	maxSizeB int64
	log      *slog.Logger
}

func NewService(client ipfsClient, maxSizeMB int64, log *slog.Logger) *Service {
	return &Service{
		ipfs:     client,
		maxSizeB: maxSizeMB * 1024 * 1024,
		log:      log,
	}
}
