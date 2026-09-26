package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/bizshuk/os-bench/model"
)

// Entry 是資料目錄中的一份已存報告。
type Entry struct {
	Path   string
	Report model.Report
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// Save 把報告寫成 <dir>/<hostname>-<UTC 時間>.json, 回傳檔案路徑。
func Save(dir string, r model.Report) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("report.Save: %w", err)
	}
	name := fmt.Sprintf("%s-%s.json",
		unsafeName.ReplaceAllString(r.Host.Hostname, "_"),
		r.GeneratedAt.UTC().Format("20060102T150405Z"))
	path := filepath.Join(dir, name)
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("report.Save: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("report.Save: %w", err)
	}
	return path, nil
}

// Load 讀取單一報告檔。
func Load(path string) (model.Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Report{}, fmt.Errorf("report.Load: %w", err)
	}
	var r model.Report
	if err := json.Unmarshal(data, &r); err != nil {
		return model.Report{}, fmt.Errorf("report.Load %s: %w", path, err)
	}
	return r, nil
}

// List 回傳 dir 內所有報告, 最新的在前; 目錄不存在時回傳空清單。
func List(dir string) ([]Entry, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("report.List: %w", err)
	}
	entries := make([]Entry, 0, len(paths))
	for _, p := range paths {
		r, err := Load(p)
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{Path: p, Report: r})
	}
	slices.SortFunc(entries, func(a, b Entry) int {
		return b.Report.GeneratedAt.Compare(a.Report.GeneratedAt)
	})
	return entries, nil
}

// Resolve 把使用者給的參數轉成報告路徑: 既有檔案原樣使用, 否則視為 dir 內的檔名。
func Resolve(dir, arg string) string {
	if _, err := os.Stat(arg); err == nil {
		return arg
	}
	if !strings.HasSuffix(arg, ".json") {
		arg += ".json"
	}
	return filepath.Join(dir, arg)
}
