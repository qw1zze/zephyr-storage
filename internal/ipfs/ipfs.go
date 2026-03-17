package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

type addResponse struct {
	Hash string `json:"Hash"`
}

func (c *Client) Add(ctx context.Context, data []byte) (string, error) {
	c.logger.DebugContext(ctx, "ipfs: add: starting", slog.Int("size", len(data)))

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("file", "blob")
	if err != nil {
		return "", fmt.Errorf("ipfs: add: create form file: %w", err)
	}

	if _, err = part.Write(data); err != nil {
		return "", fmt.Errorf("ipfs: add: write data: %w", err)
	}

	if err = w.Close(); err != nil {
		return "", fmt.Errorf("ipfs: add: close writer: %w", err)
	}

	url := c.apiURL + "/api/v0/add"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return "", fmt.Errorf("ipfs: add: new request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ipfs: add: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ipfs: add: unexpected status %d", resp.StatusCode)
	}

	var result addResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ipfs: add: decode response: %w", err)
	}

	if result.Hash == "" {
		return "", fmt.Errorf("ipfs: add: empty CID in response")
	}

	c.logger.DebugContext(ctx, "ipfs: add: success", slog.String("cid", result.Hash))
	return result.Hash, nil
}

func (c *Client) Get(ctx context.Context, cid string) ([]byte, error) {
	c.logger.DebugContext(ctx, "ipfs: get: starting", slog.String("cid", cid))

	url := c.gatewayURL + "/ipfs/" + cid

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("ipfs: get: new request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ipfs: get: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipfs: get: unexpected status %d", resp.StatusCode)
	}

	limit := c.maxSizeB + 1
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return nil, fmt.Errorf("ipfs: get: read body: %w", err)
	}

	if int64(len(data)) == limit {
		return nil, ErrTooLarge
	}

	c.logger.DebugContext(ctx, "ipfs: get: success",
		slog.String("cid", cid),
		slog.Int("size", len(data)),
	)
	return data, nil
}
