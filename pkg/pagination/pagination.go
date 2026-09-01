// Package pagination provides a generic, concurrent pagination helper for
// commands that need to fetch more items than a single server-side page
// allows.
package pagination

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	// DefaultPageSize is the default number of items requested per page,
	// matching the Conjur server's default api_resource_list_limit_max.
	DefaultPageSize = 1000

	// DefaultConcurrency is the default number of pages fetched in parallel.
	DefaultConcurrency = 3

	pageSizeEnvVar = "CONJUR_LIST_PAGE_SIZE"
)

// PageFetcher fetches a single page of items at the given offset/limit and
// must be safe for concurrent use.
type PageFetcher[T any] func(offset, limit int) ([]T, error)

// OnPageFunc is invoked in strict offset order as each page becomes
// available. A nil OnPageFunc is valid for callers who only need the final
// assembled slice.
type OnPageFunc[T any] func(pageIndex int, items []T) error

// Options configures Fetch's paging behavior.
type Options struct {
	// PageSize is the number of items requested per page. If <= 0, it
	// defaults to DefaultPageSize, then to CONJUR_LIST_PAGE_SIZE if that
	// env var is set to a valid positive integer.
	PageSize int
	// Concurrency is the number of pages fetched in parallel. If <= 0, it
	// defaults to DefaultConcurrency.
	Concurrency int
	// MaxResults caps the total number of items fetched:
	//   -1  unlimited (fetch until the server is exhausted)
	//    0  fetch nothing (PageFetcher is never called)
	//   N>0 cap the total at N
	MaxResults int
}

func (o Options) resolved() Options {
	if o.PageSize <= 0 {
		o.PageSize = DefaultPageSize
		if envSize, err := ParsePageSizeEnv(); err == nil && envSize > 0 {
			o.PageSize = envSize
		}
	}
	if o.Concurrency <= 0 {
		o.Concurrency = DefaultConcurrency
	}
	return o
}

// ParsePageSizeEnv reads and validates the CONJUR_LIST_PAGE_SIZE env var.
// It returns (0, nil) if the env var is unset, (size, nil) if it's set to a
// valid positive integer, and (0, err) if it's set to anything else —
// callers that surface output to a user (unlike the internal, silent
// fallback in resolved()) should report err rather than swallowing it.
func ParsePageSizeEnv() (int, error) {
	v := os.Getenv(pageSizeEnvVar)
	if v == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", pageSizeEnvVar, v)
	}
	return parsed, nil
}

// EffectivePageSize returns the page size Fetch uses when Options.PageSize
// is left at its zero value: DefaultPageSize, or CONJUR_LIST_PAGE_SIZE if
// that env var is set to a valid positive integer. Callers that need to
// compute a default cap in terms of "N pages" should use this instead of
// the DefaultPageSize constant directly, so the cap tracks any env
// override.
func EffectivePageSize() int {
	return Options{}.resolved().PageSize
}

// AddPageSizeFlag registers the --page-size flag on cmd. defaultPageSize is
// the page size an endpoint uses absent any override (server endpoints
// don't all share the same per-request cap); if <= 0, DefaultPageSize is
// used. The flag is hidden: it's an escape hatch for tuning against
// non-default server configurations, not something most callers — and
// essentially no SaaS customers — need to touch.
func AddPageSizeFlag(cmd *cobra.Command, defaultPageSize int) {
	if defaultPageSize <= 0 {
		defaultPageSize = DefaultPageSize
	}
	cmd.Flags().Int("page-size", 0, fmt.Sprintf("Sets the number of items requested per page when auto-paginating. Must be a positive integer. If omitted, defaults to %s, then to %d.", pageSizeEnvVar, defaultPageSize))
	cmd.Flags().Lookup("page-size").Hidden = true
}

// ResolvePageSize determines the page size to request per page for cmd, in
// order of precedence: an explicit --page-size flag (added via
// AddPageSizeFlag), then CONJUR_LIST_PAGE_SIZE, then defaultPageSize (or
// DefaultPageSize if defaultPageSize <= 0). An explicit --page-size that
// isn't a positive integer is a validation error; an invalid
// CONJUR_LIST_PAGE_SIZE is instead reported as a warning on cmd's stderr
// (matching the historical, permissive env var behavior) before falling
// back to the default.
func ResolvePageSize(cmd *cobra.Command, defaultPageSize int) (int, error) {
	if defaultPageSize <= 0 {
		defaultPageSize = DefaultPageSize
	}

	if cmd.Flags().Changed("page-size") {
		flagSize, err := cmd.Flags().GetInt("page-size")
		if err != nil {
			return 0, err
		}
		if flagSize <= 0 {
			return 0, errors.New("--page-size must be a positive integer")
		}
		return flagSize, nil
	}

	envSize, err := ParsePageSizeEnv()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v; using the default page size of %d.\n", err, defaultPageSize)
		return defaultPageSize, nil
	}
	if envSize > 0 {
		return envSize, nil
	}
	return defaultPageSize, nil
}

type page[T any] struct {
	items []T
}

