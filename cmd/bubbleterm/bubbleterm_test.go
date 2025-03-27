package bubbleterm

import (
	"chamot/cmd/hackernews"
	"chamot/cmd/ollama"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type mockHackerNews struct{}

func (m *mockHackerNews) Article(_ hackernews.Story) (string, error) {
	return "Happy Birthday to the fixed point combinator that changed the world", nil
}

func (m *mockHackerNews) Comment(_ hackernews.Story) (string, error) {
	return "World would be a very different place without YC", nil
}

func (m *mockHackerNews) Story() ([]hackernews.Story, error) {
	return []hackernews.Story{{
		Rank:       0,
		By:         "btilly",
		NumComment: 3,
		ID:         43332658,
		Kids:       []int{43339316, 43335480},
		Score:      1421,
		Time:       1741702475,
		TimeAgo:    "x days ago",
		PostTitle:  "Happy 20th birthday, Y Combinator",
		URL:        "/ycombinator",
		URLHost:    "ycombinator.com",
	}}, nil
}

type mockOllama struct{}

func (m *mockOllama) Chat(_ chan string, _ chan ollama.Response, _ chan bool) error {
	return nil
}

func TestUpdate(t *testing.T) {
	hn := &mockHackerNews{}
	ol := &mockOllama{}
	bt := NewBubbleTerm(hn, ol)

	bt.Init() // Nothing happens

	tests := []struct {
		name                 string
		input                tea.Msg
		expectedState        int
		expectedViewContains string
		expectedCmdIsNil     bool
	}{
		{
			name:                 "Initial state and view",
			input:                tea.WindowSizeMsg{Width: 80, Height: 80},
			expectedState:        0,
			expectedViewContains: "Happy 20th birthday, Y Combinator",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "Enter key moves to state 1",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter"), Alt: false, Paste: false},
			expectedState:        1,
			expectedViewContains: "World would be a very different place without YC",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "g key keeps to state 1",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g"), Alt: false, Paste: false},
			expectedState:        1,
			expectedViewContains: "World would be a very different place without YC",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "G key keeps to state 1",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G"), Alt: false, Paste: false},
			expectedState:        1,
			expectedViewContains: "World would be a very different place without YC",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "ctrl+x key moves to state 0",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+x"), Alt: false, Paste: false},
			expectedState:        0,
			expectedViewContains: "Happy 20th birthday, Y Combinator",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "Space key moves to state 2",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" "), Alt: false, Paste: false},
			expectedState:        2,
			expectedViewContains: "Happy Birthday to the fixed point combinator that changed the world",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "g key keeps to state 2",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g"), Alt: false, Paste: false},
			expectedState:        2,
			expectedViewContains: "Happy Birthday to the fixed point combinator that changed the world",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "G key keeps to state 2",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G"), Alt: false, Paste: false},
			expectedState:        2,
			expectedViewContains: "Happy Birthday to the fixed point combinator that changed the world",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "ctrl+x key moves to state 0",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+x"), Alt: false, Paste: false},
			expectedState:        0,
			expectedViewContains: "Happy 20th birthday, Y Combinator",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "o key moves to state 3",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o"), Alt: false, Paste: false},
			expectedState:        3,
			expectedViewContains: "Send a message",
			expectedCmdIsNil:     false,
		},
		{
			name:                 "Enter keeps to state 3",
			input:                tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{}, Alt: false, Paste: false},
			expectedState:        3,
			expectedViewContains: "Send a message",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "Type 'hello' keeps to state 3",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello"), Alt: false, Paste: false},
			expectedState:        3,
			expectedViewContains: "hello",
			expectedCmdIsNil:     false,
		},
		{
			name:                 "ctrl+] keeps to state 3 and keeps prompt",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+]"), Alt: false, Paste: false},
			expectedState:        3,
			expectedViewContains: "hello",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "ctrl+x key moves to state 0",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+x"), Alt: false, Paste: false},
			expectedState:        0,
			expectedViewContains: "Happy 20th birthday, Y Combinator",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "s key keeps to state 0",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s"), Alt: false, Paste: false},
			expectedState:        0,
			expectedViewContains: "Happy 20th birthday, Y Combinator",
			expectedCmdIsNil:     true,
		},
		{
			name:                 "ctrl+c key moves cmd to Quit",
			input:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+c"), Alt: false, Paste: false},
			expectedState:        0,
			expectedViewContains: "Happy 20th birthday, Y Combinator",
			expectedCmdIsNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, cmd := bt.Update(tt.input)
			bt = model.(*BubbleTerm)
			view := bt.View()

			if bt.state != tt.expectedState {
				t.Errorf("Expected state to be %d, got %d", tt.expectedState, bt.state)
			}

			if (cmd == nil) != tt.expectedCmdIsNil {
				t.Errorf("Expected cmd to be nil: %v, got %v", tt.expectedCmdIsNil, cmd)
			}

			if !strings.Contains(view, tt.expectedViewContains) {
				t.Errorf("Expected view to contain '%s', got %v", tt.expectedViewContains, view)
			}
		})
	}
}
