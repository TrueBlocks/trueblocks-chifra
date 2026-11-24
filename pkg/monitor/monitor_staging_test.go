// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package monitor

import (
	"fmt"
	"testing"
)

// stagingLine builds one line of a staging file: address, 9-digit block,
// 5-digit transaction index, tab separated.
func stagingLine(addr string, bn, tx uint32) string {
	return fmt.Sprintf("%s\t%09d\t%05d", addr, bn, tx)
}

// lastVisited is the monitor's LastScanned: the next block to scan. A record
// at that block has never been read, so getAppearances must keep it.
func TestGetAppearancesKeepsLastScannedBlock(t *testing.T) {
	addr := "0xfb18da3bca8057eb05aab6e77a32925949093533"
	other := "0xfb18da3bca8057eb05aab6e77a32925949093534"
	lines := []string{
		stagingLine(addr, 28131488, 5), // already scanned (in the finalized chunk)
		stagingLine(addr, 28131489, 170),
		stagingLine(addr, 28131489, 171),
		stagingLine(addr, 28131489, 172),
		stagingLine(addr, 28131490, 3),
		stagingLine(other, 28131489, 1), // a different address stops the scan
	}
	got := getAppearances(addr, lines, 28131489, 0)
	if got == nil {
		t.Fatal("want 4 records, got nil")
	}
	want := [][2]uint32{{28131489, 170}, {28131489, 171}, {28131489, 172}, {28131490, 3}}
	if len(*got) != len(want) {
		t.Fatalf("want %d records, got %d: %v", len(want), len(*got), *got)
	}
	for i, w := range want {
		r := (*got)[i]
		if r.BlockNumber != w[0] || r.TransactionIndex != w[1] {
			t.Errorf("record %d: want %d.%d, got %d.%d", i, w[0], w[1], r.BlockNumber, r.TransactionIndex)
		}
	}
}

// Control: records before lastVisited were already scanned and stay out.
func TestGetAppearancesSkipsScannedBlocks(t *testing.T) {
	addr := "0xfb18da3bca8057eb05aab6e77a32925949093533"
	lines := []string{stagingLine(addr, 100, 1), stagingLine(addr, 101, 2)}
	if got := getAppearances(addr, lines, 102, 0); got != nil {
		t.Fatalf("want nil, got %v", *got)
	}
}
