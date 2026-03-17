package ipfs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAdd_Success(t *testing.T) {
	var gotContentType string
	var gotPath string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")

		mediaType, params, err := mime.ParseMediaType(gotContentType)
		if err != nil {
			t.Fatalf("parse media type: %v", err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("expected multipart/form-data, got %s", mediaType)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatalf("next part: %v", err)
		}
		gotBody, err = io.ReadAll(part)
		if err != nil {
			t.Fatalf("read part: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"Name":"blob","Hash":"QmTestCID","Size":"5"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.URL, 5, 10, newTestLogger())

	cid, err := c.Add(context.Background(), []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cid != "QmTestCID" {
		t.Fatalf("expected QmTestCID, got %s", cid)
	}
	if gotPath != "/api/v0/add" {
		t.Fatalf("expected path /api/v0/add, got %s", gotPath)
	}
	if !strings.Contains(gotContentType, "multipart/form-data") {
		t.Fatalf("expected multipart/form-data, got %s", gotContentType)
	}
	if !bytes.Equal(gotBody, []byte("hello")) {
		t.Fatalf("expected body 'hello', got %q", gotBody)
	}
}

func TestAdd_KuboUnavailable(t *testing.T) {
	c := New("http://127.0.0.1:1", "", 0, 10, newTestLogger())
	c.httpClient.Timeout = 1 * time.Millisecond

	_, err := c.Add(context.Background(), []byte("data"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdd_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.URL, 5, 10, newTestLogger())

	_, err := c.Add(context.Background(), []byte("data"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdd_EmptyHash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"Hash":""}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.URL, 5, 10, newTestLogger())

	_, err := c.Add(context.Background(), []byte("data"))
	if err == nil {
		t.Fatal("expected error for empty hash, got nil")
	}
}

func TestGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ipfs/QmTest" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.URL, 5, 10, newTestLogger())

	data, err := c.Get(context.Background(), "QmTest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(data, []byte("hello")) {
		t.Fatalf("expected 'hello', got %q", data)
	}
}

func TestGet_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := New(srv.URL, srv.URL, 5, 10, newTestLogger())

	_, err := c.Get(context.Background(), "QmMissing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_TooLarge(t *testing.T) {
	bigData := make([]byte, 2*1024*1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bigData)
	}))
	defer srv.Close()

	c := New(srv.URL, srv.URL, 5, 1, newTestLogger())

	_, err := c.Get(context.Background(), "QmBig")
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestGet_KuboUnavailable(t *testing.T) {
	c := New("", "http://127.0.0.1:1", 0, 10, newTestLogger())
	c.httpClient.Timeout = 1 * time.Millisecond

	_, err := c.Get(context.Background(), "QmTest")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
