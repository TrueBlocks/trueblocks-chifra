package rpc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/decode"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/rpc/query"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/walk"
)

// erc721SupportsInterfaceData is the data needed to call the ERC-721 supportsInterface function
// 0x01ffc9a7: supportsInterface -- eips.ethereum.org/EIPS/eip-165
// 0x80ac58cd: ERC-721 interface ID -- eips.ethereum.org/EIPS/eip-721
const erc721SupportsInterfaceData = "0x01ffc9a780ac58cd00000000000000000000000000000000000000000000000000000000"

type tokenStateSelector = string

// TODO: If we used encoding we could use the function signature instead of the selector.

const tokenStateTotalSupply tokenStateSelector = "0x18160ddd"
const tokenStateDecimals tokenStateSelector = "0x313ce567"
const tokenStateSymbol tokenStateSelector = "0x95d89b41"
const tokenStateName tokenStateSelector = "0x06fdde03"
const tokenStateBalanceOf tokenStateSelector = "0x70a08231"

// GetTokenState returns token state for given block. `hexBlockNo` can be "latest" or "" for the latest
// block or decimal number or hex number with 0x prefix. (search: FromRpc)
func (conn *Connection) GetTokenState(tokenAddress base.Address, hexBlockNo string) (token *types.Token, err error) {
	if hexBlockNo != "" && hexBlockNo != "latest" && !strings.HasPrefix(hexBlockNo, "0x") {
		hexBlockNo = fmt.Sprintf("0x%x", base.MustParseUint64(hexBlockNo))
	}
	payloads := []query.BatchPayload{
		{
			Key: "name",
			Payload: &query.Payload{
				Method: "eth_call",
				Params: query.Params{
					map[string]any{
						"to":   tokenAddress,
						"data": tokenStateName,
					},
					hexBlockNo,
				},
			},
		},
		{
			Key: "symbol",
			Payload: &query.Payload{
				Method: "eth_call",
				Params: query.Params{
					map[string]any{
						"to":   tokenAddress,
						"data": tokenStateSymbol,
					},
					hexBlockNo,
				},
			},
		},
		{
			Key: "decimals",
			Payload: &query.Payload{
				Method: "eth_call",
				Params: query.Params{
					map[string]any{
						"to":   tokenAddress,
						"data": tokenStateDecimals,
					},
					hexBlockNo,
				},
			},
		},
		{
			Key: "totalSupply",
			Payload: &query.Payload{
				Method: "eth_call",
				Params: query.Params{
					map[string]any{
						"to":   tokenAddress,
						"data": tokenStateTotalSupply,
					},
					hexBlockNo,
				},
			},
		},
		// Supports interface: ERC 721
		{
			Key: "erc721",
			Payload: &query.Payload{
				Method: "eth_call",
				Params: query.Params{
					map[string]any{
						"to":   tokenAddress,
						"data": erc721SupportsInterfaceData,
					},
					hexBlockNo,
				},
			},
		},
	}

	results, err := query.QueryBatch[string](conn.Chain, payloads)
	if err != nil {
		return
	}

	name, _ := decode.ArticulateStringOrBytes(*results["name"])
	symbol, _ := decode.ArticulateStringOrBytes(*results["symbol"])

	var decimals uint64 = 0
	strDecimals := *results["decimals"]
	parsedDecimals, parseErr := strconv.ParseUint(strDecimals, 0, 8)
	if parseErr == nil {
		decimals = uint64(parsedDecimals)
	}

	totalSupply := base.HexToWei(*results["totalSupply"])

	// TODO: Maybe reconcsider this
	// TODO: According to ERC-20, name, symbol and decimals are optional, but such a token
	// TODO: would be of no use to us
	if name == "" && symbol == "" && decimals == 0 {
		return nil, errors.New(tokenAddress.Hex() + " address is not token")
	}

	tokenType := types.TokenErc20
	erc721, erc721Err := decode.ArticulateBool(*results["erc721"])
	if erc721Err == nil && erc721 {
		tokenType = types.TokenErc721
	}

	token = &types.Token{
		Address:     tokenAddress,
		Decimals:    decimals,
		Name:        name,
		Symbol:      symbol,
		TotalSupply: *totalSupply,
		TokenType:   tokenType,
	}

	return
}

