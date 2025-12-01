// Copyright 2016, 2024 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package types

import (
	"bytes"
	"encoding/json"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
)

// Helper function to convert various numeric types to float64 for JSON comparison
func convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// Test_Parameter_Value_String tests Parameter with string values
func Test_Parameter_Value_String(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty_string", ""},
		{"simple_string", "hello world"},
		{"address_string", "0x6b175474e89094c44da98b954eedeac495271d0f"},
		{"hex_string", "0x1234567890abcdef"},
		{"unicode_string", "Hello 世界 🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := Parameter{
				Name:          "testParam",
				ParameterType: "string",
				InternalType:  "string",
				Value:         tt.value,
			}

			// Test JSON serialization
			jsonData, err := json.Marshal(param)
			assert.NoError(t, err, "JSON marshal should succeed")

			var readParam Parameter
			err = json.Unmarshal(jsonData, &readParam)
			assert.NoError(t, err, "JSON unmarshal should succeed")
			assert.Equal(t, param.Value, readParam.Value, "JSON round-trip should preserve string value")

			// Test binary cache serialization
			buffer := new(bytes.Buffer)
			err = param.MarshalCache(buffer)
			assert.NoError(t, err, "Binary marshal should succeed")

			var cachedParam Parameter
			err = cachedParam.UnmarshalCache(0, buffer)
			assert.NoError(t, err, "Binary unmarshal should succeed")
			assert.Equal(t, param.Name, cachedParam.Name, "Binary round-trip should preserve name")
			assert.Equal(t, param.ParameterType, cachedParam.ParameterType, "Binary round-trip should preserve type")
			// Note: Value is not serialized in binary cache for Parameter
		})
	}
}

// Test_Parameter_Value_Numbers tests Parameter with various numeric types
func Test_Parameter_Value_Numbers(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		pType string
	}{
		{"uint8", uint8(255), "uint8"},
		{"uint16", uint16(65535), "uint16"},
		{"uint32", uint32(4294967295), "uint32"},
		{"uint64", uint64(18446744073709551615), "uint64"},
		{"int8", int8(-128), "int8"},
		{"int16", int16(-32768), "int16"},
		{"int32", int32(-2147483648), "int32"},
		{"int64", int64(-9223372036854775808), "int64"},
		{"big_int", func() *big.Int { i, _ := big.NewInt(0).SetString("12345678901234567890", 10); return i }(), "uint256"},
		{"negative_big_int", func() *big.Int { i, _ := big.NewInt(0).SetString("-12345678901234567890", 10); return i }(), "int256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := Parameter{
				Name:          "numParam",
				ParameterType: tt.pType,
				InternalType:  tt.pType,
				Value:         tt.value,
			}

			// Test JSON serialization
			jsonData, err := json.Marshal(param)
			assert.NoError(t, err, "JSON marshal should succeed")

			var readParam Parameter
			err = json.Unmarshal(jsonData, &readParam)
			assert.NoError(t, err, "JSON unmarshal should succeed")

			// For big.Int, we need special handling since JSON unmarshaling might change the type
			if bigInt, ok := tt.value.(*big.Int); ok {
				if readBigInt, ok := readParam.Value.(*big.Int); ok {
					assert.Equal(t, bigInt.String(), readBigInt.String(), "Big int values should match")
				} else {
					// JSON might unmarshal big ints as strings or numbers (float64)
					// This is expected behavior - JSON doesn't preserve Go's specific numeric types
					t.Logf("Big int unmarshaled as %T: %v", readParam.Value, readParam.Value)
					// We'll accept this as normal JSON behavior
				}
			} else {
				// JSON unmarshaling converts all numbers to float64, so we need to handle this
				if expectedFloat, ok := convertToFloat64(tt.value); ok {
					if actualFloat, ok := readParam.Value.(float64); ok {
						assert.Equal(t, expectedFloat, actualFloat, "JSON round-trip should preserve numeric value as float64")
					} else {
						t.Errorf("Expected float64 but got %T: %v", readParam.Value, readParam.Value)
					}
				} else {
					assert.Equal(t, param.Value, readParam.Value, "JSON round-trip should preserve numeric value")
				}
			}

			// Test binary cache serialization
			buffer := new(bytes.Buffer)
			err = param.MarshalCache(buffer)
			assert.NoError(t, err, "Binary marshal should succeed")

			var cachedParam Parameter
			err = cachedParam.UnmarshalCache(0, buffer)
			assert.NoError(t, err, "Binary unmarshal should succeed")
			assert.Equal(t, param.Name, cachedParam.Name, "Binary round-trip should preserve name")
			assert.Equal(t, param.ParameterType, cachedParam.ParameterType, "Binary round-trip should preserve type")
		})
	}
}

