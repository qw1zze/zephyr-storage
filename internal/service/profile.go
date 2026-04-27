package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrEmptyName = errors.New("name must not be empty")

type Profile struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

func (s *Service) SaveProfile(ctx context.Context, name, avatar string) (string, error) {
	s.log.InfoContext(ctx, "service: save profile", "name", name)

	if name == "" {
		return "", ErrEmptyName
	}

	data, err := json.Marshal(Profile{Name: name, Avatar: avatar})
	if err != nil {
		return "", fmt.Errorf("service: save profile: marshal: %w", err)
	}

	cid, err := s.ipfs.Add(ctx, data)
	if err != nil {
		return "", fmt.Errorf("service: save profile: %w", err)
	}

	s.log.InfoContext(ctx, "service: save profile: success", "cid", cid)
	return cid, nil
}

func (s *Service) GetProfile(ctx context.Context, cid string) (*Profile, error) {
	s.log.InfoContext(ctx, "service: get profile", "cid", cid)

	if cid == "" {
		return nil, ErrEmptyCID
	}

	if !isValidCID(cid) {
		return nil, ErrInvalidCID
	}

	data, err := s.ipfs.Get(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("service: get profile: %w", err)
	}

	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("service: get profile: unmarshal: %w", err)
	}

	s.log.InfoContext(ctx, "service: get profile: success", "cid", cid)
	return &p, nil
}
