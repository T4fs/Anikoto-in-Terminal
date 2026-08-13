package scraper

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anitui/anitui/internal/models"
)

const (
	anidbBase = "https://anidb.app"
	anidbUA   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

type AnidbScraper struct {
	client *http.Client
}

func NewAnidbScraper() *AnidbScraper {
	return &AnidbScraper{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *AnidbScraper) Name() string {
	return "anidb.app"
}

func (s *AnidbScraper) get(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", anidbUA)
	req.Header.Set("Accept", "text/html,application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anidb.app returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(body), "Just a moment") {
		return nil, fmt.Errorf("anidb.app blocked the request (cloudflare challenge)")
	}
	return body, nil
}

func (s *AnidbScraper) Search(query string, dub bool) ([]models.Anime, error) {
	body, err := s.get(anidbBase + "/browse?q=" + url.QueryEscape(query))
	if err != nil {
		return nil, err
	}
	page := string(body)

	re := regexp.MustCompile(`anime/([a-z0-9-]+-[0-9]+)"[^>]*title="([^"]+)"`)
	seen := make(map[string]bool)
	animeList := make([]models.Anime, 0, 20)
	for _, m := range re.FindAllStringSubmatch(page, -1) {
		slug := m[1]
		if seen[slug] {
			continue
		}
		seen[slug] = true
		title := strings.TrimSpace(html.UnescapeString(m[2]))
		if title == "" {
			continue
		}
		animeList = append(animeList, models.Anime{
			Title:  title,
			URL:    slug,
			Source: s.Name(),
		})
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)
	for i := range animeList {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			detail := s.fetchDetail(animeList[idx].URL)
			if detail == nil {
				return
			}
			if detail.Title != "" {
				animeList[idx].Title = detail.Title
			}
			if detail.Score > 0 {
				animeList[idx].Score = detail.Score
			}
			if detail.EpisodeCount > 0 {
				animeList[idx].EpisodeCount = detail.EpisodeCount
			}
			animeList[idx].Synopsis = detail.Synopsis
			if len(detail.Genres) > 0 {
				animeList[idx].Genres = detail.Genres
			}
			if detail.Type != "" {
				animeList[idx].Type = detail.Type
			}
			animeList[idx].Studio = detail.Studio
			if detail.Status != "" {
				animeList[idx].Status = detail.Status
			}
			if detail.EpisodeCount > 0 {
				animeList[idx].Description = fmt.Sprintf("%d episodes", detail.EpisodeCount)
			}
		}(i)
	}
	wg.Wait()

	return animeList, nil
}

type anidbDetail struct {
	Title        string
	Score        float64
	EpisodeCount int
	Synopsis     string
	Genres       []string
	Type         string
	Studio       string
	Status       string
}

func (s *AnidbScraper) fetchDetail(slug string) *anidbDetail {
	body, err := s.get(anidbBase + "/anime/" + slug)
	if err != nil {
		return nil
	}
	page := string(body)

	detail := &anidbDetail{}

	if m := regexp.MustCompile(`"@type":\s*"(?:TVSeries|Movie|TVMovie)"[\s\S]*?"name":\s*"([^"]*)"`).FindStringSubmatch(page); m != nil {
		detail.Title = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	if m := regexp.MustCompile(`"description":\s*"([^"]*)"`).FindStringSubmatch(page); m != nil {
		detail.Synopsis = strings.TrimSpace(html.UnescapeString(strings.ReplaceAll(m[1], `\&#`, `&#`)))
	}
	if m := regexp.MustCompile(`"genre":\s*\[([^\]]*)\]`).FindStringSubmatch(page); m != nil {
		for _, g := range regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(m[1], -1) {
			gName := strings.TrimSpace(html.UnescapeString(g[1]))
			if gName != "" {
				detail.Genres = append(detail.Genres, gName)
			}
		}
	}
	if m := regexp.MustCompile(`href="/browse\?type=([^"]+)"[^>]*class="badge[^"]*"[^>]*>([^<]+)</a>`).FindStringSubmatch(page); m != nil {
		detail.Type = strings.TrimSpace(html.UnescapeString(m[2]))
	}
	if m := regexp.MustCompile(`href="/browse\?status=([^"]+)"[^>]*class="badge[^"]*"[^>]*>([^<]+)</a>`).FindStringSubmatch(page); m != nil {
		detail.Status = mapStatus(strings.TrimSpace(html.UnescapeString(m[2])))
	}
	if m := regexp.MustCompile(`<svg class="w-3 h-3 text-yellow-400"[\s\S]*?</svg>([0-9.]+)</span>`).FindStringSubmatch(page); m != nil {
		if f, err := strconv.ParseFloat(m[1], 64); err == nil {
			detail.Score = f
		}
	}
	if m := regexp.MustCompile(`href="/studios/[0-9]+"[^>]*>([^<]+)</a>`).FindStringSubmatch(page); m != nil {
		detail.Studio = strings.TrimSpace(html.UnescapeString(m[1]))
	}

	id := animeIDFromSlug(slug)
	if id != "" {
		if eps, err := s.fetchEpisodeIDs(id); err == nil {
			detail.EpisodeCount = len(eps)
		}
	}

	return detail
}

func animeIDFromSlug(slug string) string {
	m := regexp.MustCompile(`-(\d+)$`).FindStringSubmatch(slug)
	if m == nil {
		return ""
	}
	return m[1]
}

type anidbEpisodesResponse struct {
	Episodes []struct {
		ID     int `json:"id"`
		Number int `json:"number"`
		Filler bool `json:"filler"`
	} `json:"episodes"`
}

func (s *AnidbScraper) fetchEpisodeIDs(animeID string) ([]int, error) {
	body, err := s.get(fmt.Sprintf("%s/api/frontend/anime/%s/episodes", anidbBase, animeID))
	if err != nil {
		return nil, err
	}
	var result anidbEpisodesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse episodes: %w", err)
	}
	ids := make([]int, 0, len(result.Episodes))
	for _, e := range result.Episodes {
		ids = append(ids, e.ID)
	}
	return ids, nil
}

func (s *AnidbScraper) GetEpisodes(animeURL string, dub bool) ([]models.Episode, error) {
	id := animeIDFromSlug(animeURL)
	if id == "" {
		return nil, fmt.Errorf("could not extract anime id from %q", animeURL)
	}

	body, err := s.get(fmt.Sprintf("%s/api/frontend/anime/%s/episodes", anidbBase, id))
	if err != nil {
		return nil, err
	}
	var result anidbEpisodesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse episodes: %w", err)
	}

	episodes := make([]models.Episode, 0, len(result.Episodes))
	for _, e := range result.Episodes {
		episodes = append(episodes, models.Episode{
			Number: fmt.Sprintf("EP %d", e.Number),
			URL:    strconv.Itoa(e.ID),
		})
	}

	if dub {
		if len(episodes) == 0 {
			return episodes, nil
		}
		hasDub, err := s.hasLanguage(episodes[0].URL, "eng")
		if err != nil {
			return nil, err
		}
		if !hasDub {
			return nil, fmt.Errorf("no dub available for this anime")
		}
	}

	sort.Slice(episodes, func(i, j int) bool {
		ni, _ := strconv.ParseFloat(strings.TrimPrefix(episodes[i].Number, "EP "), 64)
		nj, _ := strconv.ParseFloat(strings.TrimPrefix(episodes[j].Number, "EP "), 64)
		return ni < nj
	})

	return episodes, nil
}

