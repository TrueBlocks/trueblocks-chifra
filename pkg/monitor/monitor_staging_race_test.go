// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The scraper removes the staging file and then writes its successor in
// place. readStaging must wait for a whole, unchanged file.

func stagingBody(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "%s\t%09d\t%05d\n", "0xfb18da3bca8057eb05aab6e77a32925949093533", 28131489+i, i)
	}
	return b.String()
}

func fastRetries(t *testing.T, attempts int) {
	a, d := stagingAttempts, stagingRetryDelay
	stagingAttempts, stagingRetryDelay = attempts, 20*time.Millisecond
	t.Cleanup(func() { stagingAttempts, stagingRetryDelay = a, d })
}

func TestReadStagingWholeFile(t *testing.T) {
	fastRetries(t, 5)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "028131489-028131491.txt"), []byte(stagingBody(3)), 0o644); err != nil {
		t.Fatal(err)
	}
	rng, lines, ok := readStaging(dir)
	if !ok || len(lines) != 3 || rng.First != 28131489 || rng.Last != 28131491 {
		t.Fatalf("want 3 lines over 28131489-28131491, got ok=%v n=%d rng=%v", ok, len(lines), rng)
	}
}

// The folder is empty while the scraper rewrites: the read waits for the new file.
func TestReadStagingWaitsForNewFile(t *testing.T) {
	fastRetries(t, 30)
	dir := t.TempDir()
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(dir, "028131489-028131492.txt"), []byte(stagingBody(4)), 0o644)
	}()
	_, lines, ok := readStaging(dir)
	if !ok || len(lines) != 4 {
		t.Fatalf("want 4 lines after the file appears, got ok=%v n=%d", ok, len(lines))
	}
}

// A half-written file (not a whole number of lines) is never served.
func TestReadStagingSkipsPartialFile(t *testing.T) {
	fastRetries(t, 30)
	dir := t.TempDir()
	fn := filepath.Join(dir, "028131489-028131492.txt")
	full := stagingBody(4)
	if err := os.WriteFile(fn, []byte(full[:len(full)-30]), 0o644); err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = os.WriteFile(fn, []byte(full), 0o644)
	}()
	_, lines, ok := readStaging(dir)
	if !ok || len(lines) != 4 {
		t.Fatalf("want the whole file (4 lines), got ok=%v n=%d", ok, len(lines))
	}
}

// Control: with no file at all, it gives up and reports not ok. The caller
// then leaves every LastScanned as it was.
func TestReadStagingGivesUp(t *testing.T) {
	fastRetries(t, 3)
	if _, lines, ok := readStaging(t.TempDir()); ok || lines != nil {
		t.Fatalf("want not ok with no lines, got ok=%v n=%d", ok, len(lines))
	}
}
