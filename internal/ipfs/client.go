// Package ipfs provides a client for interacting with a Kubo IPFS node
// via its HTTP API and Gateway.
package ipfs

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Sentinel errors used across service and handler layers.
var (
	// ErrNotFound indicates the requested blob was not found in IPFS.
	ErrNotFound = errors.New("blob not found")
	// ErrTooLarge indicates the blob exceeds the configured size limit.
	ErrTooLarge = errors.New("blob exceeds size limit")
)

// Client communicates with a local Kubo IPFS node.
type Client struct {
	apiURL     string
	gatewayURL string
	httpClient *http.Client
	maxSizeB   int64
	logger     *slog.Logger
}

// New creates a new IPFS Client.
// apiURL is the Kubo HTTP API endpoint (e.g. http://localhost:5001).
// gatewayURL is the Kubo Gateway endpoint (e.g. http://localhost:8080).
// timeoutSec sets the HTTP client timeout in seconds.
// maxSizeMB sets the maximum allowed blob size in megabytes.
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
