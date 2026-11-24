// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package validate

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
)

func TestValidateBlockRange(t *testing.T) {
	cases := []struct {
		name        string
		first, last base.Blknum
		ok          bool
	}{
		{"one-block range is legal", 28131489, 28131489, true},
		{"ordinary range", 28131486, 28131492, true},
		{"open range", 0, base.NOPOSN, true},
		{"first after last is refused", 28131490, 28131489, false},
		{"unparsable first block is refused", base.NOPOSN, base.NOPOSN, false},
	}
	for _, c := range cases {
		err := ValidateBlockRange(c.first, c.last)
		if c.ok && err != nil {
			t.Errorf("%s: want nil, got %v", c.name, err)
		}
		if !c.ok && err == nil {
			t.Errorf("%s: want an error, got nil", c.name)
		}
	}
}
