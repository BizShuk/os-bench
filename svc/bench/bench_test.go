package bench

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunProducesPositiveMetricsForEveryCategory(t *testing.T) {
	opts := Options{Duration: 100 * time.Millisecond, Dir: t.TempDir(), DiskSizeMB: 8}
	var started []string
	results, err := Run(context.Background(), opts, nil, func(c string) { started = append(started, c) })
	require.NoError(t, err)

	assert.Equal(t, Categories(), started)
	seen := map[string]bool{}
	for _, r := range results {
		seen[r.Category] = true
		assert.Positive(t, r.Value, "%s/%s", r.Category, r.Name)
		assert.NotEmpty(t, r.Unit, "%s/%s", r.Category, r.Name)
		assert.Contains(t, []string{"higher", "lower"}, string(r.Better))
	}
	assert.Len(t, seen, len(Suite))
}

func TestRunOnlySelectedCategory(t *testing.T) {
	opts := Options{Duration: 50 * time.Millisecond}
	results, err := Run(context.Background(), opts, []string{"network"}, func(string) {})
	require.NoError(t, err)
	for _, r := range results {
		assert.Equal(t, "network", r.Category)
	}
}

func TestRunRejectsUnknownCategory(t *testing.T) {
	_, err := Run(context.Background(), Options{}, []string{"gpu"}, func(string) {})
	assert.ErrorContains(t, err, `unknown category "gpu"`)
}

func TestRunStopsWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(ctx, Options{Duration: time.Second}, []string{"cpu"}, func(string) {})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPercentile(t *testing.T) {
	samples := []time.Duration{5, 1, 4, 2, 3, 10, 9, 8, 7, 6}
	assert.Equal(t, time.Duration(5), percentile(samples, 50))
	assert.Equal(t, time.Duration(9), percentile(samples, 99))
	assert.Equal(t, time.Duration(10), percentile(samples, 100))
	assert.Equal(t, time.Duration(0), percentile(nil, 99))
}
