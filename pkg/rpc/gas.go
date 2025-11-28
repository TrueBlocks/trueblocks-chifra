package rpc

import (
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/rpc/query"
)

// EstimateGasAndPrice performs both eth_estimateGas and eth_gasPrice in a single batch RPC call.
// Returns estimated gas units, current gas price (in wei per gas), and any error.
func EstimateGasAndPrice(chain string, from, to base.Address, data string, value *base.Wei) (estimatedGas, gasPrice base.Gas, err error) {
	valueHex := "0x0"
	if value != nil && !value.IsZero() {
		valueHex = "0x" + value.Text(16)
	}

	// Prepare batch payload for both RPC calls
	payloads := []query.BatchPayload{
		{
			Key: "estimateGas",
			Payload: &query.Payload{
				Method: "eth_estimateGas",
				Params: query.Params{
					map[string]string{
						"from":  from.Hex(),
						"to":    to.Hex(),
						"data":  data,
						"value": valueHex,
					},
				},
			},
		},
		{
			Key: "gasPrice",
			Payload: &query.Payload{
				Method: "eth_gasPrice",
				Params: query.Params{},
			},
		},
	}

	results, err := query.QueryBatch[string](chain, payloads)
	if err != nil {
		return 0, 0, err
	}

	// Parse estimateGas result
	if gasEstStr, ok := results["estimateGas"]; ok && gasEstStr != nil {
		estimatedGas = base.MustParseValue(*gasEstStr)
	}

	// Parse gasPrice result
	if gasPriceStr, ok := results["gasPrice"]; ok && gasPriceStr != nil {
		gasPrice = base.MustParseValue(*gasPriceStr)
	}

	return estimatedGas, gasPrice, nil
}
