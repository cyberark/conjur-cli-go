package pagination

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeItems returns count dummy strings labeled by their absolute position,
// starting at offset.
func makeItems(offset, count int) []string {
	items := make([]string, count)
	for i := 0; i < count; i++ {
		items[i] = fmt.Sprintf("item-%d", offset+i)
	}
	return items
}

// datasetFetcher returns a PageFetcher backed by a fixed-size in-memory
// dataset of `total` items, safe for concurrent use, plus a call counter.
func datasetFetcher(total int) (PageFetcher[string], *int32) {
	var calls int32
	var mu sync.Mutex
	fetch := func(offset, limit int) ([]string, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		if offset >= total {
			return nil, nil
		}
		end := offset + limit
		if end > total {
			end = total
		}
		return makeItems(offset, end-offset), nil
	}
	return fetch, &calls
}

func TestFetch_SmallResultSet_NoWorkerPool(t *testing.T) {
	fetch, calls := datasetFetcher(5)

	items, truncated, err := Fetch[string](context.Background(), fetch, Options{PageSize: 1000, Concurrency: 3, MaxResults: -1}, nil)

	require.NoError(t, err)
	assert.Equal(t, makeItems(0, 5), items)
	// A short, non-empty page 0 in unlimited mode is indistinguishable from
	// "server's own page cap is below what we requested" without another
	// round-trip, so Fetch reports truncated=true here even though this
	// particular dataset was genuinely exhausted. See
	// TestFetch_Unlimited_LowServerPageCap_ReportsTruncated for the case
	// this is actually guarding against.
	assert.True(t, truncated)
	assert.EqualValues(t, 1, *calls, "only page 0 should be fetched when it comes back short")
}

// TestFetch_Unlimited_LowServerPageCap_ReportsTruncated covers Fix 1 from
// the final review: a server whose api_resource_list_limit_max is below
// DefaultPageSize returns a short (but non-empty) page 0 even though far
// more data exists. Unlimited mode must report truncated=true so callers
// (list.go) surface their existing warning instead of silently dropping
// data.
func TestFetch_Unlimited_LowServerPageCap_ReportsTruncated(t *testing.T) {
	const serverCap = 100
	const total = 500

	fetch := func(offset, limit int) ([]string, error) {
		// Simulate a server that never returns more than serverCap items
		// per request, regardless of what limit was requested.
		effectiveLimit := limit
		if effectiveLimit > serverCap {
			effectiveLimit = serverCap
		}
		end := offset + effectiveLimit
		if end > total {
			end = total
		}
		if offset >= total {
			return nil, nil
		}
		return makeItems(offset, end-offset), nil
	}

	items, truncated, err := Fetch[string](context.Background(), fetch, Options{PageSize: 1000, Concurrency: 3, MaxResults: -1}, nil)

	require.NoError(t, err)
	assert.True(t, truncated, "short but non-empty page 0 in unlimited mode must be flagged as possibly truncated")
	assert.Len(t, items, serverCap, "only page 0 is fetched; Fetch does not retry with a smaller page size")
}

func TestFetch_BoundedPagination_ExactMultiple(t *testing.T) {
	fetch, _ := datasetFetcher(1_000_000) // effectively unbounded dataset
	items, truncated, err := Fetch[string](context.Background(), fetch, Options{PageSize: 10, Concurrency: 3, MaxResults: 30}, nil)

	require.NoError(t, err)
	assert.Equal(t, makeItems(0, 30), items)
	assert.True(t, truncated, "hit the cap with full pages, so more may exist")
}

func TestFetch_BoundedPagination_ServerExhaustedBeforeCap(t *testing.T) {
	fetch, _ := datasetFetcher(22)
	items, truncated, err := Fetch[string](context.Background(), fetch, Options{PageSize: 10, Concurrency: 3, MaxResults: 100}, nil)

	require.NoError(t, err)
	assert.Equal(t, makeItems(0, 22), items)
	assert.False(t, truncated, "server ran out before the cap, nothing was left out")
}

