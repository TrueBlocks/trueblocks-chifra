package rpc

// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

import (
	"fmt"
	"os"
	"testing"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
)

func TestEstimateGasAndPrice(t *testing.T) {
	// os.Setenv("TB_DEBUG_CURL", "true")
	fmt.Println("TB_DEBUG_CURL", os.Getenv("TB_DEBUG_CURL"))

	// Simple ETH transfer between two EOA addresses
	// Using well-known addresses that can receive ETH
	from := base.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := base.HexToAddress("0x1234567890123456789012345678901234567890") // Simple EOA address

	estimatedGas, gasPrice, err := EstimateGasAndPrice(
		"mainnet",
		from,
		to,
		"0x", // empty data for simple transfer
		base.NewWei(0),
	)

	if err != nil {
		t.Fatalf("EstimateGasAndPrice failed: %v", err)
	}

	// Basic sanity checks
	if estimatedGas == 0 {
		t.Error("Expected non-zero gas estimate")
	}

	if gasPrice == 0 {
		t.Error("Expected non-zero gas price")
	}

	// Gas for simple transfer should be reasonable (21000 - 100000)
	if estimatedGas < 21000 || estimatedGas > 100000 {
		t.Errorf("Gas estimate %d seems unreasonable for simple transfer", estimatedGas)
	}

	t.Logf("Estimated gas: %d", estimatedGas)
	t.Logf("Gas price: %d wei", gasPrice)
}

func TestEstimateGasAndPrice_WithValue(t *testing.T) {
	from := base.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b")
	to := base.HexToAddress("0x1234567890123456789012345678901234567890") // Simple EOA address
	value := base.NewWeiStr("1000000000000000")                           // 0.001 ETH (small amount less likely to fail)

	estimatedGas, gasPrice, err := EstimateGasAndPrice(
		"mainnet",
		from,
		to,
		"0x",
		value,
	)

	if err != nil {
		t.Fatalf("EstimateGasAndPrice with value failed: %v", err)
	}

	if estimatedGas == 0 {
		t.Error("Expected non-zero gas estimate")
	}

	if gasPrice == 0 {
		t.Error("Expected non-zero gas price")
	}

	// Gas for simple transfer should be 21000 regardless of value
	if estimatedGas != 21000 {
		t.Logf("Warning: Expected 21000 gas for simple transfer, got %d", estimatedGas)
	}

	t.Logf("Estimated gas (with value): %d", estimatedGas)
	t.Logf("Gas price: %d wei", gasPrice)
}

func TestEstimateGasAndPrice_WithData(t *testing.T) {
	// Test calling a view function (balanceOf) which won't revert
	from := base.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b")
	tokenContract := base.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48") // USDC

	// balanceOf(address) - view function that won't revert
	// Function selector: 0x70a08231
	// Address parameter: 0x000000000000000000000000f503017d7baf7fbc0fff7492b751025c6a78179b (padded)
	data := "0x70a08231000000000000000000000000f503017d7baf7fbc0fff7492b751025c6a78179b"

	estimatedGas, gasPrice, err := EstimateGasAndPrice(
		"mainnet",
		from,
		tokenContract,
		data,
		base.NewWei(0),
	)

	if err != nil {
		t.Fatalf("EstimateGasAndPrice with data failed: %v", err)
	}

	if estimatedGas == 0 {
		t.Error("Expected non-zero gas estimate")
	}

	// Contract call with data should cost more than 21000
	if estimatedGas < 21000 {
		t.Errorf("Gas estimate %d seems too low for contract call", estimatedGas)
	}

	t.Logf("Estimated gas (contract call): %d", estimatedGas)
	t.Logf("Gas price: %d wei", gasPrice)
}

func TestEstimateGasAndPrice_ZeroAddresses(t *testing.T) {
	// Should handle zero addresses gracefully
	estimatedGas, gasPrice, err := EstimateGasAndPrice(
		"mainnet",
		base.ZeroAddr,
		base.ZeroAddr,
		"0x",
		base.NewWei(0),
	)

	// This might error or return zeros depending on RPC behavior
	// The test documents the expected behavior
	if err != nil {
		t.Logf("Zero addresses returned error (expected): %v", err)
		return
	}

	t.Logf("Zero addresses: gas=%d, price=%d", estimatedGas, gasPrice)
}

func TestEstimateGasAndPrice_InvalidChain(t *testing.T) {
	from := base.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := base.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

	_, _, err := EstimateGasAndPrice(
		"nonexistent-chain",
		from,
		to,
		"0x",
		base.NewWei(0),
	)

	if err == nil {
		t.Error("Expected error for invalid chain, got nil")
	}

	t.Logf("Invalid chain error (expected): %v", err)
}
