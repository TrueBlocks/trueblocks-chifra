package parser

import (
	"fmt"
	"reflect"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
)

// ContractArgument represents input to the smart contract method call, e.g.
// `true` in `setSomething(true)`
type ContractArgument struct {
	Tokens  []lexer.Token // the token that was parsed for this argument
	String  *ArgString    `parser:"@String"`             // the value if it's a string
	Number  *ArgNumber    `parser:"| @Decimal"`          // the value if it's a number
	Boolean *ArgBool      `parser:"| @('true'|'false')"` // the value if it's a boolean
	Hex     *ArgHex       `parser:"| @Hex"`              // the value if it's a hex string
	EnsAddr *ArgEnsDomain `parser:"| @EnsDomain"`        // the value if it's an ENS domain
}

// Interface returns the value as interface{} (any)
func (a *ContractArgument) Interface() any {
	if a.EnsAddr != nil {
		return *a.EnsAddr
	}

	if a.String != nil {
		value := *a.String
		return string(value)
	}

	if a.Number != nil {
		return a.Number.Interface()
	}

	if a.Boolean != nil {
		return *a.Boolean
	}

	if a.Hex != nil {
		if a.Hex.Address != nil {
			return *a.Hex.Address
		}
		return *a.Hex.String
	}

	return nil
}

func (a *ContractArgument) AbiType(abiType *abi.Type) (any, error) {
	// Type checking with nice error messages before conversion
	if abiType.T == abi.AddressTy && (a.Hex == nil || a.Hex.Address == nil) && a.EnsAddr == nil {
		return nil, wrongTypeError(abiType.String(), a.Tokens[0], a.Interface())
	}
	if abiType.T == abi.UintTy && a.Number == nil {
		return nil, wrongTypeError(abiType.String(), a.Tokens[0], a.Interface())
	}
	if abiType.T == abi.IntTy && a.Number == nil {
		return nil, wrongTypeError(abiType.String(), a.Tokens[0], a.Interface())
	}
	if abiType.T == abi.StringTy && a.String == nil {
		return nil, wrongTypeError(abiType.String(), a.Tokens[0], a.Interface())
	}
	if abiType.T == abi.BoolTy && a.Boolean == nil {
		return nil, wrongTypeError(abiType.String(), a.Tokens[0], a.Interface())
	}
	if abiType.T == abi.FixedBytesTy && a.Hex == nil {
		return nil, wrongTypeError("hash", a.Tokens[0], a.Interface())
	}

	// Delegate to Parameter.AbiType for all conversions
	param := types.Parameter{
		ParameterType: abiType.String(),
		Value:         fmt.Sprintf("%v", a.Interface()),
	}

	result, err := param.AbiType(abiType)
	if err != nil {
		// Wrap error with token context for better error messages
		if len(a.Tokens) > 0 {
			return nil, wrongTypeError(abiType.String(), a.Tokens[0], a.Interface())
		}
		return nil, err
	}

	return result, nil
}

// wrongTypeError returns user-friendly errors
func wrongTypeError(expectedType string, token lexer.Token, value any) error {
	t := reflect.TypeOf(value)
	if t == nil {
		return fmt.Errorf("expected %s, but got nil \"%s\"", expectedType, token)
	}
	typeName := t.String()
	kind := t.Kind()
	// kinds between this range are all (u)int, called "integer" in Solidity
	if kind > 1 && kind < 12 {
		typeName = "integer"
	}
	return fmt.Errorf("expected %s, but got %s \"%s\"", expectedType, typeName, token)
}