func TestFetch_Unlimited_DiscardsPagesPastTrueEnd(t *testing.T) {
	fetch, calls := datasetFetcher(25)
	items, truncated, err := Fetch[string](context.Background(), fetch, Options{PageSize: 10, Concurrency: 3, MaxResults: -1}, nil)

	require.NoError(t, err)
	assert.Equal(t, makeItems(0, 25), items)
	assert.False(t, truncated)
	assert.LessOrEqual(t, int(*calls), 4, "should not fetch far past the true end (page0 + 3 speculative pages max)")
}

// TestFetch_HugeMaxResults_SmallDataset_DoesNotOverAllocate covers Fix 2
// from the final review: opts.MaxResults is user-controlled (--limit) and
// can be huge (e.g. 2 billion). fetchRemaining must not attempt to
// pre-allocate a slice with that as capacity when the actual dataset is
// tiny; it should still return the correct, small result without
// panicking or exhausting memory.
func TestFetch_HugeMaxResults_SmallDataset_DoesNotOverAllocate(t *testing.T) {
	// PageSize is small enough that page 0 comes back full, so Fetch
	// proceeds into fetchRemaining -- the code path that used to
	// pre-allocate a slice with capacity == opts.MaxResults.
	fetch, _ := datasetFetcher(15)

	items, truncated, err := Fetch[string](context.Background(), fetch, Options{PageSize: 10, Concurrency: 3, MaxResults: 2_000_000_000}, nil)

	require.NoError(t, err)
	assert.Equal(t, makeItems(0, 15), items)
	assert.False(t, truncated)
}

func TestFetch_MaxResultsZero_NoFetch(t *testing.T) {
	fetch, calls := datasetFetcher(100)
	items, truncated, err := Fetch[string](context.Background(), fetch, Options{MaxResults: 0}, nil)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.False(t, truncated)
	assert.EqualValues(t, 0, *calls)
}

func TestFetch_ErrorMidFetch_PropagatesAndStopsScheduling(t *testing.T) {
	var calls int32
	var mu sync.Mutex
	fetch := func(offset, limit int) ([]string, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		if offset == 20 {
			return nil, errors.New("boom")
		}
		// A small delay on non-error pages gives the errored worker's
		// cancellation a chance to be observed by the others at their next
		// dispatchNext() check, instead of racing them to completion before
		// the error is even seen -- without this, "stops scheduling" can't
		// be asserted deterministically, since a bounded fetch's calls can
		// never exceed the page count anyway.
		time.Sleep(20 * time.Millisecond)
		return makeItems(offset, limit), nil
	}

	_, _, err := Fetch[string](context.Background(), fetch, Options{PageSize: 10, Concurrency: 3, MaxResults: 100}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
	// Scheduling must actually stop once the error is observed: page 0 plus
	// at most one in-flight page per worker beyond the failing one. With
	// Concurrency=3 and a bounded 10-page fetch, that's comfortably fewer
	// than all 10 pages -- a runaway scheduler (e.g. from a Fix 1/2
	// regression) would keep climbing toward that ceiling.
	assert.LessOrEqual(t, int(calls), 7, "worker pool should stop dispatching shortly after the error, not keep scheduling all pages")
}

func TestFetch_OnPage_CalledInStrictOrderDespiteConcurrency(t *testing.T) {
	fetch, _ := datasetFetcher(50)

	var mu sync.Mutex
	var seenOrder []int
	onPage := func(index int, items []string) error {
		mu.Lock()
		seenOrder = append(seenOrder, index)
		mu.Unlock()
		return nil
	}

	items, _, err := Fetch[string](context.Background(), fetch, Options{PageSize: 10, Concurrency: 3, MaxResults: -1}, onPage)

	require.NoError(t, err)
	assert.Equal(t, makeItems(0, 50), items)
	assert.True(t, sort.IntsAreSorted(seenOrder), "onPage must be called in ascending page order: %v", seenOrder)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, seenOrder)
}

func TestOptions_PageSizeEnvOverride(t *testing.T) {
	t.Setenv("CONJUR_LIST_PAGE_SIZE", "7")
	resolved := Options{}.resolved()
	assert.Equal(t, 7, resolved.PageSize)
	assert.Equal(t, 7, EffectivePageSize())
}

