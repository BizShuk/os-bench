package bench

import (
	"os"

	"golang.org/x/sys/unix"
)

// openDirect 開檔後設定 F_NOCACHE, 此 fd 的讀寫都不經過 unified buffer cache。
func openDirect(path string, flag int) (*os.File, error) {
	f, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		return nil, err
	}
	if _, err := unix.FcntlInt(f.Fd(), unix.F_NOCACHE, 1); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
