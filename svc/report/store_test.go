package report

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/bizshuk/os-bench/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveListLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	older := model.Report{GeneratedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Host: model.Host{Hostname: "a.local"}}
	newer := model.Report{GeneratedAt: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), Host: model.Host{Hostname: "b/../x"},
		Results: []model.Result{{Category: "cpu", Name: "compress_single", Value: 1.5, Unit: "MB/s", Better: model.BETTER_HIGHER}}}

	_, err := Save(dir, older)
	require.NoError(t, err)
	path, err := Save(dir, newer)
	require.NoError(t, err)
	assert.Equal(t, dir, filepath.Dir(path), "hostname must not escape the data dir")

	entries, err := List(dir)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "b/../x", entries[0].Report.Host.Hostname, "newest first")
	assert.Equal(t, newer.Results, entries[0].Report.Results)

	loaded, err := Load(Resolve(dir, filepath.Base(path[:len(path)-len(".json")])))
	require.NoError(t, err)
	assert.Equal(t, newer.Host.Hostname, loaded.Host.Hostname)
}

func TestListMissingDirIsEmpty(t *testing.T) {
	entries, err := List(filepath.Join(t.TempDir(), "absent"))
	require.NoError(t, err)
	assert.Empty(t, entries)
}
