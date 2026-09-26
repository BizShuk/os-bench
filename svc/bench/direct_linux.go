package bench

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// openDirect 以 O_DIRECT 開檔, 讀寫直接對裝置, 不經過 page cache;
// buffer, offset 與長度都必須對齊 DIRECT_ALIGN。不支援 O_DIRECT 的檔案系統會回傳錯誤。
func openDirect(path string, flag int) (*os.File, error) {
	f, err := os.OpenFile(path, flag|unix.O_DIRECT, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open with O_DIRECT (filesystem at --dir may not support direct I/O): %w", err)
	}
	return f, nil
}
