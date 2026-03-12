package ipfs

import "log/slog"

type Client struct {
	apiURL     string
	gatewayURL string
	log        *slog.Logger
}

func NewClient(apiURL, gatewayURL string, log *slog.Logger) *Client {
	return &Client{
		apiURL:     apiURL,
		gatewayURL: gatewayURL,
		log:        log,
	}
}
