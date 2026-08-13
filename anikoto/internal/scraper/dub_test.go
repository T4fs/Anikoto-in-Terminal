package scraper

import (
	"testing"
)

func TestAnikotoDub(t *testing.T) {
	s := NewAnikotoScraper()
	eps, err := s.GetEpisodes("naruto-eybxz", true)
	if err != nil {
		t.Fatalf("GetEpisodes dub error: %v", err)
	}
	t.Logf("dub episodes: %d (first: %s | %s)", len(eps), eps[0].Number, eps[0].Title)
	if len(eps) == 0 {
		t.Fatal("no dub episodes")
	}
	sources, err := s.GetVideoURL(eps[0].URL, true)
	if err != nil {
		t.Fatalf("GetVideoURL dub error: %v", err)
	}
	for i, src := range sources {
		t.Logf("  [%d] %s", i, src.URL)
	}
	if len(sources) == 0 {
		t.Fatal("no dub video sources")
	}
}