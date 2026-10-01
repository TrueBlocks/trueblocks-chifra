package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/cache"
)

func TestTraceUnmarshalReward(t *testing.T) {
	// Reth's trace_block(1) response has no transaction for the block reward.
	data := []byte(`[{"action":{"author":"0x05a56e2d52c817161883f50c441c3228cfe54d9f","rewardType":"block","value":"0x4563918244f40000"},"blockHash":"0x88e96d4537bea4d9c05d12549907b32561d3bf31f45aae734cdc119f13406cb6","blockNumber":1,"result":null,"subtraces":0,"traceAddress":[],"transactionHash":null,"transactionPosition":null,"type":"reward"}]`)
	var traces []Trace
	if err := json.Unmarshal(data, &traces); err != nil {
		t.Fatal(err)
	}
	if len(traces) != 1 {
		t.Fatalf("got %d traces, want 1", len(traces))
	}
	trace := traces[0]
	if trace.TraceType != "reward" || trace.BlockNumber != 1 || !trace.TransactionHash.IsZero() {
		t.Fatalf("unexpected reward trace: %+v", trace)
	}
	if trace.BlockHash != base.HexToHash("0x88e96d4537bea4d9c05d12549907b32561d3bf31f45aae734cdc119f13406cb6") {
		t.Fatalf("unexpected block hash: %s", trace.BlockHash)
	}
}

func TestTraceCache(t *testing.T) {
	expected := &TraceGroup{
		BlockNumber:      17432262,
		TransactionIndex: 44,
		Traces: []Trace{
			{
				Action: &TraceAction{
					CallType: "call",
					From:     base.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b"),
					Gas:      39927,
					Input:    "0x1fee5cd200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000a7365706f6c69612d747300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002e516d5441623866625233674d6f777876624e466f584339337a6762386f66395567756d3363554b564366414e3172000000000000000000000000000000000000",
					To:       base.HexToAddress("0x0c316b7042b419d07d343f2f4f5bd54ff731183d"),
				},
				BlockHash:   base.HexToHash("0x0bee6d19dab1ce5ddc296a83da21097c902e8d32f0d8c0b6ffad19b9bcffcd67"),
				BlockNumber: 17432262,
				Result: &TraceResult{
					GasUsed: 19058,
				},
				Timestamp:        1686186779,
				TransactionHash:  base.HexToHash("0x7b0dd622b0de6448937d564be16e08fb885895383391b890448cd284ce33f993"),
				TransactionIndex: 44,
				TraceType:        "call",
			},
		},
	}

	store, err := cache.NewStore(&cache.StoreOptions{Location: cache.MemoryCache})
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Write(expected); err != nil {
		t.Fatal(err)
	}

	readBack := &TraceGroup{
		BlockNumber:      17432262,
		TransactionIndex: 44,
	}
	if err := store.ReadFromStore(readBack); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, readBack) {
		t.Fatalf("value mismatch: got %+v want %+v\n", readBack, expected)
	}
}

func TestTraceCacheArticulated(t *testing.T) {
	expected := &TraceGroup{
		BlockNumber:      17432262,
		TransactionIndex: 44,
		Traces: []Trace{
			{
				Action: &TraceAction{
					CallType: "call",
					From:     base.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b"),
					Gas:      39927,
					Input:    "0x1fee5cd200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000a7365706f6c69612d747300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002e516d5441623866625233674d6f777876624e466f584339337a6762386f66395567756d3363554b564366414e3172000000000000000000000000000000000000",
					To:       base.HexToAddress("0x0c316b7042b419d07d343f2f4f5bd54ff731183d"),
				},
				ArticulatedTrace: &Function{
					Inputs: []Parameter{
						{
							Name:  "chain",
							Value: "sepolia-ts",
						},
						{
							Name:  "hash",
							Value: "QmTAb8fbR3gMowxvbNFoXC93zgb8of9Ugum3cUKVCfAN1r",
						},
					},
					Name: "publishHash",
				},
				BlockHash:   base.HexToHash("0x0bee6d19dab1ce5ddc296a83da21097c902e8d32f0d8c0b6ffad19b9bcffcd67"),
				BlockNumber: 17432262,
				Result: &TraceResult{
					GasUsed: 19058,
				},
				Timestamp:        1686186779,
				TransactionHash:  base.HexToHash("0x7b0dd622b0de6448937d564be16e08fb885895383391b890448cd284ce33f993"),
				TransactionIndex: 44,
				TraceType:        "call",
			},
		},
	}

	store, err := cache.NewStore(&cache.StoreOptions{Location: cache.MemoryCache})
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Write(expected); err != nil {
		t.Fatal(err)
	}

	// Read
	readBack := &TraceGroup{
		BlockNumber:      17432262,
		TransactionIndex: 44,
	}
	if err := store.ReadFromStore(readBack); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, readBack) {
		t.Fatalf("value mismatch: got %+v want %+v\n", readBack, expected)
	}
}
