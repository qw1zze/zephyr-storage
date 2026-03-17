package service

import (
	"context"
	"errors"
	"testing"

	"zephyr-storage/internal/ipfs"
)

type mockIPFSClient struct {
	addFn func(ctx context.Context, data []byte) (string, error)
	getFn func(ctx context.Context, cid string) ([]byte, error)
}

func (m *mockIPFSClient) Add(ctx context.Context, data []byte) (string, error) {
	return m.addFn(ctx, data)
}

func (m *mockIPFSClient) Get(ctx context.Context, cid string) ([]byte, error) {
	return m.getFn(ctx, cid)
}

func newTestService(client ipfsClient, maxSizeMB int64) *Service {
	return NewService(client, maxSizeMB, noopLogger())
}

func TestUpload_Success(t *testing.T) {
	mock := &mockIPFSClient{
		addFn: func(_ context.Context, _ []byte) (string, error) {
			return "QmTestCID123456789012345678901234567890123456", nil
		},
	}
	svc := newTestService(mock, 10)

	cid, err := svc.Upload(context.Background(), []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cid != "QmTestCID123456789012345678901234567890123456" {
		t.Fatalf("unexpected cid: %s", cid)
	}
}

func TestUpload_EmptyData(t *testing.T) {
	svc := newTestService(nil, 10)

	_, err := svc.Upload(context.Background(), []byte{})
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestUpload_TooLarge(t *testing.T) {
	const maxSizeMB = 1
	svc := newTestService(nil, maxSizeMB)

	data := make([]byte, maxSizeMB*1024*1024+1)

	_, err := svc.Upload(context.Background(), data)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestUpload_IPFSError(t *testing.T) {
	ipfsErr := errors.New("kubo unreachable")
	mock := &mockIPFSClient{
		addFn: func(_ context.Context, _ []byte) (string, error) {
			return "", ipfsErr
		},
	}
	svc := newTestService(mock, 10)

	_, err := svc.Upload(context.Background(), []byte("data"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGet_Success(t *testing.T) {
	mock := &mockIPFSClient{
		getFn: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("data"), nil
		},
	}
	svc := newTestService(mock, 10)

	result, err := svc.Get(context.Background(), "QmTestCID0000000000000000000000000000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "data" {
		t.Fatalf("unexpected data: %q", result)
	}
}

func TestGet_EmptyCID(t *testing.T) {
	svc := newTestService(nil, 10)

	_, err := svc.Get(context.Background(), "")
	if !errors.Is(err, ErrEmptyCID) {
		t.Fatalf("expected ErrEmptyCID, got %v", err)
	}
}

func TestGet_InvalidCID(t *testing.T) {
	svc := newTestService(nil, 10)

	_, err := svc.Get(context.Background(), "invalidcid")
	if !errors.Is(err, ErrInvalidCID) {
		t.Fatalf("expected ErrInvalidCID, got %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	mock := &mockIPFSClient{
		getFn: func(_ context.Context, _ string) ([]byte, error) {
			return nil, ipfs.ErrNotFound
		},
	}
	svc := newTestService(mock, 10)

	_, err := svc.Get(context.Background(), "QmTestCID0000000000000000000000000000000000000")
	if !errors.Is(err, ipfs.ErrNotFound) {
		t.Fatalf("expected ipfs.ErrNotFound, got %v", err)
	}
}
