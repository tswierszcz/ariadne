package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"ariadne/internal/httpcache"
)

// centralClient fetches artifacts and poms from a Maven repository
// (Maven Central by default: CVE:F6a reference jars, F6b and LIBRARY_MODE
// LB3 poms) politely:
//
//   - a file a local Maven repository (ARIADNE_MAVEN_LOCAL) holds is read
//     from disk, never fetched — only when that repository's
//     _remote.repositories says it was downloaded from a remote
//     repository (a locally installed artifact may differ from the
//     published one);
//   - requests carry ariadne's own User-Agent;
//   - a failed URL is not asked again in the same run, and once the
//     repository answers 429 (rate limited) nothing more is sent to it in
//     this run.
//
// Successful network responses go through the run's HTTP cache as before.
type centralClient struct {
	cache     *httpcache.Cache
	base      string // repository root the URLs start with
	local     string // local Maven repository root, "" = none
	userAgent string

	mu      sync.Mutex
	failed  map[string]error
	limited error
	stats   centralStats
}

type centralStats struct {
	Local, Network, Failed, Skipped int
}

var errRateLimited = errors.New("rate limited")

func (c *centralClient) fetch(ctx context.Context, url string) ([]byte, error) {
	if b, ok := c.fromLocal(url); ok {
		c.mu.Lock()
		c.stats.Local++
		c.mu.Unlock()
		return b, nil
	}
	c.mu.Lock()
	if err := c.failed[url]; err != nil {
		c.stats.Skipped++
		c.mu.Unlock()
		return nil, err
	}
	if c.limited != nil && strings.HasPrefix(url, c.base) {
		c.stats.Skipped++
		c.mu.Unlock()
		return nil, fmt.Errorf("%s not requested: %w", url, c.limited)
	}
	c.mu.Unlock()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	b, err := httpcache.Fetch(http.DefaultClient, c.cache, httpcache.Key("GET", url), httpcache.KindForURL(url), req)
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.stats.Failed++
		if c.failed == nil {
			c.failed = map[string]error{}
		}
		c.failed[url] = err
		if strings.Contains(err.Error(), "HTTP 429") && strings.HasPrefix(url, c.base) {
			c.limited = fmt.Errorf("%w: %s answered HTTP 429 earlier in this run", errRateLimited, c.base)
		}
		return nil, err
	}
	c.stats.Network++
	return b, nil
}

// fromLocal reads url's file from the local repository when it is there
// and was downloaded from a remote repository.
func (c *centralClient) fromLocal(url string) ([]byte, bool) {
	if c.local == "" || !strings.HasPrefix(url, c.base+"/") {
		return nil, false
	}
	rel := strings.TrimPrefix(url, c.base+"/")
	if strings.Contains(rel, "..") || strings.ContainsAny(rel, "?#") {
		return nil, false
	}
	p := filepath.Join(c.local, filepath.FromSlash(rel))
	if !downloadedFromRemote(filepath.Dir(p), filepath.Base(p)) {
		return nil, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return b, true
}

// downloadedFromRemote reads dir's _remote.repositories (Maven 3's
// enhanced local repository): "<file>><repoId>=" with a non-empty repoId
// is a download; an empty one is `mvn install`'s.
func downloadedFromRemote(dir, file string) bool {
	f, err := os.Open(filepath.Join(dir, "_remote.repositories"))
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		name, rest, ok := strings.Cut(l, ">")
		if !ok || name != file {
			continue
		}
		repo, _, _ := strings.Cut(rest, "=")
		return repo != ""
	}
	return false
}

// mavenLocal is ARIADNE_MAVEN_LOCAL: a local Maven repository root
// ("~/" expanded); unset or "" = none.
func mavenLocal(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			v = filepath.Join(h, v[2:])
		}
	}
	return v
}
