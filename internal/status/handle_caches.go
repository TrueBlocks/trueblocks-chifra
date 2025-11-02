package statusPkg

import (
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/output"
)

func (opts *StatusOptions) HandleCaches(rCtx *output.RenderCtx) error {
	return opts.HandleChainsAndCaches(rCtx, false)
}
