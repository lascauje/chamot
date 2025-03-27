package bubbleterm

import (
	"chamot/cmd/hackernews"
	"chamot/cmd/ollama"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	storyState int = iota
	commentState
	articleState
	chatState
)

type BubbleTerm struct {
	state      int
	story      *storyView
	comment    *commentView
	article    *articleView
	chat       *chatView
	hackerNews hackernews.API
	ollama     ollama.API
}

func NewBubbleTerm(hackerNews hackernews.API, ollama ollama.API) *BubbleTerm {
	stories, err := hackerNews.Story()
	if err != nil {
		log.Fatalf("error fetching story")
	}

	items := []list.Item{}

	for _, story := range stories {
		items = append(items, story)
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#FF6600", Dark: "#FF6600"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#FF6600", Dark: "#FF6600"}).
		Padding(0, 0, 0, 1)

	return &BubbleTerm{
		hackerNews: hackerNews,
		ollama:     ollama,
		state:      storyState,
		story:      newStoryView(style, items),
		comment:    newCommentView(),
		article:    newArticleView(),
		chat:       newChatView(style),
	}
}

func (b *BubbleTerm) Run() error {
	if _, err := tea.NewProgram(b, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("error running terminal app: %w", err)
	}

	return nil
}

func (b *BubbleTerm) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return b.ollama.Chat(b.chat.input, b.chat.output, b.chat.stop)
		},
		func() tea.Msg {
			return <-b.chat.output
		})
}

func (b *BubbleTerm) View() string {
	switch b.state {
	case storyState:
		return b.story.view()
	case commentState:
		return b.comment.view()
	case articleState:
		return b.article.view()
	case chatState:
		return b.chat.view()
	default:
		return ""
	}
}

func (b *BubbleTerm) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn // tea.Model is required by the lib
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return b, tea.Quit
		case "ctrl+x":
			if b.state == storyState {
				return b, tea.Quit
			}

			b.state = storyState
			b.article.gotoTop()
			b.comment.gotoTop()

			return b, cmd
		case "ctrl+]":
			if b.state == chatState {
				b.chat.stopChat()

				return b, cmd
			}
		case "enter":
			switch b.state {
			case storyState:
				b.state = commentState
				story := b.story.selected()

				comment, err := b.hackerNews.Comment(story)
				if err != nil {
					return b, tea.Quit
				}

				b.comment.setContent(comment)

				return b, cmd
			case chatState:
				b.chat.sendPrompt()
			}
		case " ":
			if b.state == storyState {
				b.state = articleState
				story := b.story.selected()

				article, err := b.hackerNews.Article(story)
				if err != nil {
					return b, tea.Quit
				}

				b.article.setContent(article)

				return b, cmd
			}
		case "o":
			if b.state == storyState {
				b.state = chatState
				cmd := b.chat.focus()

				return b, cmd
			}
		case "s":
			if b.state == storyState {
				story := b.story.selected()

				article, err := b.hackerNews.Article(story)
				if err != nil {
					return b, tea.Quit
				}

				b.chat.sendArticle(article)

				return b, cmd
			}
		case "g":
			switch b.state {
			case articleState:
				b.article.gotoTop()
			case commentState:
				b.comment.gotoTop()
			}
		case "G":
			switch b.state {
			case articleState:
				b.article.gotoBottom()
			case commentState:
				b.comment.gotoBottom()
			}
		}
	case ollama.Response:
		b.chat.formatResponse(msg, b.state == chatState)

		return b, func() tea.Msg { return <-b.chat.output }
	case tea.WindowSizeMsg:
		b.story.updateWindow(msg.Width, msg.Height)
		b.comment.updateWindow(msg.Width, msg.Height)
		b.article.updateWindow(msg.Width, msg.Height)
		b.chat.updateWindow(msg.Width, msg.Height)
	}

	switch b.state {
	case storyState:
		b.story.model, cmd = b.story.model.Update(msg)
		cmds = append(cmds, cmd)
	case commentState:
		b.comment.model, cmd = b.comment.model.Update(msg)
		cmds = append(cmds, cmd)
	case articleState:
		b.article.model, cmd = b.article.model.Update(msg)
		cmds = append(cmds, cmd)
	case chatState:
		b.chat.model, cmd = b.chat.model.Update(msg)
		cmds = append(cmds, cmd)
		b.chat.prompt, cmd = b.chat.prompt.Update(msg)
		cmds = append(cmds, cmd)
	}

	return b, tea.Batch(cmds...)
}
