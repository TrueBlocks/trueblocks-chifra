package decache

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/cache"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/file"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/identifiers"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/rpc"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/walk"
)

// LocationsFromTokensByPattern finds and generates cache locations for token balance and approval caches
// by scanning cache directories for files matching the provided address patterns.
// This approach is more flexible than requiring exact cache keys.
//
// For tokens cache: finds files where the holder address matches any provided address
// For approvals cache: finds files where the owner address matches any provided address
func LocationsFromTokensByPattern(conn *rpc.Connection, addresses []string, scanMode string) ([]cache.Locator, error) {
	locations := make([]cache.Locator, 0)

	if len(addresses) == 0 {
		return locations, nil
	}

	// Convert to lowercase hex for pattern matching
	addrPatterns := make([]string, len(addresses))
	for i, addr := range addresses {
		addrPatterns[i] = strings.ToLower(base.HexToAddress(addr).String())
	}

	// Get cache root paths
	tokensPath := walk.GetRootPathFromCacheType(conn.Chain, walk.Cache_Tokens)
	approvalsPath := walk.GetRootPathFromCacheType(conn.Chain, walk.Cache_Approvals)

	// Scan tokens cache for matching files (unless approvals-only mode)
	if scanMode != "approvals-only" && file.FolderExists(tokensPath) {
		tokenFiles, err := scanCacheForAddressPattern(tokensPath, addrPatterns, false)
		if err != nil {
			return nil, fmt.Errorf("scanning tokens cache: %w", err)
		}
		locations = append(locations, tokenFiles...)
	}

	// Scan approvals cache for matching files (if requested or in both mode)
	if (scanMode == "approvals-only" || scanMode == "both") && file.FolderExists(approvalsPath) {
		approvalFiles, err := scanCacheForAddressPattern(approvalsPath, addrPatterns, true)
		if err != nil {
			return nil, fmt.Errorf("scanning approvals cache: %w", err)
		}
		locations = append(locations, approvalFiles...)
	}

	return locations, nil
}

// scanCacheForAddressPattern recursively scans cache directory for files containing address patterns
func scanCacheForAddressPattern(cachePath string, addrPatterns []string, isApprovals bool) ([]cache.Locator, error) {
	locations := make([]cache.Locator, 0)

	// For efficiency, we could scan only directories that match the first few chars of addresses,
	// but for now we'll do a full walk since addresses can start with many different patterns
	err := filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue scanning
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".bin") {
			// Parse cache filename to extract addresses and block number
			if locator := parseCacheFilename(info.Name(), isApprovals, addrPatterns); locator != nil {
				locations = append(locations, locator)
			}
		}
		return nil
	})

	return locations, err
}

// parseCacheFilename parses cache filenames and returns a locator if the address pattern matches
func parseCacheFilename(filename string, isApprovals bool, addrPatterns []string) cache.Locator {
	// Remove .bin extension
	name := strings.TrimSuffix(filename, ".bin")
	parts := strings.Split(name, "-")

	if isApprovals {
		// Approvals format: owner-token-spender-blocknum
		if len(parts) >= 4 {
			ownerAddr := parts[0]
			tokenAddr := parts[1]
			spenderAddr := parts[2]

			// Check if owner address matches any pattern (exact match)
			for _, pattern := range addrPatterns {
				cleanPattern := strings.TrimPrefix(strings.ToLower(pattern), "0x")
				if ownerAddr == cleanPattern {
					if blockNum, err := strconv.ParseUint(parts[3], 10, 64); err == nil {
						return &types.Approval{
							Owner:       base.HexToAddress("0x" + ownerAddr),
							Token:       base.HexToAddress("0x" + tokenAddr),
							Spender:     base.HexToAddress("0x" + spenderAddr),
							BlockNumber: base.Blknum(blockNum),
						}
					}
				}
			}
		}
	} else {
		// Tokens format: holder-token-blocknum
		if len(parts) >= 3 {
			holderAddr := parts[0]
			tokenAddr := parts[1]

			// Check if holder address matches any pattern (exact match)
			for _, pattern := range addrPatterns {
				cleanPattern := strings.TrimPrefix(strings.ToLower(pattern), "0x")
				if holderAddr == cleanPattern {
					if blockNum, err := strconv.ParseUint(parts[2], 10, 64); err == nil {
						return &types.Token{
							Holder:      base.HexToAddress("0x" + holderAddr),
							Address:     base.HexToAddress("0x" + tokenAddr),
							BlockNumber: base.Blknum(blockNum),
						}
					}
				}
			}
		}
	}

	return nil
}

// LocationsFromTokens is kept for backward compatibility but now uses pattern-based approach
func LocationsFromTokens(conn *rpc.Connection, addresses []string, ids []identifiers.Identifier, includeApprovals bool) ([]cache.Locator, error) {
	// For decache purposes, we ignore block IDs and scan for all matching addresses
	scanMode := "tokens-only"
	if includeApprovals {
		scanMode = "both"
	}
	return LocationsFromTokensByPattern(conn, addresses, scanMode)
}
