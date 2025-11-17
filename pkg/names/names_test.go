package names

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
)

var inputValid = `
tags	address	name	symbol	source	decimals	deleted	isCustom	isPrefund	isContract	isErc20	isErc721
55-Defi	0x000000000000000000000000000000000000dead	ENS: Burn Address		EtherScan.io		false	false	false	false	false	false
50-Tokens:ERC721	0x0000000000001b84b1cb32787b0d64758d019317	HomeWork 🏠🛠️	HWK	On chain	18	false	false	false	true	true	true
50-Tokens:ERC20	0x0000000000004946c0e9f43f4dee607b0ef1fa1c	Chi Gastoken by 1inch	CHI	On chain	18	false	false	false	true	true	false
30-Contracts	0x000000000000541e251335090ac5b47176af4f7e	dex.blue		EtherScan.io		false	false	false	true	false	false
`

// missing "address" field
var inputMissingFields = `
tags	name	symbol	source	decimals	deleted	isCustom	isPrefund	isContract	isErc20	isErc721
55-Defi	ENS: Burn Address		EtherScan.io		false	false	false	false	false	false
50-Tokens:ERC721	HomeWork 🏠🛠️	HWK	On chain	18	false	false	false	true	true	true
50-Tokens:ERC20	Chi Gastoken by 1inch	CHI	On chain	18	false	false	false	true	true	false
30-Contracts	dex.blue		EtherScan.io		false	false	false	true	false	false
`

func TestNewNameReader(t *testing.T) {
	r, err := NewNameReader(strings.NewReader(inputValid))
	if err != nil {
		t.Fatal(err)
	}

	expected := []types.Name{
		{
			Tags:    "55-Defi",
			Address: base.HexToAddress("0x000000000000000000000000000000000000dead"),
			Name:    "ENS: Burn Address",
			Source:  "EtherScan.io",
		},
		{
			Tags:       "50-Tokens:ERC721",
			Address:    base.HexToAddress("0x0000000000001b84b1cb32787b0d64758d019317"),
			Name:       "HomeWork 🏠🛠️",
			Symbol:     "HWK",
			Decimals:   18,
			Source:     "On chain",
			IsContract: true,
			IsErc20:    true,
			IsErc721:   true,
		},
		{
			Tags:       "50-Tokens:ERC20",
			Address:    base.HexToAddress("0x0000000000004946c0e9f43f4dee607b0ef1fa1c"),
			Name:       "Chi Gastoken by 1inch",
			Symbol:     "CHI",
			Decimals:   18,
			Source:     "On chain",
			IsContract: true,
			IsErc20:    true,
			IsErc721:   false,
		},
		{
			Tags:       "30-Contracts",
			Address:    base.HexToAddress("0x000000000000541e251335090ac5b47176af4f7e"),
			Name:       "dex.blue",
			Source:     "EtherScan.io",
			IsContract: true,
		},
	}
	result := make([]types.Name, 0, len(expected))

	for {
		read, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, read)
	}

	if len(result) != len(expected) {
		t.Fatalf("length mismatch: want %d, got %d", len(result), len(expected))
	}

	for index, resultItem := range result {
		if !reflect.DeepEqual(resultItem, expected[index]) {
			t.Fatalf("wrong value: %+v", resultItem)
		}
	}

	// Test invalid input (missing required fields)
	r, err = NewNameReader(strings.NewReader(inputMissingFields))
	if err == nil {
		t.Fatal("error expected")
	}
}

func TestGetNamesForAddresses(t *testing.T) {
	// Test with empty slice - this should return all names (current behavior)
	result, err := GetNamesForAddresses("mainnet", []string{})
	if err != nil {
		t.Fatal(err)
	}
	// Empty input returns all names, so we just verify we got some results
	if len(result) == 0 {
		t.Fatal("expected some results when no addresses specified")
	}

	// Test with default parts (should use types.All when no parts specified)
	testAddresses := []string{
		"0x000000000000000000000000000000000000dead",
		"0x0000000000000000000000000000000000000001",
		"0x1234567890123456789012345678901234567890", // non-existent address
	}

	result, err = GetNamesForAddresses("mainnet", testAddresses) // uses types.All by default
	if err != nil {
		t.Fatal(err)
	}

	// The result should only contain addresses that have names
	// We can't guarantee specific names will be present without loading the actual database
	// but we can verify the function works and returns a valid map
	for addr, name := range result {
		if name.Address != addr {
			t.Fatalf("address mismatch: map key %s != name.Address %s", addr.Hex(), name.Address.Hex())
		}
	}
}
