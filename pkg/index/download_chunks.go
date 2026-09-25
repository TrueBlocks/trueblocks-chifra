// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package index

// Fetching, unzipping, validating and saving both index and bloom chunks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/colors"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/config"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/debug"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/progress"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/sigintTrap"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/utils"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/walk"
	ants "github.com/panjf2000/ants/v2"
)

// jobResult carries a download body and the metadata needed to validate and save it.
type jobResult struct {
	rng      string
	fileSize int64
	contents io.Reader
	theChunk *types.ChunkRecord
}

type progressChan chan<- *progress.ProgressMsg

// Types of errors put into the progressChannel

var ErrUserHitControlC = errors.New("user hit control + c")
var ErrDownloadError = errors.New("download error")
var ErrSizeMismatch = errors.New("downloaded chunk size mismatch")
var ErrMissingSize = errors.New("missing expected chunk size")
var ErrDownloadStalled = errors.New("download stalled")

// WorkerArguments are types meant to hold worker function arguments. We cannot
// pass the arguments directly, because a worker function is expected to take one
// parameter of type interface{}.
type downloadWorkerArguments struct {
	ctx             context.Context
	progressChannel progressChan
	gatewayUrl      string
	downloadWg      *sync.WaitGroup
	nRetries        int
}

// worker function type as accepted by Ants
type workerFunction func(interface{})

// getDownloadWorker returns a worker function that downloads a chunk and writes it to disc
func getDownloadWorker(chain string, workerArgs downloadWorkerArguments, chunkType walk.CacheType) workerFunction {
	progressChannel := workerArgs.progressChannel

	return func(param interface{}) {
		chunk := param.(types.ChunkRecord)

		defer workerArgs.downloadWg.Done()

		select {
		case <-workerArgs.ctx.Done():
			return

		default:
			hash := chunk.BloomHash
			if chunkType == walk.Index_Final {
				hash = chunk.IndexHash
			}
			if hash == "" {
				return
			}

			// TODO: Do we really need the colored display?
			bHash := utils.FormattedHash(false, chunk.BloomHash.String())
			iHash := utils.FormattedHash(false, chunk.IndexHash.String())
			tHash := utils.FormattedHash(false, hash.String())
			msg := fmt.Sprintf("%s %s %s", chunk.Range, bHash, iHash)
			msg = strings.ReplaceAll(msg, tHash, colors.BrightCyan+tHash+colors.Off)
			progressChannel <- &progress.ProgressMsg{
				Payload: &chunk,
				Event:   progress.Start,
				Message: msg,
			}

			err := downloadChunkToDisc(workerArgs.ctx, chain, chunkType, chunk, workerArgs.gatewayUrl, hash.String(), workerArgs.nRetries)
			if errors.Is(workerArgs.ctx.Err(), context.Canceled) {
				return
			}
			if err != nil {
				progressChannel <- &progress.ProgressMsg{
					Payload: &chunk,
					Event:   progress.Error,
					Error:   fmt.Errorf("%w: %w", ErrDownloadError, err),
				}
				return
			}
			progressChannel <- &progress.ProgressMsg{
				Payload: &chunk,
				Event:   progress.Finished,
				Message: chunkType.String(),
			}
		}
	}
}

// fetchResult type make it easier to return both download content and
// download size information (for validation purposes)
type fetchResult struct {
	Body       io.ReadCloser
	ContentLen int64 // download size in bytes
}

func downloadChunkToDisc(ctx context.Context, chain string, chunkType walk.CacheType, chunk types.ChunkRecord, gateway, hash string, nRetries int) error {
	delay := downloadRetryDelay
	var lastErr error
	for attempt := 1; attempt <= nRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := downloadAttempt(ctx, chain, chunkType, chunk, gateway, hash)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retryableDownloadErr(lastErr) || attempt == nRetries {
			break
		}
		logger.Warn("Failed download", chunk.Range, "(will retry)", attempt, "of", nRetries)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		if delay > 0 && delay < 8*time.Second {
			delay *= 2
		}
	}
	return lastErr
}

// downloadAttempt fetches one chunk and publishes it. The stall timer runs only while
// waiting on the gateway (headers, then each body read), so a gateway that stops sending
// fails the attempt instead of hanging it, while slow disk writes are not counted.
func downloadAttempt(ctx context.Context, chain string, chunkType walk.CacheType, chunk types.ChunkRecord, gateway, hash string) error {
	attemptCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stall := time.AfterFunc(downloadStallTimeout, func() { cancel(ErrDownloadStalled) })
	defer stall.Stop()

	download, err := fetchFromIpfsGateway(attemptCtx, gateway, hash)
	stall.Stop()
	if err == nil {
		defer download.Body.Close()
		err = writeBytesToDisc(chain, chunkType, &jobResult{
			rng:      chunk.Range,
			fileSize: download.ContentLen,
			contents: &stallReader{r: download.Body, timer: stall, timeout: downloadStallTimeout},
			theChunk: &chunk,
		})
	}
	stalled := ctx.Err() == nil && errors.Is(context.Cause(attemptCtx), ErrDownloadStalled)
	if stalled && (errors.Is(err, context.Canceled) || errors.Is(err, ErrDownloadStalled)) {
		return fmt.Errorf("%w for %s: no data for %s", ErrDownloadStalled, chunk.Range, downloadStallTimeout)
	}
	return err
}

