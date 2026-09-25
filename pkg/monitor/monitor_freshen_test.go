package monitor

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/config"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/file"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/index"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/manifest"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/ranges"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
)

func TestPartitionFreshenResultsSkipsPastHole(t *testing.T) {
	w := logger.GetLoggerWriter()
	defer logger.SetLoggerWriter(w)
	logger.SetLoggerWriter(io.Discard)

	results := []index.AppearanceResult{
		{Range: ranges.FileRange{First: 200, Last: 299}},
		{Range: ranges.FileRange{First: 100, Last: 199}, Err: errors.New("download failed")},
		{Range: ranges.FileRange{First: 0, Last: 99}},
	}
	keep, err := partitionFreshenResults(results)
	if err == nil || err.Error() != "000000100-000000199: download failed" {
		t.Fatalf("err=%v", err)
	}
	if len(keep) != 1 || keep[0].Range.First != 0 || keep[0].Range.Last != 99 {
		t.Fatalf("keep=%v", keep)
	}
}

func TestPartitionFreshenResultsNoErrorKeepsAll(t *testing.T) {
	results := []index.AppearanceResult{
		{Range: ranges.FileRange{First: 200, Last: 299}},
		{Range: ranges.FileRange{First: 0, Last: 99}},
	}
	keep, err := partitionFreshenResults(results)
	if err != nil {
		t.Fatal(err)
	}
	if len(keep) != 2 {
		t.Fatalf("keep=%v", keep)
	}
	if keep[0].Range.First != 0 || keep[1].Range.First != 200 {
		t.Fatalf("order=%v", keep)
	}
}

func TestPartitionFreshenResultsDropsWholeFailedRange(t *testing.T) {
	w := logger.GetLoggerWriter()
	defer logger.SetLoggerWriter(w)
	logger.SetLoggerWriter(io.Discard)

	results := []index.AppearanceResult{
		{Range: ranges.FileRange{First: 100, Last: 199}},
		{Range: ranges.FileRange{First: 100, Last: 199}, Err: errors.New("eof")},
		{Range: ranges.FileRange{First: 0, Last: 99}},
	}
	keep, err := partitionFreshenResults(results)
	if err == nil || err.Error() != "000000100-000000199: eof" {
		t.Fatalf("err=%v", err)
	}
	if len(keep) != 1 || keep[0].Range.First != 0 {
		t.Fatalf("keep=%v", keep)
	}
}

type freshenHole int

const (
	corruptIndexInManifest freshenHole = iota // restorable, so quarantined
	corruptIndexNoManifest                    // not restorable, so left in place
	missingBloom                              // published chunk failed to download
	unpublishedGap                            // local scrape gap the manifest does not list
)

func TestPartitionFreshenResultsStopsAtCancellation(t *testing.T) {
	results := []index.AppearanceResult{
		{Range: ranges.FileRange{First: 0, Last: 99}},
		{Range: ranges.FileRange{First: 100, Last: 199}, Err: index.ErrUserHitControlC},
		{Range: ranges.FileRange{First: 200, Last: 299}},
	}
	keep, err := partitionFreshenResults(results)
	if !errors.Is(err, index.ErrUserHitControlC) {
		t.Fatalf("err=%v", err)
	}
	if len(keep) != 1 || keep[0].Range.First != 0 {
		t.Fatalf("keep=%v", keep)
	}
}

func TestFreshenMonitorsStopsAtCorruptChunkAndResumes(t *testing.T) {
	testFreshenStopsAtHole(t, corruptIndexInManifest)
}

func TestFreshenMonitorsKeepsCorruptChunkWithoutManifest(t *testing.T) {
	testFreshenStopsAtHole(t, corruptIndexNoManifest)
}

func TestFreshenMonitorsStopsAtMissingBloom(t *testing.T) {
	testFreshenStopsAtHole(t, missingBloom)
}

func TestFreshenMonitorsAllowsUnpublishedGap(t *testing.T) {
	testFreshenStopsAtHole(t, unpublishedGap)
}

