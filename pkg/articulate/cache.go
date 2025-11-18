package articulate

import (
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/abi"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/rpc"
)

type AbiCache struct {
	Conn           *rpc.Connection
	Chain          string
	AbiMap         abi.SelectorSyncMap
	loadedMap      abi.AddressSyncMap
	skipMap        abi.AddressSyncMap
	suppressErrors bool
}

func NewAbiCache(conn *rpc.Connection, suppressErrors, loadKnown bool) *AbiCache {
	ret := &AbiCache{
		Conn:           conn,
		Chain:          conn.Chain,
		AbiMap:         abi.SelectorSyncMap{},
		loadedMap:      abi.AddressSyncMap{},
		skipMap:        abi.AddressSyncMap{},
		suppressErrors: suppressErrors,
	}

	if loadKnown {
		if err := ret.AbiMap.LoadKnownAbis(conn.Chain); err != nil {
			// report error, but continue processing
			if !suppressErrors {
				logger.Error("error preloading known abis", "error", err)
			}
		}
	}

	return ret
}

func (abiCache *AbiCache) returnError(err error) error {
	if abiCache.suppressErrors {
		return nil
	}
	return err
}
