package report

import (
	"strings"
	"testing"
	"time"

	"ariadne/internal/httpcache"
)

func TestDescribeFreshness(t *testing.T) {
	if s, _, w := DescribeFreshness([]httpcache.Freshness{{Kind: httpcache.KindOSV, Fresh: 3, Fetched: 1}}, time.Hour); s != FreshnessComplete || len(w) != 0 {
		t.Errorf("complete: %s %v", s, w)
	}
	s, lines, w := DescribeFreshness([]httpcache.Freshness{{Kind: httpcache.KindOSV, Fresh: 1, UnknownAge: 5}}, time.Hour)
	if s != FreshnessUnknown || len(w) != 1 || !strings.Contains(lines[0], "5 of unknown age") {
		t.Errorf("unknown: %s %v %v", s, lines, w)
	}
	s, _, w = DescribeFreshness([]httpcache.Freshness{
		{Kind: httpcache.KindMutable, UnknownAge: 1},
		{Kind: httpcache.KindOSV, Stale: 2, RefreshFailures: []string{"HTTP 503"}},
	}, time.Hour)
	if s != FreshnessStale || len(w) != 3 {
		t.Errorf("stale wins over unknown: %s %v", s, w)
	}
}
