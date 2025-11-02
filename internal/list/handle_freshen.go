package listPkg

import "github.com/TrueBlocks/trueblocks-chifra/v6/pkg/monitor"

func (opts *ListOptions) FreshenMonitorsForList(monitorArray *[]monitor.Monitor) (bool, error) {
	var updater = monitor.NewUpdater(opts.Globals.Chain, opts.Globals.TestMode, false /* skipFreshen */, opts.Addrs)
	updater.PublisherAddr = opts.PublisherAddr
	return updater.FreshenMonitors(monitorArray)
}