type anidbLanguagesResponse struct {
	Languages []struct {
		Code     string `json:"code"`
		Name     string `json:"name"`
		EmbedURL string `json:"embed_url"`
	} `json:"languages"`
}

func (s *AnidbScraper) hasLanguage(episodeID, code string) (bool, error) {
	body, err := s.get(fmt.Sprintf("%s/api/frontend/episode/%s/languages", anidbBase, episodeID))
	if err != nil {
		return false, err
	}
	var result anidbLanguagesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("parse languages: %w", err)
	}
	for _, l := range result.Languages {
		if l.Code == code {
			return true, nil
		}
	}
	return false, nil
}

func (s *AnidbScraper) GetVideoURL(episodeURL string, dub bool) ([]models.VideoSource, error) {
	lang := "jpn"
	if dub {
		lang = "eng"
	}

	body, err := s.get(fmt.Sprintf("%s/api/frontend/episode/%s/languages", anidbBase, episodeURL))
	if err != nil {
		return nil, err
	}
	var result anidbLanguagesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse languages: %w", err)
	}

	embed := ""
	for _, l := range result.Languages {
		if l.Code == lang {
			embed = l.EmbedURL
			break
		}
	}
	if embed == "" {
		return nil, fmt.Errorf("no %s audio track available for this episode", lang)
	}

	embedBody, err := s.get(embed)
	if err != nil {
		return nil, err
	}
	fileMatch := regexp.MustCompile(`file: '([^']*)'`).FindSubmatch(embedBody)
	if fileMatch == nil {
		return nil, fmt.Errorf("could not find stream url on embed page")
	}
	masterURL := string(fileMatch[1])
	if !strings.HasPrefix(masterURL, "http") {
		masterURL = anidbBase + masterURL
	}

	playlistBody, err := s.get(masterURL)
	if err != nil {
		return nil, err
	}

	sources := parseAnidbVariants(string(playlistBody), masterURL)
	if len(sources) == 0 {
		sources = []models.VideoSource{{URL: masterURL, Quality: "default", Type: "hls"}}
	}

	return sources, nil
}

var anidbVariantRe = regexp.MustCompile(`(?m)^#EXT-X-STREAM-INF:[^\n]*RESOLUTION=(\d+)x(\d+)[^\n]*\n\s*(https?://\S+)`)

func parseAnidbVariants(playlist, masterURL string) []models.VideoSource {
	sources := make([]models.VideoSource, 0, 4)
	for _, m := range anidbVariantRe.FindAllStringSubmatch(playlist, -1) {
		height := m[2]
		sources = append(sources, models.VideoSource{
			URL:     strings.TrimSpace(m[3]),
			Quality: height + "p",
			Type:    "hls",
		})
	}

	sort.Slice(sources, func(i, j int) bool {
		hi, _ := strconv.Atoi(strings.TrimSuffix(sources[i].Quality, "p"))
		hj, _ := strconv.Atoi(strings.TrimSuffix(sources[j].Quality, "p"))
		return hi > hj
	})

	return sources
}