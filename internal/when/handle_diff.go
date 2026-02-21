// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package whenPkg

import (
	"errors"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/identifiers"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/output"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/tslib"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/ethereum/go-ethereum"
)

func (opts *WhenOptions) HandleDiff(rCtx *output.RenderCtx) error {
	chain := opts.Globals.Chain

	fetchData := func(modelChan chan types.Modeler, errorChan chan error) {
		for _, br := range opts.BlockIds {
			if br.ModifierType != identifiers.Period {
				errorChan <- errors.New("the --diff option requires a period modifier (e.g., :weekly, :monthly)")
				continue
			}

			blockNums, err := br.ResolveBlocks(chain)
			if err != nil {
				errorChan <- err
				if errors.Is(err, ethereum.NotFound) {
					continue
				}
				rCtx.Cancel()
				return
			}

			if len(blockNums) < 2 {
				errorChan <- errors.New("the --diff option requires a block range that resolves to at least two blocks")
				continue
			}

			showProgress := opts.Globals.ShowProgress()
			bar := logger.NewBar(logger.BarOptions{
				Enabled: showProgress,
				Total:   int64(len(blockNums) - 1),
			})

			for i := 0; i < len(blockNums)-1; i++ {
				bn := blockNums[i]
				nextBn := blockNums[i+1]
				nBlocks := base.Blknum(nextBn - bn)

				block, err := opts.Conn.GetBlockHeaderByNumber(bn)
				if err != nil {
					errorChan <- err
					if errors.Is(err, ethereum.NotFound) {
						continue
					}
					rCtx.Cancel()
					return
				}

				nb, _ := tslib.FromBnToNamedBlock(chain, block.BlockNumber)
				if nb == nil {
					nb = &types.NamedBlock{
						BlockNumber: block.BlockNumber,
						Timestamp:   block.Timestamp,
					}
				}
				nb.NBlocks = nBlocks

				modelChan <- nb
				bar.Tick()
			}
			bar.Finish(true)
		}
	}

	return output.StreamMany(rCtx, fetchData, opts.Globals.OutputOpts())
}
