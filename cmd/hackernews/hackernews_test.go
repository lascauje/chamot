package hackernews

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStory(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v0/topstories.json" {
			mockStoryIDs := []int{43332658, 16582136, 40077533}
			jsonData, _ := json.Marshal(mockStoryIDs)
			_, _ = w.Write(jsonData)

			return
		}

		if r.URL.Path == "/v0/item/43332658.json" {
			story := Story{
				By:         "btilly",
				NumComment: 3,
				ID:         43332658,
				Kids:       []int{43339316, 43335480},
				Score:      1421,
				Time:       1741702475,
				PostTitle:  "Happy 20th birthday, Y Combinator",
				URL:        "ycombinator.com",
				Rank:       1,
				TimeAgo:    "x days ago",
				URLHost:    "ycombinator.com",
			}
			jsonData, _ := json.Marshal(story)
			_, _ = w.Write(jsonData)

			return
		}

		if r.URL.Path == "/v0/item/16582136.json" {
			story := Story{
				By:         "Cogito",
				NumComment: 436,
				ID:         16582136,
				Kids:       []int{16582247},
				Score:      6015,
				Time:       1520999430,
				PostTitle:  "Stephen Hawking has died",
				URL:        "",
				Rank:       1,
				URLHost:    "",
				TimeAgo:    "x days ago",
			}
			jsonData, _ := json.Marshal(story)
			_, _ = w.Write(jsonData)

			return
		}

		if r.URL.Path == "/v0/item/40077533.json" {
			story := Story{
				By:         "bratao",
				NumComment: 923,
				ID:         40077533,
				Kids:       []int{},
				Score:      2199,
				Time:       1713455842,
				PostTitle:  "Meta Llama 3",
				URL:        "llama.meta.com",
				Rank:       1,
				URLHost:    "llama.meta.com",
				TimeAgo:    "x days ago",
			}
			jsonData, _ := json.Marshal(story)
			_, _ = w.Write(jsonData)

			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	h := NewHackerNews(mockServer.URL+"/v0/topstories.json",
		mockServer.URL+"/v0/item/%d.json", mockServer.URL+"/item?id=%d", 3)

	storyIDs, _ := h.Story()

	tests := []struct {
		expectedCount int
	}{
		{expectedCount: 3},
	}

	for _, tt := range tests {
		t.Run("Check number of story IDs", func(t *testing.T) {
			if got := len(storyIDs); got != tt.expectedCount {
				t.Errorf("len(storyIDs) = %v; want %v", got, tt.expectedCount)
			}
		})
	}
}

func TestComment(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v0/item/43339316.json" {
			comment := Comment{
				By:     "Dave_Rosenthal",
				ID:     43339316,
				Parent: 43332658,
				Kids:   []int{43340657},
				Text:   "I worked with pg in a three-person startup in a basement in Harvard square just before he started YC",
				Time:   1741746022,
				Level:  1,
			}
			jsonData, _ := json.Marshal(comment)
			_, _ = w.Write(jsonData)

			return
		}

		if r.URL.Path == "/v0/item/43340657.json" {
			comment := Comment{
				By:     "knuckleheadsmif",
				ID:     43340657,
				Parent: 43339316,
				Text:   "In the mid 90s (1996?) So myself and 3 others worked for Intuit and we traveled, from Mountain View CA",
				Time:   1741764049,
				Level:  2,
				Kids:   []int{},
			}
			jsonData, _ := json.Marshal(comment)
			_, _ = w.Write(jsonData)

			return
		}

		if r.URL.Path == "/v0/item/43335480.json" {
			comment := Comment{
				By:     "CSMastermind",
				ID:     43335480,
				Parent: 43332658,
				Text:   "I remember the initial PG announcement about the founder",
				Time:   1741717262,
				Kids:   []int{},
				Level:  1,
			}
			jsonData, _ := json.Marshal(comment)
			_, _ = w.Write(jsonData)

			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	story := Story{
		By:         "btilly",
		NumComment: 3,
		ID:         43332658,
		Kids:       []int{43339316, 43335480},
		Score:      1421,
		Time:       1741702475,
		PostTitle:  "Happy 20th birthday, Y Combinator",
		URL:        "/ycombinator",
		Rank:       1,
		TimeAgo:    "x days ago",
		URLHost:    "ycombinator.com",
	}

	h := NewHackerNews(mockServer.URL+"/v0/topstories.json",
		mockServer.URL+"/v0/item/%d.json", mockServer.URL+"/item?id=%d", 3)

	comment, _ := h.Comment(story)

	tests := []struct {
		substring string
		expected  bool
	}{
		{"Dave_Rosenthal", true},
		{"knuckleheadsmif", true},
		{"CSMastermind", true},
	}

	for _, tt := range tests {
		t.Run(tt.substring, func(t *testing.T) {
			if got := strings.Contains(comment, tt.substring); got != tt.expected {
				t.Errorf("strings.Contains(comment, %s) = %v; want %v", tt.substring, got, tt.expected)
			}
		})
	}
}

func TestArticle(t *testing.T) {
	httpContent := `<html><head><title>Hacker News</title></head><body><h1>new | past | comments | ask | show | jobs | submit</h1></body></html>` //nolint: lll // ...

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(httpContent))
	}))
	defer mockServer.Close()

	h := NewHackerNews(mockServer.URL+"/v0/topstories.json",
		mockServer.URL+"/v0/item/%d.json", mockServer.URL+"/item?id=%d", 3)

	story := Story{
		By:         "btilly",
		NumComment: 3,
		ID:         43332658,
		Kids:       []int{43339316, 43335480},
		Score:      1421,
		Time:       1741702475,
		PostTitle:  "Happy 20th birthday, Y Combinator",
		URL:        mockServer.URL + "/ycombinator",
		Rank:       1,
		URLHost:    "/ycombinator",
		TimeAgo:    "x days ago",
	}

	article, _ := h.Article(story)

	tests := []struct {
		substring string
		expected  bool
	}{
		{"new | past | comments | ask | show | jobs | submit", true},
	}

	for _, tt := range tests {
		t.Run(tt.substring, func(t *testing.T) {
			if got := strings.Contains(article, tt.substring); got != tt.expected {
				t.Errorf("strings.Contains(article, %s) = %v; want %v", tt.substring, got, tt.expected)
			}
		})
	}
}
