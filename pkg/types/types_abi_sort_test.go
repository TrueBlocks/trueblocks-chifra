package types

import (
	"reflect"
	"sort"
	"testing"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
)

func TestSortByAddressAscending(t *testing.T) {
	abis := []Abi{
		{Address: base.HexToAddress("0x2"), FileSize: 2048, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-01", NFunctions: 10, NEvents: 5, Name: "AbiTwo", Path: "/path/two"},
		{Address: base.HexToAddress("0x3"), FileSize: 4096, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-03", NFunctions: 12, NEvents: 3, Name: "AbiThree", Path: "/path/three"},
		{Address: base.HexToAddress("0x1"), FileSize: 1024, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: false, LastModDate: "2024-08-02", NFunctions: 8, NEvents: 6, Name: "AbiOne", Path: "/path/one"},
	}

	expected := []Abi{
		{Address: base.HexToAddress("0x1"), FileSize: 1024, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: false, LastModDate: "2024-08-02", NFunctions: 8, NEvents: 6, Name: "AbiOne", Path: "/path/one"},
		{Address: base.HexToAddress("0x2"), FileSize: 2048, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-01", NFunctions: 10, NEvents: 5, Name: "AbiTwo", Path: "/path/two"},
		{Address: base.HexToAddress("0x3"), FileSize: 4096, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-03", NFunctions: 12, NEvents: 3, Name: "AbiThree", Path: "/path/three"},
	}

	sort.Slice(abis, AbiCmp(abis, AbiBy(AbiAddress, Ascending)))

	if !reflect.DeepEqual(abis, expected) {
		t.Errorf("got %v, want %v", abis, expected)
	}
}

func TestSortByFileSizeDescending(t *testing.T) {
	abis := []Abi{
		{Address: base.HexToAddress("0x2"), FileSize: 2048, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-01", NFunctions: 10, NEvents: 5, Name: "AbiTwo", Path: "/path/two"},
		{Address: base.HexToAddress("0x3"), FileSize: 4096, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-03", NFunctions: 12, NEvents: 3, Name: "AbiThree", Path: "/path/three"},
		{Address: base.HexToAddress("0x1"), FileSize: 1024, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: false, LastModDate: "2024-08-02", NFunctions: 8, NEvents: 6, Name: "AbiOne", Path: "/path/one"},
	}

	expected := []Abi{
		{Address: base.HexToAddress("0x3"), FileSize: 4096, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-03", NFunctions: 12, NEvents: 3, Name: "AbiThree", Path: "/path/three"},
		{Address: base.HexToAddress("0x2"), FileSize: 2048, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-01", NFunctions: 10, NEvents: 5, Name: "AbiTwo", Path: "/path/two"},
		{Address: base.HexToAddress("0x1"), FileSize: 1024, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: false, LastModDate: "2024-08-02", NFunctions: 8, NEvents: 6, Name: "AbiOne", Path: "/path/one"},
	}

	sort.Slice(abis, AbiCmp(abis, AbiBy(AbiFileSize, Descending)))

	if !reflect.DeepEqual(abis, expected) {
		t.Errorf("got %v, want %v", abis, expected)
	}
}

func TestSortByIsKnownDescendingThenByNameAscending(t *testing.T) {
	abis := []Abi{
		{Address: base.HexToAddress("0x2"), FileSize: 2048, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-01", NFunctions: 10, NEvents: 5, Name: "AbiTwo", Path: "/path/two"},
		{Address: base.HexToAddress("0x1"), FileSize: 1024, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: false, LastModDate: "2024-08-02", NFunctions: 8, NEvents: 6, Name: "AbiOne", Path: "/path/one"},
		{Address: base.HexToAddress("0x3"), FileSize: 4096, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-03", NFunctions: 12, NEvents: 3, Name: "AbiThree", Path: "/path/three"},
	}

	expected := []Abi{
		{Address: base.HexToAddress("0x3"), FileSize: 4096, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-03", NFunctions: 12, NEvents: 3, Name: "AbiThree", Path: "/path/three"},
		{Address: base.HexToAddress("0x2"), FileSize: 2048, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: true, LastModDate: "2024-08-01", NFunctions: 10, NEvents: 5, Name: "AbiTwo", Path: "/path/two"},
		{Address: base.HexToAddress("0x1"), FileSize: 1024, Functions: []Function{}, HasConstructor: false, HasFallback: false, IsEmpty: false, IsKnown: false, LastModDate: "2024-08-02", NFunctions: 8, NEvents: 6, Name: "AbiOne", Path: "/path/one"},
	}

	sort.Slice(abis, AbiCmp(abis, AbiBy(AbiIsKnown, Descending), AbiBy(AbiName, Ascending)))

	if !reflect.DeepEqual(abis, expected) {
		t.Errorf("got %v, want %v", abis, expected)
	}
}
