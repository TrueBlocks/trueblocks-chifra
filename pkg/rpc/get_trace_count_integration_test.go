//go:build integration
// +build integration

// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package rpc

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/utils"
)

func Test_traceGet(t *testing.T) {
	conn := TempConnection(utils.GetTestChain())

	// Block 46147 tx 0 is the first ever Ethereum transaction (simple ETH transfer, 1 trace).
	simpleTx := "0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060"

	// Index 0 should NOT exist for a 1-trace tx (the root is the only trace,
	// and flat indexing excludes the root).
	has, err := conn.traceGet(simpleTx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("expected no sub-trace at index 0 for a simple transfer")
	}

	// Block 2675055 tx 2: cleanup tx with 701 traces.
	cleanupTx := "0x884d0fc7b45b52494a9139e073fdffab0d462c9a24a6298c9c1c1cd90fd7a051"

	// Index 0 should exist.
	has, err = conn.traceGet(cleanupTx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("expected trace at index 0 for cleanup tx")
	}

	// Index 699 (the last sub-trace: 701 total - 1 root - 1 for 0-indexing = 699).
	has, err = conn.traceGet(cleanupTx, 699)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("expected trace at index 699 for cleanup tx")
	}

	// Index 700 should NOT exist.
	has, err = conn.traceGet(cleanupTx, 700)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("expected no trace at index 700 for cleanup tx (boundary)")
	}
}

func Test_HasNTraces(t *testing.T) {
	conn := TempConnection(utils.GetTestChain())

	simpleTx := "0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060"

	// Every tx has at least 1 trace.
	has, err := conn.HasNTraces(simpleTx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("every tx should have at least 1 trace")
	}

	// A simple transfer does NOT have 2 traces.
	has, err = conn.HasNTraces(simpleTx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("simple transfer should not have 2 traces")
	}

	// Cleanup tx has 701 traces.
	cleanupTx := "0x884d0fc7b45b52494a9139e073fdffab0d462c9a24a6298c9c1c1cd90fd7a051"

	has, err = conn.HasNTraces(cleanupTx, 701)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("cleanup tx should have at least 701 traces")
	}

	has, err = conn.HasNTraces(cleanupTx, 702)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("cleanup tx should NOT have 702 traces")
	}
}

func Test_GetTracesCount(t *testing.T) {
	conn := TempConnection(utils.GetTestChain())

	tests := []struct {
		name     string
		txHash   string
		expected uint64
	}{
		{
			name:     "simple transfer (1 trace)",
			txHash:   "0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060",
			expected: 1,
		},
		{
			name:     "cleanup tx (701 traces)",
			txHash:   "0x884d0fc7b45b52494a9139e073fdffab0d462c9a24a6298c9c1c1cd90fd7a051",
			expected: 701,
		},
		{
			name:     "DDoS attack tx (17122 traces)",
			txHash:   "0x0f5d5dbb913d6223f1e15306631d2f65400ccc36e64f5a4502e682c3b6f21a7e",
			expected: 17122,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cnt, err := conn.GetTracesCount(tt.txHash)
			if err != nil {
				t.Fatal(err)
			}
			if cnt != tt.expected {
				t.Errorf("GetTracesCount(%s) = %d, want %d", tt.name, cnt, tt.expected)
			}
		})
	}
}