func retryableDownloadErr(err error) bool {
	var statusErr *gatewayStatusError
	if errors.As(err, &statusErr) {
		return !statusErr.permanent()
	}
	return err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, ErrUserHitControlC) &&
		!errors.Is(err, ErrMissingSize) &&
		!errors.Is(err, os.ErrPermission) &&
		!errors.Is(err, syscall.ENOSPC)
}

var downloadRetryDelay = time.Second
var downloadStallTimeout = time.Minute

// stallReader arms timer for the duration of each read so only an idle gateway trips it.
type stallReader struct {
	r       io.Reader
	timer   *time.Timer
	timeout time.Duration
}

func (s *stallReader) Read(p []byte) (int, error) {
	s.timer.Reset(s.timeout)
	n, err := s.r.Read(p)
	s.timer.Stop()
	return n, err
}

type gatewayStatusError struct {
	url  string
	code int
}

func (e *gatewayStatusError) Error() string {
	return fmt.Sprintf("fetchFromIpfsGateway %s returned status code: %d", e.url, e.code)
}

// permanent reports client errors (e.g. an unpinned CID) that retrying cannot fix.
func (e *gatewayStatusError) permanent() bool {
	return e.code >= 400 && e.code < 500 && e.code != http.StatusRequestTimeout && e.code != http.StatusTooManyRequests
}

// fetchFromIpfsGateway downloads a chunk from an IPFS gateway using HTTP
func fetchFromIpfsGateway(ctx context.Context, gateway, hash string) (*fetchResult, error) {
	url, _ := url.Parse(gateway)
	url.Path = path.Join(url.Path, hash)

	debug.DebugCurlStr(url.String())
	request, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("NewRequestWithContext %s returned error: %w", url, err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("DefaultClient.Do %s returned error: %w", url, err)
	}
	if response.StatusCode != 200 {
		response.Body.Close()
		return nil, &gatewayStatusError{url: url.String(), code: response.StatusCode}
	}

	return &fetchResult{
		Body:       response.Body,
		ContentLen: response.ContentLength,
	}, nil
}

// DownloadChunks downloads, unzips and saves the chunk of type indicated by chunkType
// for each chunk in chunks. ProgressMsg is reported to progressChannel.
func DownloadChunks(chain string, chunksToDownload []types.ChunkRecord, chunkType walk.CacheType, poolSize int, progressChannel progressChan) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	trapChannel := sigintTrap.Enable(ctx, cancel, func() { logger.Warn(sigintTrap.TrapMessage) })
	defer sigintTrap.Disable(trapChannel)

	var downloadWg sync.WaitGroup
	downloadWorkerArgs := downloadWorkerArguments{
		ctx:             ctx,
		progressChannel: progressChannel,
		downloadWg:      &downloadWg,
		gatewayUrl:      config.GetChain(chain).IpfsGateway,
		nRetries:        8,
	}
	downloadPool, err := ants.NewPoolWithFunc(poolSize, getDownloadWorker(chain, downloadWorkerArgs, chunkType))
	defer downloadPool.Release()
	if err != nil {
		logger.Panic(err)
	}

	for _, chunk := range chunksToDownload {
		downloadWg.Add(1)
		_ = downloadPool.Invoke(chunk)
	}
	downloadWg.Wait()

	if errors.Is(ctx.Err(), context.Canceled) {
		progressChannel <- &progress.ProgressMsg{
			Event: progress.Cancelled,
		}
		return
	}

	progressChannel <- &progress.ProgressMsg{
		Event: progress.AllDone,
	}
}

// writeBytesToDisc save the downloaded bytes to disc
func writeBytesToDisc(chain string, chunkType walk.CacheType, res *jobResult) error {
	fullPath := filepath.Join(config.PathToIndex(chain), "finalized", res.rng+".bin")
	if chunkType == walk.Index_Bloom {
		fullPath = ToBloomPath(fullPath)
	}
	return writeReaderToPath(fullPath, res.contents, expectedChunkSize(chunkType, res), res.rng)
}

func writeReaderToPath(fullPath string, contents io.Reader, expected int64, rng string) error {
	if expected <= 0 {
		return fmt.Errorf("%w for %s", ErrMissingSize, rng)
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	return writeFileAtomic(fullPath, func(w io.Writer) error {
		written, err := io.Copy(w, contents)
		if err != nil {
			// Information about this error
			// https://community.k6.io/t/warn-0040-request-failed-error-stream-error-stream-id-3-internal-error/777/2
			return fmt.Errorf("error copying %s file in writeBytesToDisc: [%w]", rng, err)
		}
		if written != expected {
			return fmt.Errorf("%w for %s: wrote %d, expected %d", ErrSizeMismatch, rng, written, expected)
		}
		return nil
	})
}

func expectedChunkSize(chunkType walk.CacheType, res *jobResult) int64 {
	if chunkType == walk.Index_Bloom && res.theChunk.BloomSize > 0 {
		return res.theChunk.BloomSize
	}
	if chunkType != walk.Index_Bloom && res.theChunk.IndexSize > 0 {
		return res.theChunk.IndexSize
	}
	return res.fileSize
}
