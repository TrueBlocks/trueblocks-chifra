package sdk

import "github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"

var sdkTestMode = false

func init() {
	sdkTestMode = base.IsTestMode()
}
