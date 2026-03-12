package service

import "context"

func (s *Service) Upload(ctx context.Context, data []byte) (string, error) {
	s.log.InfoContext(ctx, "upload blob called", "size", len(data))
	return "not_implemented", nil
}

func (s *Service) Get(ctx context.Context, cid string) ([]byte, error) {
	s.log.InfoContext(ctx, "get blob called", "cid", cid)
	return nil, nil
}
