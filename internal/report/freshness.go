package report

import (
	"fmt"
	"strings"
	"time"

	"ariadne/internal/httpcache"
)

// Data-freshness states of a run (CVE:F2).
const (
	FreshnessComplete = "complete" // every expiring response was within the TTL or fetched now
	FreshnessStale    = "stale"    // some were older than the TTL (offline, or -allow-stale after a failed refresh)
	FreshnessUnknown  = "unknown"  // some had no recorded fetch time (cached before CVE:F2)
	FreshnessReplayed = "replayed" // every response came from a data snapshot (CVE:F3 -snapshot-in)
)

// DescribeFreshness turns the cache's per-kind counts into the run's
// freshness state, one report line per kind, and the warnings to show.
func DescribeFreshness(fr []httpcache.Freshness, ttl time.Duration) (state string, lines, warnings []string) {
	state = FreshnessComplete
	for _, f := range fr {
		var parts []string
		add := func(n int, what string) {
			if n > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", n, what))
			}
		}
		add(f.Fresh, "from cache")
		add(f.Fetched, "fetched")
		add(f.Refreshed, "refreshed")
		add(f.Revalidated, "revalidated (304)")
		add(f.Stale, "stale")
		add(f.UnknownAge, "of unknown age")
		line := fmt.Sprintf("%s: %s", f.Kind, strings.Join(parts, ", "))
		if !f.OldestValidated.IsZero() {
			line += " — oldest validated " + f.OldestValidated.Format(time.RFC3339)
		}
		lines = append(lines, line)
		if f.Stale > 0 {
			state = FreshnessStale
			warnings = append(warnings, fmt.Sprintf("data freshness: %d %s responses older than the %s TTL were used", f.Stale, f.Kind, ttl))
		}
		if f.UnknownAge > 0 {
			if state == FreshnessComplete {
				state = FreshnessUnknown
			}
			warnings = append(warnings, fmt.Sprintf("data freshness: %d %s responses have no recorded fetch time (cached before CVE:F2); run online to revalidate them", f.UnknownAge, f.Kind))
		}
		for _, e := range f.RefreshFailures {
			warnings = append(warnings, "data freshness: refresh failed, cached copy used (-allow-stale): "+e)
		}
	}
	return state, lines, warnings
}
