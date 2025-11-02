// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package config

import "github.com/TrueBlocks/trueblocks-chifra/v6/pkg/configtypes"

func GetVersion() configtypes.VersionGroup {
	return GetRootConfig().Version
}
