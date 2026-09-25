package index

// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

import (
	"fmt"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/colors"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/manifest"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/progress"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/ranges"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/walk"
)

// DownloadOneChunk a filename to an index portion, finds the corresponding CID (hash)
// entry in the manifest, and downloads the index chunk to the local drive
func DownloadOneChunk(chain string, man *manifest.Manifest, fileRange ranges.FileRange) error {
	// Find bloom filter's CID
	matchedPin := man.ChunkMap[fileRange.String()]
	if matchedPin == nil || matchedPin.Range == "" {
		return fmt.Errorf("filename path missing in chunks: %s", fileRange)
	}

	logger.Info("Bloom filter hit, downloading index portion", (colors.Yellow + fileRange.String() + colors.Off), "from IPFS.")

	// Start downloading the filter
	chunks := []types.ChunkRecord{*matchedPin}
	progressChannel := make(chan *progress.ProgressMsg)

	go func() {
		DownloadChunks(chain, chunks, walk.Index_Final, 4 /* poolSize */, progressChannel)
		close(progressChannel)
	}()

	var firstErr error
	for event := range progressChannel {
		switch event.Event {
		case progress.AllDone:
			// completed normally
		case progress.Cancelled:
			if firstErr == nil {
				firstErr = ErrUserHitControlC
			}
		case progress.Error:
			if firstErr == nil {
				if event.Error != nil {
					firstErr = fmt.Errorf("error while downloading: %w", event.Error)
				} else if event.Message != "" {
					firstErr = fmt.Errorf("error while downloading: %s", event.Message)
				} else {
					firstErr = fmt.Errorf("error while downloading")
				}
			}
		}
	}
	return firstErr
}