func TestOptions_PageSizeEnvOverride_InvalidFallsBackToDefault(t *testing.T) {
	t.Setenv("CONJUR_LIST_PAGE_SIZE", "not-a-number")
	resolved := Options{}.resolved()
	assert.Equal(t, DefaultPageSize, resolved.PageSize)

	t.Setenv("CONJUR_LIST_PAGE_SIZE", "0")
	resolved = Options{}.resolved()
	assert.Equal(t, DefaultPageSize, resolved.PageSize)
}

func TestParsePageSizeEnv_Unset(t *testing.T) {
	t.Setenv(pageSizeEnvVar, "")
	size, err := ParsePageSizeEnv()
	require.NoError(t, err)
	assert.Equal(t, 0, size)
}

func TestParsePageSizeEnv_Valid(t *testing.T) {
	t.Setenv(pageSizeEnvVar, "42")
	size, err := ParsePageSizeEnv()
	require.NoError(t, err)
	assert.Equal(t, 42, size)
}

func TestParsePageSizeEnv_InvalidReturnsError(t *testing.T) {
	t.Setenv(pageSizeEnvVar, "not-a-number")
	_, err := ParsePageSizeEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), pageSizeEnvVar)

	t.Setenv(pageSizeEnvVar, "0")
	_, err = ParsePageSizeEnv()
	require.Error(t, err)

	t.Setenv(pageSizeEnvVar, "-5")
	_, err = ParsePageSizeEnv()
	require.Error(t, err)
}

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
	cmd.SetArgs(nil)
	return cmd
}

func TestAddPageSizeFlag_IsHidden(t *testing.T) {
	cmd := newTestCmd()
	AddPageSizeFlag(cmd, 500)
	assert.True(t, cmd.Flags().Lookup("page-size").Hidden)
}

func TestResolvePageSize_DefaultsToCallerSuppliedDefault(t *testing.T) {
	cmd := newTestCmd()
	AddPageSizeFlag(cmd, 500)

	size, err := ResolvePageSize(cmd, 500)
	require.NoError(t, err)
	assert.Equal(t, 500, size)
}

func TestResolvePageSize_FlagOverridesDefault(t *testing.T) {
	cmd := newTestCmd()
	AddPageSizeFlag(cmd, 500)
	require.NoError(t, cmd.Flags().Set("page-size", "17"))

	size, err := ResolvePageSize(cmd, 500)
	require.NoError(t, err)
	assert.Equal(t, 17, size)
}

func TestResolvePageSize_RejectsNonPositiveFlag(t *testing.T) {
	cmd := newTestCmd()
	AddPageSizeFlag(cmd, 500)
	require.NoError(t, cmd.Flags().Set("page-size", "0"))

	_, err := ResolvePageSize(cmd, 500)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--page-size must be a positive integer")
}

func TestResolvePageSize_EnvOverridesCallerDefault(t *testing.T) {
	t.Setenv(pageSizeEnvVar, "9")
	cmd := newTestCmd()
	AddPageSizeFlag(cmd, 500)

	size, err := ResolvePageSize(cmd, 500)
	require.NoError(t, err)
	assert.Equal(t, 9, size)
}

func TestResolvePageSize_FlagTakesPrecedenceOverEnv(t *testing.T) {
	t.Setenv(pageSizeEnvVar, "9")
	cmd := newTestCmd()
	AddPageSizeFlag(cmd, 500)
	require.NoError(t, cmd.Flags().Set("page-size", "17"))

	size, err := ResolvePageSize(cmd, 500)
	require.NoError(t, err)
	assert.Equal(t, 17, size)
}

func TestResolvePageSize_InvalidEnvWarnsAndFallsBackToCallerDefault(t *testing.T) {
	t.Setenv(pageSizeEnvVar, "not-a-number")
	cmd := newTestCmd()
	AddPageSizeFlag(cmd, 500)
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	size, err := ResolvePageSize(cmd, 500)
	require.NoError(t, err)
	assert.Equal(t, 500, size)
	assert.Contains(t, stderr.String(), pageSizeEnvVar)
}

func TestAddPageSizeFlag_NonPositiveDefaultFallsBackToPackageDefault(t *testing.T) {
	cmd := newTestCmd()
	AddPageSizeFlag(cmd, 0)

	size, err := ResolvePageSize(cmd, 0)
	require.NoError(t, err)
	assert.Equal(t, DefaultPageSize, size)
}
