//go:build !darwin && !linux

package claudetui

import "errors"

// Do not enable the durable ledger on platforms without interprocess
// transaction fencing. The rest of the plugin remains functional normally.
func lockInteractiveToolJournal(string) (func(), error) {
	return nil, errors.New("durable tool journal requires an interprocess lock on this platform")
}