// GetBalanceAtToken returns token balance for given block.
func (conn *Connection) GetBalanceAtToken(token, holder base.Address, bn base.Blknum) (*base.Wei, error) {
	if token.Equal(base.FAKE_ETH_ADDRESS) {
		return conn.getBalanceAt(holder, bn)
	}

	// Try reading from disk cache first
	tokenBalance := &types.Token{
		Holder:      holder,
		Address:     token,
		BlockNumber: bn,
	}
	if err := conn.ReadFromCache(tokenBalance); err == nil {
		return &tokenBalance.Balance, nil
	}

	hexBlockNo := fmt.Sprintf("0x%x", bn)
	if hexBlockNo != "" && hexBlockNo != "latest" && !strings.HasPrefix(hexBlockNo, "0x") {
		hexBlockNo = fmt.Sprintf("0x%x", base.MustParseUint64(hexBlockNo))
	}

	payloads := []query.BatchPayload{{
		Key: "balance",
		Payload: &query.Payload{
			Method: "eth_call",
			Params: query.Params{
				map[string]any{
					"to":   token.Hex(),
					"data": tokenStateBalanceOf + holder.Pad32(),
				},
				hexBlockNo,
			},
		},
	}}

	output, err := query.QueryBatch[string](conn.Chain, payloads)
	if err != nil {
		return nil, err
	}

	var balance *base.Wei
	if output["balance"] == nil {
		balance = base.NewWei(0)
	} else {
		balance = base.HexToWei(*output["balance"])
	}

	// Write to disk cache
	tokenBalance.Balance = *balance
	tokenBalance.Timestamp = conn.GetBlockTimestamp(bn)
	if err := conn.WriteToCache(tokenBalance, walk.Cache_Tokens, tokenBalance.Timestamp); err != nil {
		logger.Warn("Failed to write token balance to cache:", err)
	}

	return balance, nil
}

// GetAllowanceAtToken returns token allowance for given block.
func (conn *Connection) GetAllowanceAtToken(token, owner, spender base.Address, bn base.Blknum) (*base.Wei, error) {
	if token.IsZero() {
		return base.NewWei(0), fmt.Errorf("token address cannot be zero")
	}
	if owner.IsZero() {
		return base.NewWei(0), fmt.Errorf("owner address cannot be zero")
	}
	if spender.IsZero() {
		return base.NewWei(0), fmt.Errorf("spender address cannot be zero")
	}

	approval := &types.Approval{
		Owner:       owner,
		Token:       token,
		Spender:     spender,
		BlockNumber: bn,
	}
	if err := conn.ReadFromCache(approval); err == nil {
		return &approval.Allowance, nil
	}

	funcSig := "0xdd62ed3e"
	data := funcSig + owner.Pad32() + spender.Pad32()

	var blockParam string
	if bn == 0 {
		blockParam = "latest"
	} else {
		blockParam = fmt.Sprintf("0x%x", bn)
	}

	payloads := []query.BatchPayload{{
		Key: "allowance",
		Payload: &query.Payload{
			Method: "eth_call",
			Params: query.Params{
				map[string]any{
					"to":   token.Hex(),
					"data": data,
				},
				blockParam,
			},
		},
	}}

	output, err := query.QueryBatch[string](conn.Chain, payloads)
	if err != nil {
		return base.NewWei(0), fmt.Errorf("failed to query token allowance: %w", err)
	}

	var allowance *base.Wei
	if output["allowance"] == nil {
		allowance = base.NewWei(0)
	} else {
		allowance = base.HexToWei(*output["allowance"])
		if allowance == nil {
			return base.NewWei(0), fmt.Errorf("failed to parse allowance response")
		}
	}

	// Write to disk cache
	approval.Allowance = *allowance
	approval.Timestamp = conn.GetBlockTimestamp(bn)
	if err := conn.WriteToCache(approval, walk.Cache_Approvals, approval.Timestamp); err != nil {
		logger.Warn("Failed to write approval to cache:", err)
	}

	return allowance, nil
}
