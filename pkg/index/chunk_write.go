package index

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/colors"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/config"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/file"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/ranges"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
)

type writeReport struct {
	chain        string
	fileRange    ranges.FileRange
	nAddresses   int
	nAppearances int
}

func (c *writeReport) Report(isSnapped bool, fileSize int64) string {
	report := `%s wrote {%d} address and {%d} appearance records to {$INDEX/%s.bin}`
	if isSnapped {
		report += ` @(snapped to grid)}`
	}
	report += " (size: {%d} , span: {%d})"
	return colors.ColoredWith(fmt.Sprintf(report, c.chain, c.nAddresses, c.nAppearances, c.fileRange, fileSize, c.fileRange.Span()), colors.BrightBlue)
}

func (chunk *Chunk) Write(chain string, publisher base.Address, fileName string, addrAppearanceMap map[string][]types.AppRecord, nApps int) (*writeReport, error) {
	_ = publisher
	// We're going to build two tables. An addressTable and an appearanceTable. We do this as we spin
	// through the map

	// Create space for the two tables...
	addressTable := make([]types.AddrRecord, 0, len(addrAppearanceMap))
	appearanceTable := make([]types.AppRecord, 0, nApps)

	// We want to sort the items in the map by address (maps in GoLang are not sorted)
	sorted := []string{}
	for address := range addrAppearanceMap {
		sorted = append(sorted, address)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// We need somewhere to store our progress...
	offset := uint32(0)
	bl := Bloom{}

	// For each address in the sorted list...
	for _, addrStr := range sorted {
		// ...get its appearances and append them to the appearanceTable....
		apps := addrAppearanceMap[addrStr]
		appearanceTable = append(appearanceTable, apps...)

		// ...add the address to the bloom filter...
		address := base.HexToAddress(addrStr)
		bl.InsertAddress(address)

		// ...and append the record to the addressTable.
		addressTable = append(addressTable, types.AddrRecord{
			Address: address,
			Offset:  offset,
			Count:   uint32(len(apps)),
		})

		// Finally, note the next offset
		offset += uint32(len(apps))
	}

	// At this point, the two tables and the bloom filter are fully populated. We're ready to write to disc...

	// First, we backup the existing chunk if there is one...
	indexFn := ToIndexPath(fileName)
	tmpPath := filepath.Join(config.PathToCache(chain), "tmp")
	if backup, err := file.MakeBackup(tmpPath, indexFn); err == nil {
		defer func() {
			backup.Restore()
		}()

		header := indexHeader{
			Magic:           file.MagicNumber,
			Hash:            base.BytesToHash(config.HeaderHash(config.ExpectedVersion())),
			AddressCount:    uint32(len(addressTable)),
			AppearanceCount: uint32(len(appearanceTable)),
		}
		if err := writeFileAtomic(indexFn, func(w io.Writer) error {
			if err := binary.Write(w, binary.LittleEndian, header); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, addressTable); err != nil {
				return err
			}
			return binary.Write(w, binary.LittleEndian, appearanceTable)
		}); err != nil {
			return nil, err
		}

		// The bloom is replaced atomically, so on failure any previous bloom is intact and
		// the backup mechanism restores the matching index.
		if _, err = bl.writeBloom(ToBloomPath(indexFn)); err != nil {
			return nil, err
		}

		// We're successfully written the chunk, so we don't need this any more. If the pin
		// fails we don't want to have to re-do this chunk, so remove this here.
		backup.Clear()
		return &writeReport{
			chain:        chain,
			fileRange:    ranges.RangeFromFilename(indexFn),
			nAddresses:   len(addressTable),
			nAppearances: len(appearanceTable),
		}, nil
	} else {
		return nil, err
	}
}

// Tag updates the manifest version in the chunk's header
func (chunk *Chunk) Tag(tag, fileName string) (err error) {
	blVers, idxVers, err := versions(fileName)
	if err != nil {
		return err
	}
	if blVers == tag && idxVers == tag {
		return nil
	}

	bloomFn := ToBloomPath(fileName)
	indexFn := ToIndexPath(fileName)
	indexBackup := indexFn + ".backup"
	bloomBackup := bloomFn + ".backup"

	defer func() {
		// If the backup files still exist when the function ends, something went wrong, reset everything
		if file.FileExists(indexBackup) || file.FileExists(bloomBackup) {
			_, _ = file.Copy(bloomFn, bloomBackup)
			_, _ = file.Copy(indexFn, indexBackup)
			_ = os.Remove(bloomBackup)
			_ = os.Remove(indexBackup)
		}
	}()

	if _, err = file.Copy(indexBackup, indexFn); err != nil {
		return err
	} else if _, err = file.Copy(bloomBackup, bloomFn); err != nil {
		return err
	}

	if err = chunk.Bloom.updateTag(tag, bloomFn); err != nil {
		return err
	}

	if err = chunk.Index.updateTag(tag, indexFn); err != nil {
		return err
	}

	_ = os.Remove(indexBackup)
	_ = os.Remove(bloomBackup)

	return nil
}
