package bubbleterm

import (
	"chamot/cmd/hackernews"
	"chamot/cmd/ollama"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type storyView struct {
	style lipgloss.Style
	model list.Model
}

type commentView struct {
	style lipgloss.Style
	model viewport.Model
}

type articleView struct {
	style lipgloss.Style
	model viewport.Model
}

type chatView struct {
	model     viewport.Model
	messages  []string
	prompt    textarea.Model
	responses []string
	input     chan string
	output    chan ollama.Response
	stop      chan bool
}

func newStoryView(style lipgloss.Style, items []list.Item) *storyView {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = style
	delegate.Styles.SelectedDesc = style

	model := list.New(items, delegate, 0, 0)
	model.Styles.Title = lipgloss.NewStyle().Bold(true)
	model.Title = "🐫 Chamot"
	model.SetFilteringEnabled(false)
	model.SetShowHelp(false)
	model.SetShowStatusBar(false)

	return &storyView{
		style: lipgloss.NewStyle().Margin(1, 2),
		model: model,
	}
}

func newCommentView() *commentView {
	return &commentView{
		style: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1),
		model: viewport.New(0, 0),
	}
}

func newArticleView() *articleView {
	return &articleView{
		style: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1),
		model: viewport.New(0, 0),
	}
}

func newChatView(style lipgloss.Style) *chatView {
	model := viewport.New(0, 0)
	model.KeyMap = viewport.KeyMap{
		PageDown: key.NewBinding(),
		PageUp:   key.NewBinding(),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
		),
		Up:   key.NewBinding(),
		Down: key.NewBinding(),
	}

	prompt := textarea.New()
	prompt.Placeholder = "Send a message..."
	prompt.Prompt = "┃ "
	prompt.Cursor.Style = style
	prompt.CharLimit = 100
	prompt.ShowLineNumbers = false
	prompt.SetHeight(1)
	prompt.KeyMap.InsertNewline.SetEnabled(false)

	return &chatView{
		model:     model,
		messages:  []string{"Hi, I'm 🐈 Cha(t)mot. What can I help with?" + "\n\n"},
		prompt:    prompt,
		responses: []string{},
		input:     make(chan string, 10),
		output:    make(chan ollama.Response, 1),
		stop:      make(chan bool, 1),
	}
}

func (s *storyView) view() string {
	return s.style.Render(s.model.View())
}

func (s *storyView) selected() hackernews.Story {
	return s.model.SelectedItem().(hackernews.Story)
}

func (s *storyView) updateWindow(width int, height int) {
	x, y := s.style.GetFrameSize()
	s.model.SetSize(width-x, height-y)
}

func (c *commentView) headerView() string {
	return lipgloss.JoinHorizontal(lipgloss.Center, strings.Repeat(" ", max(0, c.model.Width)))
}

func (c *commentView) footerView() string {
	info := c.style.Render(fmt.Sprintf("%3.f%%", c.model.ScrollPercent()*100))
	line := strings.Repeat(" ", max(0, c.model.Width-lipgloss.Width(info)))

	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (c *commentView) view() string {
	return fmt.Sprintf("%s\n%s\n%s", c.headerView(), c.model.View(), c.footerView())
}

func (c *commentView) gotoTop() {
	c.model.SetYOffset(0)
}

func (c *commentView) gotoBottom() {
	c.model.SetYOffset(c.model.TotalLineCount())
}

func (c *commentView) setContent(comment string) {
	render, err := glamour.RenderWithEnvironmentConfig(comment)
	if err != nil {
		c.model.SetContent("")
	}

	c.model.SetContent(render)
}

func (c *commentView) updateWindow(width int, height int) {
	margin := lipgloss.Height(c.headerView()) + lipgloss.Height(c.footerView())
	c.model.Width = width
	c.model.Height = height - margin
}

func (a *articleView) headerView() string {
	line := strings.Repeat(" ", max(0, a.model.Width))

	return lipgloss.JoinHorizontal(lipgloss.Center, line)
}

func (a *articleView) footerView() string {
	info := a.style.Render(fmt.Sprintf("%3.f%%", a.model.ScrollPercent()*100))
	line := strings.Repeat(" ", max(0, a.model.Width-lipgloss.Width(info)))

	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (a *articleView) view() string {
	return fmt.Sprintf("%s\n%s\n%s", a.headerView(), a.model.View(), a.footerView())
}

func (a *articleView) gotoTop() {
	a.model.SetYOffset(0)
}

func (a *articleView) gotoBottom() {
	a.model.SetYOffset(a.model.TotalLineCount())
}

func (a *articleView) setContent(article string) {
	render, err := glamour.RenderWithEnvironmentConfig(article)
	if err != nil {
		a.model.SetContent("")
	}

	a.model.SetContent(render)
}

func (a *articleView) updateWindow(width int, height int) {
	margin := lipgloss.Height(a.headerView()) + lipgloss.Height(a.footerView())
	a.model.Width = width
	a.model.Height = height - margin
}

func (c *chatView) view() string {
	return fmt.Sprintf("%s%s%s", c.model.View(), "\n\n", c.prompt.View())
}

func (c *chatView) stopChat() {
	// Don't block the main loop
	select {
	case c.stop <- true:
	default:
	}
}
func (c *chatView) sendPrompt() {
	// Don't block the main loop
	select {
	case c.input <- c.prompt.Value():
	default:
	}
	c.prompt.Reset()
	c.model.GotoBottom()
}

func (c *chatView) sendArticle(article string) {
	article = "summarize this article in 10 lines" + article

	// Don't block the main loop
	select {
	case c.input <- article:
		const gap = "\n\n"
		c.messages = append(c.messages, gap+"*Summarizing the article in progress...⏳*"+gap)
		render, err := glamour.RenderWithEnvironmentConfig(strings.Join(c.messages, ""))

		if err != nil {
			c.model.SetContent("")
		}

		c.model.SetContent(render)
		c.model.GotoBottom()
	default:
	}
}

func (c *chatView) updateWindow(width int, height int) {
	c.prompt.SetWidth(width)
	c.model.Width = width
	c.model.Height = height - c.prompt.Height() - lipgloss.Height("\n\n")

	if len(c.messages) > 0 {
		render, err := glamour.RenderWithEnvironmentConfig(strings.Join(c.messages, ""))
		if err != nil {
			c.model.SetContent("")
		}

		c.model.SetContent(render)
	}

	c.model.GotoBottom()
}

func (c *chatView) formatResponse(response ollama.Response, foreground bool) {
	c.responses = append(c.responses, response.Response)

	if response.Done {
		const gap = "\n\n"
		c.responses = append(c.responses, gap+"---"+gap)
	}

	if foreground {
		c.messages = append(c.messages, c.responses...)
		render, err := glamour.RenderWithEnvironmentConfig(strings.Join(c.messages, ""))

		if err != nil {
			c.model.SetContent("")
		}

		c.model.SetContent(render)
		c.model.GotoBottom()
		c.responses = []string{}
	}
}

func (c *chatView) focus() tea.Cmd {
	cmd := c.prompt.Focus()

	c.messages = append(c.messages, c.responses...)
	md, err := glamour.RenderWithEnvironmentConfig(strings.Join(c.messages, ""))

	if err != nil {
		c.model.SetContent("")
	}

	c.model.SetContent(md)
	c.model.GotoBottom()
	c.responses = []string{}

	return cmd
}
