package hackernews

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/go-shiori/go-readability"
	"golang.org/x/sync/errgroup"
)

type API interface {
	Comment(story Story) (string, error)
	Story() ([]Story, error)
	Article(story Story) (string, error)
}

type HackerNews struct {
	urlStory   string
	urlItem    string
	urlWebItem string
	numStory   int
	client     *http.Client
	linkRegexp *regexp.Regexp // Keep this here to avoid recompiling each time
}

type Story struct {
	Rank       int    `json:"-"`
	ID         int    `json:"-"`
	By         string `json:"by"`
	PostTitle  string `json:"title"`
	URL        string `json:"url"`
	URLHost    string `json:"-"`
	Time       int    `json:"time"`
	TimeAgo    string `json:"-"`
	Kids       []int  `json:"kids"`
	Score      int    `json:"score"`
	NumComment int    `json:"descendants"`
}

type Comment struct {
	ID     int    `json:"id"`
	By     string `json:"by"`
	Text   string `json:"text"`
	Time   int64  `json:"time"`
	Kids   []int  `json:"kids"`
	Parent int    `json:"parent"`
	Level  int    `json:"-"`
}

func NewHackerNews(urlStory string, urlItem string, urlWebItem string, numStory int) *HackerNews {
	return &HackerNews{
		urlStory:   urlStory,
		urlItem:    urlItem,
		urlWebItem: urlWebItem,
		numStory:   numStory,
		client: &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       10 * time.Second,
		},
		linkRegexp: regexp.MustCompile(`\[(.*?)\]\((.*?)\)`),
	}
}

func (h *HackerNews) Story() ([]Story, error) {
	storyIDs, err := h.fetchStoryIDs()
	if err != nil {
		return []Story{}, err
	}

	g := errgroup.Group{}

	stories := make([]Story, len(storyIDs))
	buffer := make(chan Story, len(storyIDs))

	for rank, storyID := range storyIDs {
		g.Go(func() error {
			return h.fetchStory(storyID, rank, buffer)
		})
	}

	if err := g.Wait(); err != nil {
		return stories, fmt.Errorf("error decoding story: %w", err)
	}

	close(buffer)

	for story := range buffer {
		stories[story.Rank] = story
	}

	return stories, nil
}

func (h *HackerNews) Comment(story Story) (string, error) {
	var err error

	comments := map[int]Comment{}
	buffer := make(chan Comment, story.NumComment)

	go func() {
		err = h.fetchComments(story.Kids, 0, buffer)
		close(buffer)
	}()

	if err != nil {
		return "", fmt.Errorf("error fetching comment: %w", err)
	}

	for comment := range buffer {
		comments[comment.ID] = comment
	}

	res := ""
	for _, id := range story.Kids {
		res += h.formatComment(comments, id)
	}

	res = fmt.Sprintf("**| 🐮 %d Co(w)mments [%s]**%s", story.NumComment, fmt.Sprintf(h.urlWebItem, story.ID), res)

	return res, nil
}

func (h *HackerNews) Article(story Story) (string, error) {
	article, err := readability.FromURL(story.URL, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("error reading article: %w", err)
	}

	render, err := htmltomarkdown.ConvertString(article.Title + article.Content)
	if err != nil {
		return "", fmt.Errorf("error rendering article: %w", err)
	}

	return render, nil
}

// Define BubbleTerm functions here to avoid copying the struct into another one

func (s Story) Title() string {
	return fmt.Sprintf("%d. %s", s.Rank+1, s.PostTitle)
}

func (s Story) Description() string {
	return fmt.Sprintf("%d points | by %s | %s | %d comments | %s", s.Score, s.By, s.TimeAgo, s.NumComment, s.URLHost)
}

func (s Story) FilterValue() string { return s.PostTitle }

func (h *HackerNews) timeAgo(t time.Time) string {
	const hoursDay = 24

	now := time.Now()
	duration := now.Sub(t)

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < hoursDay*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	default:
		days := int(duration.Hours() / hoursDay)

		return fmt.Sprintf("%d days ago", days)
	}
}

func (h *HackerNews) fetchStory(storyID int, rank int, buffer chan Story) error {
	var story Story

	resp, err := h.client.Get(fmt.Sprintf(h.urlItem, storyID))
	if err != nil {
		return fmt.Errorf("error fetching story: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := json.NewDecoder(resp.Body).Decode(&story); err != nil {
		return fmt.Errorf("error decoding story: %w", err)
	}

	story.ID = storyID
	story.Rank = rank
	story.TimeAgo = h.timeAgo(time.Unix(int64(story.Time), 0))

	url, err := url.Parse(story.URL)
	if err != nil {
		return fmt.Errorf("error parsing url story: %w", err)
	}

	story.URLHost = url.Host

	buffer <- story

	return nil
}

func (h *HackerNews) fetchStoryIDs() ([]int, error) {
	var storyIDs []int

	resp, err := h.client.Get(h.urlStory)
	if err != nil {
		return storyIDs, fmt.Errorf("error fetching story: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := json.NewDecoder(resp.Body).Decode(&storyIDs); err != nil {
		return storyIDs, fmt.Errorf("error decoding story: %w", err)
	}

	if len(storyIDs) > h.numStory {
		storyIDs = storyIDs[:h.numStory]
	}

	return storyIDs, nil
}

func (h *HackerNews) fetchComment(commentID int, level int) (Comment, error) {
	var comment Comment

	resp, err := h.client.Get(fmt.Sprintf(h.urlItem, commentID))
	if err != nil {
		return comment, fmt.Errorf("error fetching comment: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return comment, fmt.Errorf("error decoding comment: %w", err)
	}

	comment.Text, err = htmltomarkdown.ConvertString(comment.Text)
	if err != nil {
		return comment, fmt.Errorf("error rendering comment: %w", err)
	}

	comment.Level = level

	return comment, nil
}

func (h *HackerNews) fetchComments(commentIDs []int, level int, buffer chan Comment) error {
	g := errgroup.Group{}

	for _, id := range commentIDs {
		g.Go(func() error {
			comment, err := h.fetchComment(id, level)
			if err != nil {
				return err
			}

			comment.Level = level
			buffer <- comment

			if len(comment.Kids) > 0 {
				return h.fetchComments(comment.Kids, level+1, buffer)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error fetching comment: %w", err)
	}

	return nil
}

func (h *HackerNews) formatComment(comments map[int]Comment, id int) string {
	var res string

	const gap = "\n\n"

	if comment, ok := comments[id]; ok {
		if comment.Text != "" {
			blockquotes := strings.Repeat(">", comment.Level+1)
			separator := ""
			header := ""

			if comment.Level == 0 {
				separator = gap + "---" + gap
				header = "# "
			}
			// Remove duplicated links created by htmltomarkdown lib
			comment.Text = h.linkRegexp.ReplaceAllString(comment.Text, "[link]($1)")
			res += fmt.Sprintf(separator+blockquotes+header+"%s | %s %s"+gap,
				comment.By,
				h.timeAgo(time.Unix(comment.Time, 0)),
				strings.ReplaceAll(gap+comment.Text, "\n", "\n"+blockquotes))
		}

		for _, id := range comment.Kids {
			res += h.formatComment(comments, id)
		}
	}

	return res
}
