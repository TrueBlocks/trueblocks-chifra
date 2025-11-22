//go:build integration
// +build integration

package rpc

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/utils"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/walk"
)

func TestTokenOptimization(t *testing.T) {
	fmt.Println("=== TrueBlocks Token Optimization Test Suite ===")

	// Test 1: Token Cache Location Generation
	t.Run("TokenCacheLocations", func(t *testing.T) {
		testTokenCacheLocations(t)
	})

	// Test 2: Approval Cache Location Generation
	t.Run("ApprovalCacheLocations", func(t *testing.T) {
		testApprovalCacheLocations(t)
	})

	// Test 3: RPC Connection Setup
	t.Run("RpcConnection", func(t *testing.T) {
		testRpcConnection(t)
	})

	// Test 4: Cache File Path Verification
	t.Run("CacheFilePaths", func(t *testing.T) {
		testCacheFilePaths(t)
	})

	// Test 5: Token Balance with Cache
	t.Run("TokenBalanceWithCache", func(t *testing.T) {
		testTokenBalanceWithCache(t)
	})

	// Test 6: Approval Allowance with Cache
	t.Run("ApprovalAllowanceWithCache", func(t *testing.T) {
		testApprovalAllowanceWithCache(t)
	})
}

func testTokenCacheLocations(t *testing.T) {
	token := &types.Token{
		Holder:      base.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b"),
		Address:     base.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
		BlockNumber: 17000000,
	}

	dir, id, ext := token.CacheLocations()
	t.Logf("Token cache: directory=%s, id=%s, ext=%s", dir, id, ext)

	expectedDir := "tokens/f5/03/01"
	if dir != expectedDir {
		t.Errorf("Token cache directory incorrect: got %s, expected %s", dir, expectedDir)
	}

	if ext != "bin" {
		t.Errorf("Token cache extension incorrect: got %s, expected bin", ext)
	}
}

func testApprovalCacheLocations(t *testing.T) {
	approval := &types.Approval{
		Owner:       base.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b"),
		Token:       base.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
		Spender:     base.HexToAddress("0x7a250d5630b4cf539739df2c5dacb4c659f2488d"),
		BlockNumber: 17000000,
	}

	dir, id, ext := approval.CacheLocations()
	t.Logf("Approval cache: directory=%s, id=%s, ext=%s", dir, id, ext)

	expectedDir := "approvals/f5/03/01"
	if dir != expectedDir {
		t.Errorf("Approval cache directory incorrect: got %s, expected %s", dir, expectedDir)
	}

	if ext != "bin" {
		t.Errorf("Approval cache extension incorrect: got %s, expected bin", ext)
	}
}

func testRpcConnection(t *testing.T) {
	chain := utils.GetTestChain()
	caches := map[walk.CacheType]bool{
		walk.Cache_Tokens:    true,
		walk.Cache_Approvals: true,
	}

	conn := TestConnection(chain, true, caches)
	if conn == nil {
		t.Fatal("Failed to create RPC connection")
	}

	if conn.Chain != chain {
		t.Errorf("Chain mismatch: got %s, expected %s", conn.Chain, chain)
	}

	t.Logf("RPC connection created successfully for chain: %s", conn.Chain)
}

func testCacheFilePaths(t *testing.T) {
	chain := utils.GetTestChain()

	// Test token cache path
	tokenPath := walk.GetRootPathFromCacheType(chain, walk.Cache_Tokens)
	t.Logf("Token cache root path: %s", tokenPath)

	// Test approval cache path
	approvalPath := walk.GetRootPathFromCacheType(chain, walk.Cache_Approvals)
	t.Logf("Approval cache root path: %s", approvalPath)

	// Expected paths should contain 'tokens' and 'approvals'
	if filepath.Base(tokenPath) != "tokens" {
		t.Errorf("Token cache path should end with 'tokens', got: %s", tokenPath)
	}

	if filepath.Base(approvalPath) != "approvals" {
		t.Errorf("Approval cache path should end with 'approvals', got: %s", approvalPath)
	}
}

func testTokenBalanceWithCache(t *testing.T) {
	chain := utils.GetTestChain()
	conn := TempConnection(chain)

	// DAI token
	tokenAddress := base.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
	holderAddress := base.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b")
	blockNumber := base.Blknum(17000000)

	// First call should populate cache
	balance1, err := conn.GetBalanceAtToken(tokenAddress, holderAddress, blockNumber)
	if err != nil {
		t.Fatal("First balance call failed:", err)
	}

	// Second call should use cache
	balance2, err := conn.GetBalanceAtToken(tokenAddress, holderAddress, blockNumber)
	if err != nil {
		t.Fatal("Second balance call failed:", err)
	}

	if balance1.Cmp(balance2) != 0 {
		t.Errorf("Balance mismatch: first=%s, second=%s", balance1.String(), balance2.String())
	}

	t.Logf("Token balance test passed: %s", balance1.String())
}

func testApprovalAllowanceWithCache(t *testing.T) {
	chain := utils.GetTestChain()
	conn := TempConnection(chain)

	// DAI token
	tokenAddress := base.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
	ownerAddress := base.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b")
	spenderAddress := base.HexToAddress("0x7a250d5630b4cf539739df2c5dacb4c659f2488d")
	blockNumber := base.Blknum(17000000)

	// First call should populate cache
	allowance1, err := conn.GetAllowanceAtToken(tokenAddress, ownerAddress, spenderAddress, blockNumber)
	if err != nil {
		t.Fatal("First allowance call failed:", err)
	}

	// Second call should use cache
	allowance2, err := conn.GetAllowanceAtToken(tokenAddress, ownerAddress, spenderAddress, blockNumber)
	if err != nil {
		t.Fatal("Second allowance call failed:", err)
	}

	if allowance1.Cmp(allowance2) != 0 {
		t.Errorf("Allowance mismatch: first=%s, second=%s", allowance1.String(), allowance2.String())
	}

	t.Logf("Approval allowance test passed: %s", allowance1.String())
}
