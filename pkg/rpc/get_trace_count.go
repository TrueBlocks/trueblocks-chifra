// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/rpc/query"
)

// traceGet calls trace_get with a flat index and returns true if a trace exists at that
// position. The index is 0-based into the pre-order traversal of sub-traces (excluding
// the root trace). Returns false when the index is past the end of the trace tree.
func (conn *Connection) traceGet(txHash string, index uint64) (bool, error) {
	method := "trace_get"
	params := query.Params{txHash, []string{fmt.Sprintf("0x%x", index)}}
	raw, err := query.Query[json.RawMessage](conn.Chain, method, params)
	if err != nil {
		return false, err
	}
	return raw != nil && string(*raw) != "null", nil
}

// HasNTraces returns true if the transaction has at least n traces (including the root).
// This requires a single trace_get RPC call regardless of the actual trace count, making
// it suitable for fast threshold checks (e.g., DDoS detection).
func (conn *Connection) HasNTraces(txHash string, n uint64) (bool, error) {
	if n <= 1 {
		// Every transaction has at least 1 trace (the root)
		return true, nil
	}
	// Flat index is 0-based and excludes the root, so the last valid index for
	// n total traces is n-2.
	return conn.traceGet(txHash, n-2)
}

// GetTracesCount returns the total number of traces in a transaction (including the root)
// using a binary search with trace_get. This avoids downloading the entire trace tree —
// roughly 2*log2(N) RPC calls instead of fetching all N traces.
func (conn *Connection) GetTracesCount(txHash string) (uint64, error) {
	// Every tx has at least 1 trace (the root). Check if there are at least 2.
	if has, err := conn.HasNTraces(txHash, 2); err != nil {
		return 0, err
	} else if !has {
		return 1, nil
	}

	// Find an upper bound by doubling
	upper := uint64(2)
	for {
		has, err := conn.HasNTraces(txHash, upper)
		if err != nil {
			return 0, err
		}
		if !has {
			break
		}
		upper *= 2
	}

	// Binary search between upper/2 and upper
	lo, hi := upper/2, upper
	for lo < hi {
		mid := (lo + hi + 1) / 2
		has, err := conn.HasNTraces(txHash, mid)
		if err != nil {
			return 0, err
		}
		if has {
			lo = mid
		} else {
			hi = mid - 1
		}
	}

	return lo, nil
}