func testFreshenStopsAtHole(t *testing.T, hole freshenHole) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	const chain = "testchain"
	address := base.HexToAddress("0x1234567890123456789012345678901234567890")
	root := config.PathToIndex(chain)
	hash := config.HeaderHash(config.ExpectedVersion())
	writeChunk := func(first uint32) string {
		t.Helper()
		rng := ranges.FileRange{First: base.Blknum(first), Last: base.Blknum(first + 99)}
		// A bloom with every bit set guarantees that the index is consulted.
		bloom := make([]byte, 42+index.BLOOM_WIDTH_IN_BYTES)
		binary.LittleEndian.PutUint16(bloom, file.SmallMagicNumber)
		copy(bloom[2:34], hash)
		binary.LittleEndian.PutUint32(bloom[34:], 1)
		for i := 42; i < len(bloom); i++ {
			bloom[i] = 0xff
		}
		path := filepath.Join(root, "finalized", rng.String()+".bin")
		if err := os.WriteFile(index.ToBloomPath(path), bloom, 0600); err != nil {
			t.Fatal(err)
		}
		data := make([]byte, index.HeaderWidth+index.AddrRecordWidth+index.AppRecordWidth)
		binary.LittleEndian.PutUint32(data, file.MagicNumber)
		copy(data[4:36], hash)
		binary.LittleEndian.PutUint32(data[36:], 1)
		binary.LittleEndian.PutUint32(data[40:], 1)
		copy(data[44:64], address.Bytes())
		binary.LittleEndian.PutUint32(data[68:], 1)
		binary.LittleEndian.PutUint32(data[72:], first+1)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	writeChunk(0)
	broken := writeChunk(100)
	writeChunk(200)
	if hole == corruptIndexInManifest || hole == missingBloom {
		man := manifest.Manifest{Chain: chain}
		for _, rng := range []string{"000000000-000000099", "000000100-000000199", "000000200-000000299"} {
			man.Chunks = append(man.Chunks, types.ChunkRecord{Range: rng, IndexHash: "fakehash"})
		}
		manBytes, err := json.Marshal(man)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(config.PathToManifestFile(chain), manBytes, 0600); err != nil {
			t.Fatal(err)
		}
	}
	switch hole {
	case corruptIndexInManifest, corruptIndexNoManifest:
		if err := os.Truncate(broken, index.HeaderWidth); err != nil {
			t.Fatal(err)
		}
	default:
		if err := os.Remove(index.ToBloomPath(broken)); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(broken); err != nil {
			t.Fatal(err)
		}
	}
	freshen := func() error {
		updater := NewUpdater(chain, true, false, []string{address.Hex()})
		updater.MaxTasks = 3
		monitors := make([]Monitor, 0, 1)
		_, err := updater.FreshenMonitors(&monitors)
		return err
	}
	checkMonitor := func(last uint32, count int64) {
		t.Helper()
		mon, err := NewMonitor(chain, address, false)
		if err != nil {
			t.Fatal(err)
		}
		defer mon.Close()
		if err := mon.ReadMonitorHeader(); err != nil {
			t.Fatal(err)
		}
		if mon.LastScanned != last || mon.Count() != count {
			t.Fatalf("lastScanned=%d count=%d, want %d %d", mon.LastScanned, mon.Count(), last, count)
		}
	}
	err := freshen()
	if hole == unpublishedGap {
		if err != nil {
			t.Fatal(err)
		}
		checkMonitor(300, 2)
		return
	}
	if hole == missingBloom && (err == nil || !strings.Contains(err.Error(), "missing bloom filter for published chunk 000000100-000000199")) {
		t.Fatalf("expected missing bloom error, got %v", err)
	}
	if hole != missingBloom && !errors.Is(err, index.ErrCorruptIndex) {
		t.Fatalf("expected corruption error, got %v", err)
	}
	checkMonitor(100, 1)
	_, quarantineErr := os.Stat(broken + ".corrupt")
	_, brokenErr := os.Stat(broken)
	if hole == corruptIndexInManifest && (quarantineErr != nil || brokenErr == nil) {
		t.Fatalf("expected chunk quarantined: %v %v", quarantineErr, brokenErr)
	}
	if hole == corruptIndexNoManifest && (quarantineErr == nil || brokenErr != nil) {
		t.Fatalf("expected chunk left in place: %v %v", quarantineErr, brokenErr)
	}
	writeChunk(100)
	if err := freshen(); err != nil {
		t.Fatal(err)
	}
	checkMonitor(300, 3)
}

func TestStopAtMissingBloom(t *testing.T) {
	rng := func(first, last base.Blknum) ranges.FileRange { return ranges.FileRange{First: first, Last: last} }
	tests := []struct {
		name      string
		published []string
		local     []ranges.FileRange
		keep      int
		wantErr   string
	}{
		{"nested stale bloom covers range", []string{"000000200-000000299"}, []ranges.FileRange{rng(0, 299), rng(100, 199), rng(300, 399)}, 3, ""},
		{"missing published chunk", []string{"000000100-000000199"}, []ranges.FileRange{rng(0, 99), rng(200, 299)}, 1, "missing bloom filter for published chunk 000000100-000000199"},
		{"malformed manifest range", []string{"000000100-bad"}, []ranges.FileRange{rng(0, 99), rng(200, 299)}, 0, `malformed manifest range "000000100-bad"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_CACHE_HOME", t.TempDir())
			const chain = "testchain"
			man := manifest.Manifest{Chain: chain}
			for _, r := range tc.published {
				man.Chunks = append(man.Chunks, types.ChunkRecord{Range: r})
			}
			manBytes, err := json.Marshal(man)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(config.PathToIndex(chain), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(config.PathToManifestFile(chain), manBytes, 0600); err != nil {
				t.Fatal(err)
			}
			jobs := make([]freshenJob, 0, len(tc.local))
			for _, r := range tc.local {
				jobs = append(jobs, freshenJob{rng: r})
			}
			updater := MonitorUpdate{Chain: chain, FirstBlock: 100}
			kept, err := updater.stopAtMissingBloom(jobs)
			if (err == nil) != (tc.wantErr == "") || (err != nil && err.Error() != tc.wantErr) {
				t.Fatalf("err=%v, want %q", err, tc.wantErr)
			}
			if len(kept) != tc.keep {
				t.Fatalf("kept %d jobs, want %d", len(kept), tc.keep)
			}
		})
	}
}
