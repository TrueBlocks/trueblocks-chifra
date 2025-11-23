// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package exportPkg

import (
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/monitor"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/output"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/rpc"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/topics"
)

func (opts *ExportOptions) HandleApprovals(rCtx *output.RenderCtx, monitorArray []monitor.Monitor) error {
	logFilter := rpc.NewLogFilter(opts.Emitter, []string{topics.ApprovalTopic.Hex()})
	return opts.handleShowCore(rCtx, monitorArray, logFilter)
}