// Fetch retrieves items via fetch, paginating and parallelizing according
// to opts, and returns them assembled in strict offset order. truncated is
// true when MaxResults was a positive cap and the fetch stopped because of
// that cap (every page up to the cap came back full) rather than because
// the server ran out of data. It is also true in unlimited mode
// (MaxResults < 0) when page 0 alone came back short and non-empty, since
// that's indistinguishable from the server's own per-request page cap
// (e.g. api_resource_list_limit_max) being lower than the page size
// requested — see the comment at that check for details.
func Fetch[T any](ctx context.Context, fetch PageFetcher[T], opts Options, onPage OnPageFunc[T]) ([]T, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	opts = opts.resolved()

	if opts.MaxResults == 0 {
		return make([]T, 0), false, nil
	}

	firstPageLimit := opts.PageSize
	if opts.MaxResults > 0 && opts.MaxResults < firstPageLimit {
		firstPageLimit = opts.MaxResults
	}

	page0Items, err := fetch(0, firstPageLimit)
	if err != nil {
		return nil, false, err
	}

	if onPage != nil {
		if err := onPage(0, page0Items); err != nil {
			return nil, false, err
		}
	}

	if len(page0Items) < firstPageLimit {
		// In unlimited mode, a non-empty but short page 0 doesn't necessarily
		// mean the server is out of data at position 0 — it may mean the
		// server's own page-size cap (e.g. api_resource_list_limit_max) is
		// lower than the page size we requested. We can't distinguish those
		// two cases here without another round-trip, so err on the side of
		// warning: report truncated so callers (list.go) surface the
		// existing "may be truncated" warning rather than silently dropping
		// data. A genuinely empty page 0 still means "no data at all".
		truncated := opts.MaxResults < 0 && len(page0Items) > 0
		return page0Items, truncated, nil
	}
	if opts.MaxResults > 0 && len(page0Items) >= opts.MaxResults {
		return page0Items, true, nil
	}

	return fetchRemaining(ctx, fetch, opts, page0Items, onPage)
}

// fetchRemaining dispatches pages 1..N through a bounded worker pool and
// reassembles them in strict order, discarding any speculatively-fetched
// pages past the true end of the data.
func fetchRemaining[T any](ctx context.Context, fetch PageFetcher[T], opts Options, page0Items []T, onPage OnPageFunc[T]) ([]T, bool, error) {
	bounded := opts.MaxResults > 0

	var (
		mu             sync.Mutex
		results        = map[int]page[T]{}
		nextToDispatch = 1
		knownLastIndex = -1 // -1 = not yet discovered
		shortPageSeen  = false
	)

	if bounded {
		totalPages := (opts.MaxResults + opts.PageSize - 1) / opts.PageSize
		knownLastIndex = totalPages - 1
	}

	g, gctx := errgroup.WithContext(ctx)

	dispatchNext := func() int {
		mu.Lock()
		defer mu.Unlock()
		if knownLastIndex >= 0 && nextToDispatch > knownLastIndex {
			return -1
		}
		idx := nextToDispatch
		nextToDispatch++
		return idx
	}

	worker := func() error {
		for {
			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}

			idx := dispatchNext()
			if idx == -1 {
				return nil
			}

			limit := opts.PageSize
			offset := idx * opts.PageSize
			if bounded {
				if remaining := opts.MaxResults - offset; remaining < limit {
					limit = remaining
				}
			}

			items, err := fetch(offset, limit)
			if err != nil {
				return err
			}

			mu.Lock()
			results[idx] = page[T]{items: items}
			if len(items) < limit {
				shortPageSeen = true
				if knownLastIndex < 0 || idx < knownLastIndex {
					knownLastIndex = idx
				}
			}
			mu.Unlock()
		}
	}

	for i := 0; i < opts.Concurrency; i++ {
		g.Go(worker)
	}

	if err := g.Wait(); err != nil {
		return nil, false, err
	}

	assembled := page0Items
	if bounded {
		// opts.MaxResults comes straight from user input (e.g. --limit
		// 2000000000) and can be arbitrarily large. Pre-allocating a slice
		// with that as capacity risks a multi-gigabyte allocation for a
		// request that may only ever return a handful of real items. Cap
		// the pre-allocation at a reasonable near-term buffer (page size
		// times the number of pages that can be in flight at once, plus
		// page 0); anything beyond that just grows normally via append.
		prealloc := min(opts.MaxResults, opts.PageSize*(opts.Concurrency+1))
		grown := make([]T, len(assembled), prealloc)
		copy(grown, assembled)
		assembled = grown
	}

	for idx := 1; knownLastIndex < 0 || idx <= knownLastIndex; idx++ {
		p, ok := results[idx]
		if !ok {
			break
		}
		if len(p.items) == 0 {
			// Skip empty pages (reached end of data)
			break
		}
		assembled = append(assembled, p.items...)
		if onPage != nil {
			if err := onPage(idx, p.items); err != nil {
				return nil, false, err
			}
		}
	}

	truncated := bounded && !shortPageSeen
	return assembled, truncated, nil
}
