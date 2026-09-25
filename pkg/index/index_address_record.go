package index

import (
	"bytes"
	"encoding/binary"
	"io"
	"sort"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
)

const (
	// AddrRecordWidth - size of Address Record
	AddrRecordWidth = 28
)

func (chunk *Index) searchForAddressRecord(address base.Address) (int, error) {
	var searchErr error
	compareFunc := func(pos int) bool {
		if searchErr != nil {
			return true
		}
		if pos == int(chunk.Header.AddressCount) {
			return true
		}

		readLocation := int64(HeaderWidth + pos*AddrRecordWidth)
		_, err := chunk.File.Seek(readLocation, io.SeekStart)
		if err != nil {
			searchErr = err
			return true
		}

		addressRec := types.AddrRecord{}
		if err = binary.Read(chunk.File, binary.LittleEndian, &addressRec); err != nil {
			searchErr = err
			return true
		}

		return bytes.Compare(addressRec.Address.Bytes(), address.Bytes()) >= 0
	}

	pos := sort.Search(int(chunk.Header.AddressCount), compareFunc)
	if searchErr != nil {
		return -1, searchErr
	}
	if pos >= int(chunk.Header.AddressCount) {
		return -1, nil
	}

	readLocation := int64(HeaderWidth + pos*AddrRecordWidth)
	if _, err := chunk.File.Seek(readLocation, io.SeekStart); err != nil {
		return -1, err
	}
	rec := types.AddrRecord{}
	if err := binary.Read(chunk.File, binary.LittleEndian, &rec); err != nil {
		return -1, err
	}

	if !bytes.Equal(rec.Address.Bytes(), address.Bytes()) {
		return -1, nil
	}

	return pos, nil
}
