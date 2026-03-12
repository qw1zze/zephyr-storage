package service

import (
	"log/slog"

	"zephyr-storage/internal/ipfs"
)

type Service struct {
	ipfs *ipfs.Client
	log  *slog.Logger
}

func NewService(ipfsClient *ipfs.Client, log *slog.Logger) *Service {
	return &Service{
		ipfs: ipfsClient,
		log:  log,
	}
}
