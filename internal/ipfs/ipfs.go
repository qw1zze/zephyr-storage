package ipfs

import (
	"context"
	"log/slog"
)

func (c *Client) Add(ctx context.Context, data []byte) (string, error) {
	c.log.DebugContext(ctx, "ipfs add called", slog.Int("size", len(data)))
	return "", nil
}

func (c *Client) Get(ctx context.Context, cid string) ([]byte, error) {
	c.log.DebugContext(ctx, "ipfs get called", slog.String("cid", cid))
	return nil, nil
}
