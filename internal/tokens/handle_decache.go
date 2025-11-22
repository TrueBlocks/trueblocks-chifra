// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package tokensPkg

import (
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/decache"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/output"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/validate"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/walk"
)

func (opts *TokensOptions) HandleDecache(rCtx *output.RenderCtx) error {
	// Basic validation - ensure we have addresses
	if len(opts.Addrs) == 0 {
		return validate.Usage("Please provide at least one address.")
	}

	fetchData := func(modelChan chan types.Modeler, errorChan chan error) {
		showProgress := opts.Globals.ShowProgress()

		if opts.Approvals {
			// If --approvals specified, scan and clear only approvals cache
			approvalsToRemove, err := decache.LocationsFromTokensByPattern(opts.Conn, opts.Addrs, "approvals-only")
			if err != nil {
				errorChan <- err
				return
			}

			if msg, err := decache.Decache(opts.Conn, approvalsToRemove, showProgress, walk.Cache_Approvals); err != nil {
				errorChan <- err
			} else {
				s := types.Message{
					Msg: msg,
				}
				modelChan <- &s
			}
		} else {
			// If no --approvals, scan and clear both tokens and approvals caches
			allItemsToRemove, err := decache.LocationsFromTokensByPattern(opts.Conn, opts.Addrs, "both")
			if err != nil {
				errorChan <- err
				return
			}

			tokenCacheTypes := []walk.CacheType{walk.Cache_Tokens, walk.Cache_Approvals}
			for _, cacheType := range tokenCacheTypes {
				if msg, err := decache.Decache(opts.Conn, allItemsToRemove, showProgress, cacheType); err != nil {
					errorChan <- err
				} else {
					s := types.Message{
						Msg: msg,
					}
					modelChan <- &s
				}
			}
		}
	}

	opts.Globals.NoHeader = true
	return output.StreamMany(rCtx, fetchData, opts.Globals.OutputOpts())
}