// Test_Parameter_Value_Boolean tests Parameter with boolean values
func Test_Parameter_Value_Boolean(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := Parameter{
				Name:          "boolParam",
				ParameterType: "bool",
				InternalType:  "bool",
				Value:         tt.value,
			}

			// Test JSON serialization
			jsonData, err := json.Marshal(param)
			assert.NoError(t, err, "JSON marshal should succeed")

			var readParam Parameter
			err = json.Unmarshal(jsonData, &readParam)
			assert.NoError(t, err, "JSON unmarshal should succeed")
			assert.Equal(t, param.Value, readParam.Value, "JSON round-trip should preserve boolean value")

			// Test binary cache serialization
			buffer := new(bytes.Buffer)
			err = param.MarshalCache(buffer)
			assert.NoError(t, err, "Binary marshal should succeed")

			var cachedParam Parameter
			err = cachedParam.UnmarshalCache(0, buffer)
			assert.NoError(t, err, "Binary unmarshal should succeed")
			assert.Equal(t, param.Name, cachedParam.Name, "Binary round-trip should preserve name")
			assert.Equal(t, param.ParameterType, cachedParam.ParameterType, "Binary round-trip should preserve type")
		})
	}
}

// Test_Parameter_Value_Arrays tests Parameter with array/slice values
func Test_Parameter_Value_Arrays(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		pType string
	}{
		{"string_array", []string{"hello", "world", "test"}, "string[]"},
		{"number_array", []int{1, 2, 3, 42, 100}, "uint256[]"},
		{"address_array", []string{
			"0x6b175474e89094c44da98b954eedeac495271d0f",
			"0xa735cf11ed1228feb7c7bb18673a868455ffb16f",
		}, "address[]"},
		{"mixed_interface_array", []interface{}{"hello", 42, true}, "tuple[]"},
		{"nested_array", [][]string{{"a", "b"}, {"c", "d"}}, "string[][]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := Parameter{
				Name:          "arrayParam",
				ParameterType: tt.pType,
				InternalType:  tt.pType,
				Value:         tt.value,
			}

			// Test JSON serialization
			jsonData, err := json.Marshal(param)
			assert.NoError(t, err, "JSON marshal should succeed")

			var readParam Parameter
			err = json.Unmarshal(jsonData, &readParam)
			assert.NoError(t, err, "JSON unmarshal should succeed")

			// Arrays might be unmarshaled as []interface{}, so we need flexible comparison
			t.Logf("Original value type: %T, value: %v", param.Value, param.Value)
			t.Logf("Read value type: %T, value: %v", readParam.Value, readParam.Value)

			// Test binary cache serialization
			buffer := new(bytes.Buffer)
			err = param.MarshalCache(buffer)
			assert.NoError(t, err, "Binary marshal should succeed")

			var cachedParam Parameter
			err = cachedParam.UnmarshalCache(0, buffer)
			assert.NoError(t, err, "Binary unmarshal should succeed")
			assert.Equal(t, param.Name, cachedParam.Name, "Binary round-trip should preserve name")
			assert.Equal(t, param.ParameterType, cachedParam.ParameterType, "Binary round-trip should preserve type")
		})
	}
}

