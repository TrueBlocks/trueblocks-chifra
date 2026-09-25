// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package index

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/progress"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/walk"
)

func TestWriteReaderToPathRejectsConcurrentTruncation(t *testing.T) {
	dir := t.TempDir()
	full := filepath.Join(dir, "000000001-000000002.bin")
	good := bytes.Repeat([]byte("a"), 16)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := writeReaderToPath(full, bytes.NewReader(good), 16, "good"); err != nil {
			t.Errorf("good write: %v", err)
		}
	}()
	var shortErr error
	go func() {
		defer wg.Done()
		shortErr = writeReaderToPath(full, bytes.NewReader([]byte("bbb")), 16, "short")
	}()
	wg.Wait()
	if !errors.Is(shortErr, ErrSizeMismatch) {
		t.Fatalf("expected ErrSizeMismatch for short write, got: %v", shortErr)
	}

	got, err := os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, good) {
		t.Fatalf("published %q", got)
	}
}

func TestWriteReaderToPathPublishesReadableFile(t *testing.T) {
	full := filepath.Join(t.TempDir(), "000000001-000000002.bin")
	if err := writeReaderToPath(full, bytes.NewReader([]byte("abc")), 3, "x"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(full)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Fatalf("mode=%v", info.Mode().Perm())
	}
}

func TestWriteReaderToPathMissingSize(t *testing.T) {
	err := writeReaderToPath(filepath.Join(t.TempDir(), "x.bin"), bytes.NewReader([]byte("abc")), 0, "x")
	if !errors.Is(err, ErrMissingSize) {
		t.Fatalf("err=%v", err)
	}
}

func TestDownloadChunkToDiscRetries(t *testing.T) {
	origDelay, origStall := downloadRetryDelay, downloadStallTimeout
	downloadRetryDelay = 5 * time.Millisecond
	downloadStallTimeout = 100 * time.Millisecond
	defer func() { downloadRetryDelay, downloadStallTimeout = origDelay, origStall }()

	finalBytes := bytes.Repeat([]byte("z"), 64)

	tests := []struct {
		name       string
		firstServe func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "body smaller than manifest size",
			firstServe: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(finalBytes[:32])
			},
		},
		{
			name: "body interrupted before Content-Length",
			firstServe: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(finalBytes)))
				w.WriteHeader(http.StatusOK)
				if flusher, ok := w.(http.Flusher); ok {
					_, _ = w.Write(finalBytes[:16])
					flusher.Flush()
				}
				// Close connection ungracefully via panic/hijack or by returning with fewer bytes
			},
		},
		{
			name: "gateway stalls mid-body",
			firstServe: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(finalBytes)))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(finalBytes[:16])
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var attempts int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				att := atomic.AddInt32(&attempts, 1)
				if att == 1 {
					tc.firstServe(w, r)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(finalBytes)
			}))
			defer server.Close()

			dir := t.TempDir()
			chainDir := filepath.Join(dir, "unchained", "testchain")
			t.Setenv("XDG_CACHE_HOME", dir)

			chunk := types.ChunkRecord{
				Range:     "000000001-000000002",
				IndexSize: int64(len(finalBytes)),
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := downloadChunkToDisc(ctx, "testchain", walk.Index_Final, chunk, server.URL, "fakehash", 3)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if atomic.LoadInt32(&attempts) < 2 {
				t.Fatalf("expected at least 2 attempts, got %d", attempts)
			}

			destPath := filepath.Join(chainDir, "finalized", chunk.Range+".bin")
			got, err := os.ReadFile(destPath)
			if err != nil {
				t.Fatalf("reading destination: %v", err)
			}
			if !bytes.Equal(got, finalBytes) {
				t.Fatalf("published %q, want %q", got, finalBytes)
			}
		})
	}
}

// Run in a subprocess because SIGINT is process-wide and sigintTrap's cleanup is shared.
func TestDownloadChunksInterruptBeforeHeaders(t *testing.T) {
	if os.Getenv("CHIFRA_TEST_DOWNLOAD_INTERRUPT") == "1" {
		http.DefaultClient = &http.Client{Transport: interruptTransport{}}
		updates := make(chan *progress.ProgressMsg)
		go func() {
			DownloadChunks("mainnet", []types.ChunkRecord{{Range: "000000001-000000002", IndexHash: "fakehash"}}, walk.Index_Final, 1, updates)
			close(updates)
		}()
		cancelled := false
		for event := range updates {
			if event.Event == progress.Cancelled {
				cancelled = true
			}
			if event.Event == progress.Error || event.Event == progress.AllDone {
				t.Fatalf("unexpected event after interrupt: %+v", event)
			}
		}
		if !cancelled {
			t.Fatal("missing cancellation event")
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDownloadChunksInterruptBeforeHeaders$")
	cmd.Env = append(os.Environ(), "CHIFRA_TEST_DOWNLOAD_INTERRUPT=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("download interrupt: %v\n%s", err, output)
	}
}

type interruptTransport struct{}

func (interruptTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		return nil, err
	}
	if err := process.Signal(os.Interrupt); err != nil {
		return nil, err
	}
	<-req.Context().Done()
	return nil, req.Context().Err()
}

func TestDownloadChunkToDiscStatusRetries(t *testing.T) {
	origDelay := downloadRetryDelay
	downloadRetryDelay = time.Millisecond
	defer func() { downloadRetryDelay = origDelay }()

	tests := []struct {
		status   int
		attempts int32
	}{
		{http.StatusNotFound, 1},
		{http.StatusTooManyRequests, 3},
		{http.StatusBadGateway, 3},
	}
	for _, tc := range tests {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			var attempts int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&attempts, 1)
				w.WriteHeader(tc.status)
			}))
			defer server.Close()
			t.Setenv("XDG_CACHE_HOME", t.TempDir())

			chunk := types.ChunkRecord{Range: "000000001-000000002", IndexSize: 64}
			err := downloadChunkToDisc(context.Background(), "testchain", walk.Index_Final, chunk, server.URL, "fakehash", 3)
			var statusErr *gatewayStatusError
			if !errors.As(err, &statusErr) || statusErr.code != tc.status {
				t.Fatalf("err=%v", err)
			}
			if got := atomic.LoadInt32(&attempts); got != tc.attempts {
				t.Fatalf("attempts=%d, want %d", got, tc.attempts)
			}
		})
	}
}
