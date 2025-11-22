package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/abi"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/decache"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/file"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/logger"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/rpc"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/types"
	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/walk"
)

func (mon *Monitor) Decache(conn *rpc.Connection, showProgress bool) (string, error) {
	if apps, cnt, err := mon.ReadAndFilterAppearances(types.NewEmptyFilter(), true /* withCount */); err != nil {
		return "", err
	} else {
		if cnt > 0 {
			monitorCacheTypes := []walk.CacheType{
				// walk.Cache_Statements, see  below
				walk.Cache_Traces,
				walk.Cache_Transactions,
				walk.Cache_Receipts,
				walk.Cache_Withdrawals,
			}
			for _, cacheType := range monitorCacheTypes {
				itemsToRemove, err := decache.LocationsFromAddressAndAppearances(mon.Address, apps, cacheType)
				if err != nil {
					return "", err
				}

				if msg, err := decache.Decache(conn, itemsToRemove, showProgress, cacheType); err != nil {
					return "", err

				} else {
					logger.Progress(showProgress, msg)
				}
			}
		}

		if err := mon.RemoveStatements(len(apps), showProgress); err != nil {
			return "", err
		}

		abiPath := abi.PathToAbisCache(conn.Chain, mon.Address.Hex())
		if file.FileExists(abiPath) {
			os.Remove(abiPath)
			logger.Progress(showProgress, "Abi "+abiPath+" file removed.")
		}

		return mon.RemoveMonitor(), nil
	}
}

func (mon *Monitor) RemoveMonitor() string {
	clear := strings.Repeat(" ", 30)
	mon.Close()
	mon.Delete()
	if wasRemoved, err := mon.Remove(); !wasRemoved || err != nil {
		return fmt.Sprintf("Monitor for %s was not removed (%s).%s", mon.Address.Hex(), err.Error(), clear)
	} else {
		return fmt.Sprintf("Monitor for %s was permanently removed.%s", mon.Address.Hex(), clear)
	}
}

func (mon *Monitor) GetRemoveWarning() string {
	addr := mon.Address.Hex()
	count := mon.Count()
	var warning = strings.ReplaceAll("Are sure you want to decache {0}{1} (Yn)?", "{0}", addr)
	if count > 5000 {
		return strings.ReplaceAll(strings.ReplaceAll(warning, "{1}", ". It may take a long time to process {2} records."), "{2}", fmt.Sprintf("%d", count))
	}
	return strings.ReplaceAll(warning, "{1}", "")
}

func (mon *Monitor) RemoveStatements(l int, showProgress bool) error {
	removedCount := 0

	bar := logger.NewBar(logger.BarOptions{
		Prefix:  "Removing statements cache files",
		Enabled: showProgress,
		Type:    logger.Expanding,
		Total:   int64(l), // Initial estimate based on appearances
	})

	// Use the hierarchical directory structure to optimize scanning
	// Address f503017d... creates directories like statements/f5/03/01/
	addrHex := strings.ToLower(mon.Address.Hex()[2:]) // Remove 0x prefix
	targetDirs := []string{
		filepath.Join(walk.GetRootPathFromCacheType(mon.Chain, walk.Cache_Statements), addrHex[:2], addrHex[2:4], addrHex[4:6]),
	}

	for _, targetDir := range targetDirs {
		if !file.FolderExists(targetDir) {
			continue
		}

		visitFunc := func(path string, vP any) (bool, error) {
			_ = vP
			// Since we're only walking directories that match our address prefix,
			// we still need to check the filename starts with our address
			filename := filepath.Base(path)
			if strings.HasPrefix(filename, addrHex) && strings.HasSuffix(filename, ".bin") {
				if err := os.Remove(path); err == nil {
					removedCount++
				}
				// TODO: Clean empty folders here
			}
			bar.Tick()
			return true, nil
		}

		err := walk.ForEveryFileInFolder(targetDir, visitFunc, nil)
		if err != nil {
			bar.Finish(true)
			return err
		}
	}

	bar.Finish(true)

	// Log the actual number of files removed
	logger.Progress(showProgress, fmt.Sprintf("%d statement files removed for %s", removedCount, mon.Address.Hex()))

	return nil
}
