package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyData = errors.New("data must not be empty")
	ErrTooLarge = errors.New("data exceeds size limit")
	ErrEmptyCID = errors.New("cid must not be empty")
	ErrInvalidCID = errors.New("invalid cid format")
)

func (s *Service) Upload(ctx context.Context, data []byte) (string, error) {
	s.log.InfoContext(ctx, "service: upload", "size", len(data))

	if len(data) == 0 {
		return "", ErrEmptyData
	}

	if int64(len(data)) > s.maxSizeB {
		return "", ErrTooLarge
	}

	cid, err := s.ipfs.Add(ctx, data)
	if err != nil {
		return "", fmt.Errorf("service: upload: %w", err)
	}

	s.log.InfoContext(ctx, "service: upload: success", "cid", cid)
	return cid, nil
}

func (s *Service) Get(ctx context.Context, cid string) ([]byte, error) {
	s.log.InfoContext(ctx, "service: get", "cid", cid)

	if cid == "" {
		return nil, ErrEmptyCID
	}

	if !isValidCID(cid) {
		return nil, ErrInvalidCID
	}

	data, err := s.ipfs.Get(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("service: get: %w", err)
	}

	s.log.InfoContext(ctx, "service: get: success", "cid", cid, "size", len(data))
	return data, nil
}

func isValidCID(cid string) bool {
	n := len(cid)
	if n < 46 || n > 59 {
		return false
	}
	return strings.HasPrefix(cid, "Qm") || strings.HasPrefix(cid, "bafy")
}