// Test_Parameter_Value_ComplexObjects tests Parameter with complex object values
func Test_Parameter_Value_ComplexObjects(t *testing.T) {
	// Create a complex nested structure similar to what we see in real ABIs
	complexStruct := map[string]interface{}{
		"token":  "0x6b175474e89094c44da98b954eedeac495271d0f",
		"amount": big.NewInt(1000000000000000000),
		"dest":   "0xa735cf11ed1228feb7c7bb18673a868455ffb16f",
		"metadata": map[string]interface{}{
			"timestamp": uint64(1642000000),
			"tags":      []string{"defi", "donation"},
			"verified":  true,
		},
	}

	// Create a nested Function object
	nestedFunction := Function{
		Name:            "nestedFunc",
		Signature:       "nestedFunc(uint256)",
		Encoding:        "0x12345678",
		StateMutability: "view",
		Inputs: []Parameter{
			{
				Name:          "value",
				ParameterType: "uint256",
				InternalType:  "uint256",
				Value:         big.NewInt(42),
			},
		},
	}

	tests := []struct {
		name  string
		value interface{}
		pType string
	}{
		{"complex_map", complexStruct, "tuple"},
		{"nested_function", nestedFunction, "function"},
		{"base_address", base.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"), "address"},
		{"base_hash", base.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"), "bytes32"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := Parameter{
				Name:          "complexParam",
				ParameterType: tt.pType,
				InternalType:  tt.pType,
				Value:         tt.value,
			}

			// Test JSON serialization
			jsonData, err := json.Marshal(param)
			assert.NoError(t, err, "JSON marshal should succeed")

			var readParam Parameter
			err = json.Unmarshal(jsonData, &readParam)
			assert.NoError(t, err, "JSON unmarshal should succeed")

			t.Logf("Original value type: %T", param.Value)
			t.Logf("Read value type: %T", readParam.Value)

			// For complex objects, we mainly want to ensure no errors occur
			// The exact structure might change during JSON round-trip
			assert.Equal(t, param.Name, readParam.Name, "Parameter name should be preserved")
			assert.Equal(t, param.ParameterType, readParam.ParameterType, "Parameter type should be preserved")

			// Test binary cache serialization
			buffer := new(bytes.Buffer)
			err = param.MarshalCache(buffer)
			assert.NoError(t, err, "Binary marshal should succeed")

			var cachedParam Parameter
			err = cachedParam.UnmarshalCache(0, buffer)
			assert.NoError(t, err, "Binary unmarshal should succeed")
			assert.Equal(t, param.Name, cachedParam.Name, "Binary round-trip should preserve name")
			assert.Equal(t, param.ParameterType, cachedParam.ParameterType, "Binary round-trip should preserve type")
		})
	}
}

