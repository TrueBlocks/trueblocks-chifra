package index

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/config"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/file"
)

// writeBloom writes a single Bloom filter to file. We do not make a backup copy of the file
// because the caller is responsible for that. This is because the caller may be writing the
// entire chunk (both Bloom and Index) and we want either both to succeed or both to fail.
func (bl *Bloom) writeBloom(fileName string) ( /* changed */ bool, error) {
	bl.Header.Magic = file.SmallMagicNumber
	bl.Header.Hash = base.BytesToHash(config.HeaderHash(config.ExpectedVersion()))
	err := writeFileAtomic(fileName, func(w io.Writer) error {
		if err := binary.Write(w, binary.LittleEndian, bl.Header); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, bl.Count); err != nil {
			return err
		}
		for _, bb := range bl.Blooms {
			if err := binary.Write(w, binary.LittleEndian, bb.NInserted); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, bb.Bytes); err != nil {
				return err
			}
		}
		return nil
	})
	return err == nil, err
}

// updateTag writes a the header back to the bloom file
func (bl *Bloom) updateTag(tag, fileName string) error {
	var err error
	if bl.File, err = os.OpenFile(fileName, os.O_RDWR, 0644); err != nil {
		return err

	} else {
		defer func() {
			_ = bl.File.Sync()
			_ = bl.File.Close()
			bl.File = nil
		}()

		bl.Header.Magic = file.SmallMagicNumber
		bl.Header.Hash = base.BytesToHash(config.HeaderHash(tag))

		_, _ = bl.File.Seek(0, io.SeekStart) // already true, but can't hurt
		return binary.Write(bl.File, binary.LittleEndian, bl.Header)
	}
}
