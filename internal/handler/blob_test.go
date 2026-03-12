package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"zephyr-storage/internal/ipfs"
	"zephyr-storage/internal/service"
)

// mockBlobService is a hand-rolled mock for the blobService interface.
type mockBlobService struct {
	uploadFn func(ctx context.Context, data []byte) (string, error)
	getFn    func(ctx context.Context, cid string) ([]byte, error)
}

func (m *mockBlobService) Upload(ctx context.Context, data []byte) (string, error) {
	return m.uploadFn(ctx, data)
}

func (m *mockBlobService) Get(ctx context.Context, cid string) ([]byte, error) {
	return m.getFn(ctx, cid)
}

// maxSizeMB used across handler tests (10 bytes for easy overflow testing).
const testMaxSizeB = 10

func newTestApp(svc blobService) *fiber.App {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Use testMaxSizeB directly by building handler with 0 MB and overriding.
	h := &Handler{svc: svc, maxSizeB: testMaxSizeB, log: log}

	app := fiber.New(fiber.Config{
		// Large enough so Fiber doesn't reject before our handler runs.
		BodyLimit: 1024 * 1024,
	})
	app.Post("/v1/blobs", h.UploadBlob)
	app.Get("/v1/blobs/:cid", h.GetBlob)
	return app
}

func doRequest(app *fiber.App, method, url string, body []byte, contentType string) *http.Response {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, url, reqBody)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		panic(err)
	}
	return resp
}

func readJSON(resp *http.Response) map[string]interface{} {
	var m map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&m)
	return m
}

// --- UploadBlob tests ---

func TestUploadBlob_Success(t *testing.T) {
	mock := &mockBlobService{
		uploadFn: func(_ context.Context, _ []byte) (string, error) {
			return "QmTestCID", nil
		},
	}
	app := newTestApp(mock)

	resp := doRequest(app, http.MethodPost, "/v1/blobs", []byte("hi"), "application/octet-stream")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readJSON(resp)
	if body["cid"] != "QmTestCID" {
		t.Fatalf("expected cid QmTestCID, got %v", body["cid"])
	}
}

func TestUploadBlob_WrongContentType(t *testing.T) {
	app := newTestApp(&mockBlobService{})

	resp := doRequest(app, http.MethodPost, "/v1/blobs", []byte("{}"), "application/json")

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", resp.StatusCode)
	}
	body := readJSON(resp)
	if body["error"] != "content-type must be application/octet-stream" {
		t.Fatalf("unexpected error: %v", body["error"])
	}
}

func TestUploadBlob_TooLarge(t *testing.T) {
	app := newTestApp(&mockBlobService{})

	// testMaxSizeB + 1 bytes to exceed handler limit
	data := make([]byte, testMaxSizeB+1)
	resp := doRequest(app, http.MethodPost, "/v1/blobs", data, "application/octet-stream")

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", resp.StatusCode)
	}
	body := readJSON(resp)
	if body["error"] != "payload too large" {
		t.Fatalf("unexpected error: %v", body["error"])
	}
}

func TestUploadBlob_ServiceError(t *testing.T) {
	mock := &mockBlobService{
		uploadFn: func(_ context.Context, _ []byte) (string, error) {
			return "", service.ErrEmptyData // any unexpected error maps to 500
		},
	}
	// Use a service that returns an unmapped error to trigger 500.
	errSvc := &mockBlobService{
		uploadFn: func(_ context.Context, _ []byte) (string, error) {
			return "", io.ErrUnexpectedEOF
		},
	}
	_ = mock
	app := newTestApp(errSvc)

	resp := doRequest(app, http.MethodPost, "/v1/blobs", []byte("x"), "application/octet-stream")

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	body := readJSON(resp)
	if body["error"] != "internal server error" {
		t.Fatalf("unexpected error: %v", body["error"])
	}
}

// --- GetBlob tests ---

func TestGetBlob_Success(t *testing.T) {
	payload := []byte("binary data")
	mock := &mockBlobService{
		getFn: func(_ context.Context, _ string) ([]byte, error) {
			return payload, nil
		},
	}
	app := newTestApp(mock)

	resp := doRequest(app, http.MethodGet, "/v1/blobs/QmTestCID", nil, "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("expected octet-stream, got %s", resp.Header.Get("Content-Type"))
	}
	got, _ := io.ReadAll(resp.Body)
	if !bytes.Equal(got, payload) {
		t.Fatalf("expected %q, got %q", payload, got)
	}
}

func TestGetBlob_NotFound(t *testing.T) {
	mock := &mockBlobService{
		getFn: func(_ context.Context, _ string) ([]byte, error) {
			return nil, ipfs.ErrNotFound
		},
	}
	app := newTestApp(mock)

	resp := doRequest(app, http.MethodGet, "/v1/blobs/QmTestCID", nil, "")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body := readJSON(resp)
	if body["error"] != "blob not found" {
		t.Fatalf("unexpected error: %v", body["error"])
	}
}

func TestGetBlob_InvalidCID(t *testing.T) {
	mock := &mockBlobService{
		getFn: func(_ context.Context, _ string) ([]byte, error) {
			return nil, service.ErrInvalidCID
		},
	}
	app := newTestApp(mock)

	resp := doRequest(app, http.MethodGet, "/v1/blobs/badcid", nil, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := readJSON(resp)
	if body["error"] != "invalid cid format" {
		t.Fatalf("unexpected error: %v", body["error"])
	}
}
