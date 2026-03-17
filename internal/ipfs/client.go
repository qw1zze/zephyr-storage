package ipfs

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
)

var (
	ErrNotFound = errors.New("blob not found")
	ErrTooLarge = errors.New("blob exceeds size limit")
)

type Client struct {
	apiURL     string
	gatewayURL string
	httpClient *http.Client
	maxSizeB   int64
	logger     *slog.Logger
}

func New(
	apiURL string,
	gatewayURL string,
	timeoutSec int,
	maxSizeMB int64,
	logger *slog.Logger,
) *Client {
	return &Client{
		apiURL:     apiURL,
		gatewayURL: gatewayURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
		maxSizeB: maxSizeMB * 1024 * 1024,
		logger:   logger,
	}
}
