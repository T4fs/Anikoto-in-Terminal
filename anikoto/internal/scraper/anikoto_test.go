package scraper

import (
	"testing"
	"time"
)

func TestAnikotoPipeline(t *testing.T) {
	s := NewAnikotoScraper()

	t.Log("Searching for naruto...")
	t0 := time.Now()
	results, err := s.Search("naruto", false)
	t.Logf("Search: %d results, err=%v, took=%v", len(results), err, time.Since(t0))
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("No search results")
	}

	found := false
	for _, r := range results {
		t.Logf("Result: %q | %s | score=%.1f year=%d eps=%d type=%s", r.Title, r.URL, r.Score, r.Year, r.EpisodeCount, r.Type)
		if r.Title == "Naruto" {
			found = true
			break
		}
	}
	if !found {
		t.Log("Naruto not in top results (not failing, listing all above)")
	}

	r := results[0]
	t.Log("Loading episodes...")
	t1 := time.Now()
	eps, err := s.GetEpisodes(r.URL, false)
	t.Logf("Episodes: %d, err=%v, took=%v", len(eps), err, time.Since(t1))
	if err != nil {
		t.Fatalf("GetEpisodes error: %v", err)
	}
	if len(eps) == 0 {
		t.Fatal("No episodes")
	}
	t.Logf("First episode: %s | %s", eps[0].Number, eps[0].Title)

	t.Log("Fetching video URL...")
	t2 := time.Now()
	sources, err := s.GetVideoURL(eps[0].URL, false)
	t.Logf("Sources: %d, err=%v, took=%v", len(sources), err, time.Since(t2))
	if err != nil {
		t.Fatalf("GetVideoURL error: %v", err)
	}
	for i, src := range sources {
		t.Logf("  [%d] %s %s: %s", i, src.Quality, src.Type, src.URL)
	}
	if len(sources) == 0 {
		t.Fatal("No video sources found")
	}
}