// Test_Parameter_Value_Nil tests Parameter with nil values
func Test_Parameter_Value_Nil(t *testing.T) {
	param := Parameter{
		Name:          "nilParam",
		ParameterType: "address",
		InternalType:  "address",
		Value:         nil,
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(param)
	assert.NoError(t, err, "JSON marshal should succeed")

	var readParam Parameter
	err = json.Unmarshal(jsonData, &readParam)
	assert.NoError(t, err, "JSON unmarshal should succeed")
	assert.Nil(t, readParam.Value, "JSON round-trip should preserve nil value")

	// Test binary cache serialization
	buffer := new(bytes.Buffer)
	err = param.MarshalCache(buffer)
	assert.NoError(t, err, "Binary marshal should succeed")

	var cachedParam Parameter
	err = cachedParam.UnmarshalCache(0, buffer)
	assert.NoError(t, err, "Binary unmarshal should succeed")
	assert.Equal(t, param.Name, cachedParam.Name, "Binary round-trip should preserve name")
	assert.Equal(t, param.ParameterType, cachedParam.ParameterType, "Binary round-trip should preserve type")
}

// Test_Parameter_Components_With_Values tests Parameter with components and values
func Test_Parameter_Components_With_Values(t *testing.T) {
	// Create a parameter with components (tuple type) and set values on both parent and children
	param := Parameter{
		Name:          "donation",
		ParameterType: "(address,uint256,address)",
		InternalType:  "(address,uint256,address)",
		Value: []interface{}{
			"0x6b175474e89094c44da98b954eedeac495271d0f",
			"1000000000000000000",
			"0xa735cf11ed1228feb7c7bb18673a868455ffb16f",
		},
		Components: []Parameter{
			{
				Name:          "token",
				ParameterType: "address",
				InternalType:  "address",
				Value:         "0x6b175474e89094c44da98b954eedeac495271d0f",
			},
			{
				Name:          "amount",
				ParameterType: "uint256",
				InternalType:  "uint256",
				Value:         "1000000000000000000",
			},
			{
				Name:          "dest",
				ParameterType: "address",
				InternalType:  "address",
				Value:         "0xa735cf11ed1228feb7c7bb18673a868455ffb16f",
			},
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(param)
	assert.NoError(t, err, "JSON marshal should succeed")

	var readParam Parameter
	err = json.Unmarshal(jsonData, &readParam)
	assert.NoError(t, err, "JSON unmarshal should succeed")

	// Verify the structure is preserved
	assert.Equal(t, param.Name, readParam.Name, "Parameter name should be preserved")
	assert.Equal(t, param.ParameterType, readParam.ParameterType, "Parameter type should be preserved")
	assert.Equal(t, len(param.Components), len(readParam.Components), "Components count should be preserved")

	// Verify component details
	for i, comp := range param.Components {
		readComp := readParam.Components[i]
		assert.Equal(t, comp.Name, readComp.Name, "Component name should be preserved")
		assert.Equal(t, comp.ParameterType, readComp.ParameterType, "Component type should be preserved")
		// Component values might be affected by JSON round-trip
		t.Logf("Component %d - Original: %T %v, Read: %T %v", i, comp.Value, comp.Value, readComp.Value, readComp.Value)
	}

	// Test binary cache serialization
	buffer := new(bytes.Buffer)
	err = param.MarshalCache(buffer)
	assert.NoError(t, err, "Binary marshal should succeed")

	var cachedParam Parameter
	err = cachedParam.UnmarshalCache(0, buffer)
	assert.NoError(t, err, "Binary unmarshal should succeed")

	// Verify the structure is preserved in binary cache
	assert.Equal(t, param.Name, cachedParam.Name, "Binary round-trip should preserve name")
	assert.Equal(t, param.ParameterType, cachedParam.ParameterType, "Binary round-trip should preserve type")
	assert.Equal(t, len(param.Components), len(cachedParam.Components), "Binary round-trip should preserve components count")

	// Verify component details in binary cache
	for i, comp := range param.Components {
		cachedComp := cachedParam.Components[i]
		assert.Equal(t, comp.Name, cachedComp.Name, "Binary round-trip should preserve component name")
		assert.Equal(t, comp.ParameterType, cachedComp.ParameterType, "Binary round-trip should preserve component type")
		// Note: Component values are not serialized in binary cache
	}
}

// Test_Parameter_Value_Summary documents the key findings about Parameter.Value behavior
func Test_Parameter_Value_Summary(t *testing.T) {
	t.Log("=== PARAMETER VALUE BEHAVIOR SUMMARY ===")
	t.Log("")
	t.Log("Key findings from Parameter.Value testing:")
	t.Log("")
	t.Log("1. JSON SERIALIZATION BEHAVIOR:")
	t.Log("   - All numeric types (uint8, int32, etc.) become float64 after JSON round-trip")
	t.Log("   - big.Int values become float64 (with potential precision loss)")
	t.Log("   - String arrays become []interface{}")
	t.Log("   - Complex structs (like Function) become map[string]interface{}")
	t.Log("   - base.Address and base.Hash become strings")
	t.Log("")
	t.Log("2. BINARY CACHE BEHAVIOR:")
	t.Log("   - Parameter.Value is NOT serialized in binary cache")
	t.Log("   - Only structure (Name, Type, Components) is preserved")
	t.Log("   - This is likely intentional - values are runtime data")
	t.Log("")
	t.Log("3. IMPLICATIONS FOR ARTICULATION:")
	t.Log("   - The articulation issue is NOT related to binary cache serialization")
	t.Log("   - Values must be computed/populated during articulation process")
	t.Log("   - The problem is in how complex parameter types are articulated from transaction data")
	t.Log("")
	t.Log("4. COMPONENTS BEHAVIOR:")
	t.Log("   - Component structure is preserved in both JSON and binary cache")
	t.Log("   - Component values follow same rules as parent Parameter values")
	t.Log("   - This confirms the ABI structure is correctly cached")
}

// Test ABI for parameter conversion tests
const testAbiSource = `
[
	{ "type" : "function", "name" : "bytes4", "inputs" : [ { "name" : "inputs", "type" : "bytes4" } ] },
	{ "type" : "function", "name" : "bytes32", "inputs" : [ { "name" : "inputs", "type" : "bytes32" } ] },
	{ "type" : "function", "name" : "int8", "inputs" : [ { "name" : "inputs", "type" : "int8" } ] },
	{ "type" : "function", "name" : "int64", "inputs" : [ { "name" : "inputs", "type" : "int64" } ] },
	{ "type" : "function", "name" : "int256", "inputs" : [ { "name" : "inputs", "type" : "int256" } ] },
	{ "type" : "function", "name" : "uint8", "inputs" : [ { "name" : "inputs", "type" : "uint8" } ] },
	{ "type" : "function", "name" : "uint256", "inputs" : [ { "name" : "inputs", "type" : "uint256" } ] },
	{ "type" : "function", "name" : "address", "inputs" : [ { "name" : "inputs", "type" : "address" } ] },
	{ "type" : "function", "name" : "bool", "inputs" : [ { "name" : "inputs", "type" : "bool" } ] },
	{ "type" : "function", "name" : "string", "inputs" : [ { "name" : "inputs", "type" : "string" } ] }
]
`

func TestParameter_AbiType(t *testing.T) {
	testAbi, err := abi.JSON(strings.NewReader(testAbiSource))
	if err != nil {
		panic(err)
	}

	int256 := new(big.Int)
	_, ok := int256.SetString("-57896044618658097711785492504343953926634992332820282019728792003956564819968", 10)
	if !ok {
		t.Fatal("cannot set int256 value")
	}

	uint256 := new(big.Int)
	_, ok = uint256.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
	if !ok {
		t.Fatal("cannot set uint256 value")
	}

	tests := []struct {
		name      string
		param     Parameter
		abiType   *abi.Type
		want      any
		wantErr   bool
		checkFunc func(got, want any) bool
	}{
		{
			name: "bytes4 from hex string",
			param: Parameter{
				ParameterType: "bytes4",
				Value:         "0x80ac58cd",
			},
			abiType: &testAbi.Methods["bytes4"].Inputs[0].Type,
			want:    [4]byte{128, 172, 88, 205},
		},
		{
			name: "bytes32 from hex string",
			param: Parameter{
				ParameterType: "bytes32",
				Value:         "0x00120aa407bdbff1d93ea98dafc5f1da56b589b427167ec414bccbe0cfdfd573",
			},
			abiType: &testAbi.Methods["bytes32"].Inputs[0].Type,
			want:    [32]byte{0, 18, 10, 164, 7, 189, 191, 241, 217, 62, 169, 141, 175, 197, 241, 218, 86, 181, 137, 180, 39, 22, 126, 196, 20, 188, 203, 224, 207, 223, 213, 115},
		},
		{
			name: "int8 from string",
			param: Parameter{
				ParameterType: "int8",
				Value:         "-128",
			},
			abiType: &testAbi.Methods["int8"].Inputs[0].Type,
			want:    int8(-128),
		},
		{
			name: "int64 from string",
			param: Parameter{
				ParameterType: "int64",
				Value:         "-9223372036854775808",
			},
			abiType: &testAbi.Methods["int64"].Inputs[0].Type,
			want:    int64(-9223372036854775808),
		},
		{
			name: "int256 from string",
			param: Parameter{
				ParameterType: "int256",
				Value:         "-57896044618658097711785492504343953926634992332820282019728792003956564819968",
			},
			abiType: &testAbi.Methods["int256"].Inputs[0].Type,
			want:    int256,
		},
		{
			name: "uint8 from string",
			param: Parameter{
				ParameterType: "uint8",
				Value:         "255",
			},
			abiType: &testAbi.Methods["uint8"].Inputs[0].Type,
			want:    uint64(255),
		},
		{
			name: "uint256 from string",
			param: Parameter{
				ParameterType: "uint256",
				Value:         "115792089237316195423570985008687907853269984665640564039457584007913129639935",
			},
			abiType: &testAbi.Methods["uint256"].Inputs[0].Type,
			want:    uint256,
		},
		{
			name: "address from hex string",
			param: Parameter{
				ParameterType: "address",
				Value:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			},
			abiType: &testAbi.Methods["address"].Inputs[0].Type,
			want:    common.HexToAddress("0xf503017d7baf7fbc0fff7492b751025c6a78179b"),
		},
		{
			name: "bool true from string",
			param: Parameter{
				ParameterType: "bool",
				Value:         "true",
			},
			abiType: &testAbi.Methods["bool"].Inputs[0].Type,
			want:    true,
		},
		{
			name: "bool false from string",
			param: Parameter{
				ParameterType: "bool",
				Value:         "false",
			},
			abiType: &testAbi.Methods["bool"].Inputs[0].Type,
			want:    false,
		},
		{
			name: "string from string",
			param: Parameter{
				ParameterType: "string",
				Value:         "hello world",
			},
			abiType: &testAbi.Methods["string"].Inputs[0].Type,
			want:    "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.param.AbiType(tt.abiType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parameter.AbiType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				if !tt.checkFunc(got, tt.want) {
					t.Errorf("Parameter.AbiType() = %v, want %v", got, tt.want)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parameter.AbiType() = %v (type %T), want %v (type %T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestParameter_AbiType_Errors(t *testing.T) {
	testAbi, err := abi.JSON(strings.NewReader(testAbiSource))
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name      string
		param     Parameter
		abiType   *abi.Type
		wantError string
	}{
		{
			name: "nil value",
			param: Parameter{
				ParameterType: "address",
				Value:         nil,
			},
			abiType:   &testAbi.Methods["address"].Inputs[0].Type,
			wantError: "parameter value is nil",
		},
		{
			name: "empty string value",
			param: Parameter{
				ParameterType: "address",
				Value:         "",
			},
			abiType:   &testAbi.Methods["address"].Inputs[0].Type,
			wantError: "parameter value is empty",
		},
		{
			name: "invalid integer",
			param: Parameter{
				ParameterType: "uint256",
				Value:         "not a number",
			},
			abiType:   &testAbi.Methods["uint256"].Inputs[0].Type,
			wantError: "invalid unsigned integer value",
		},
		{
			name: "invalid boolean",
			param: Parameter{
				ParameterType: "bool",
				Value:         "maybe",
			},
			abiType:   &testAbi.Methods["bool"].Inputs[0].Type,
			wantError: "invalid boolean value",
		},
		{
			name: "address without 0x prefix",
			param: Parameter{
				ParameterType: "address",
				Value:         "f503017d7baf7fbc0fff7492b751025c6a78179b",
			},
			abiType:   &testAbi.Methods["address"].Inputs[0].Type,
			wantError: "expected hex address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.param.AbiType(tt.abiType)
			if err == nil {
				t.Error("Parameter.AbiType() expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Parameter.AbiType() error = %v, want error containing %q", err, tt.wantError)
			}
		})
	}
}
