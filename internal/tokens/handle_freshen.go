// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package tokensPkg

import "github.com/TrueBlocks/trueblocks-chifra/v6/pkg/monitor"

func (opts *TokensOptions) FreshenMonitorsForExport(monitorArray *[]monitor.Monitor) (bool, error) {
	var updater = monitor.NewUpdater(opts.Globals.Chain, opts.Globals.TestMode, opts.Globals.Decache /* skipFreshen */, opts.Addrs)
	return updater.FreshenMonitors(monitorArray)
}
