// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package validate

import (
	"fmt"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
)

// ValidateBlockRange checks a --first_block / --last_block pair. Both ends
// are inclusive, so first == last is a valid one-block range. A first block
// later than the last block is refused. A first block of NOPOSN (an
// unparsable value) is refused too, with the message it always had.
func ValidateBlockRange(first, last base.Blknum) error {
	if first == base.NOPOSN {
		msg := fmt.Sprintf("first_block (%d) must be strictly earlier than last_block (%d).", first, last)
		return Usage(msg)
	}
	if first > last {
		msg := fmt.Sprintf("first_block (%d) must not be later than last_block (%d).", first, last)
		return Usage(msg)
	}
	return nil
}
