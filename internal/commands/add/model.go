package add

import (
	"fmt"
	"strconv"
	"strings"

	sshconn "github.com/smsimone/sshm/internal/ssh_conn"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/cursor"
)

const (
	labelField = iota
	hostField
	portField
	profileField
	tagsField
)

var (
	focusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type Model struct {
	focusIndex  int
	inputFields []textinput.Model
	profiles    []string
	cursorMode  cursor.Mode
	submitting  bool
}

var _ tea.Model = (*Model)(nil)

func InitialModel(conn *sshconn.Connection) Model {
	m := Model{inputFields: make([]textinput.Model, 5)}

	p, _ := sshconn.LoadProfiles()
	profiles := make([]string, len(*p))
	for i := range *p {
		profiles[i] = (*p)[i].Label
	}
	m.profiles = profiles

	var t textinput.Model
	for i := range m.inputFields {
		t = textinput.New()
		t.SetWidth(32)
		t.CharLimit = 32

		s := t.Styles()
		s.Cursor.Color = lipgloss.Color("205")
		s.Focused.Prompt = focusedStyle
		s.Focused.Text = focusedStyle
		s.Blurred.Prompt = blurredStyle
		s.Focused.Text = focusedStyle
		t.SetStyles(s)

		switch i {
		case labelField:
			t.Placeholder = "Label"
			t.CharLimit = 64
			t.Focus()
		case hostField:
			t.Placeholder = "Host"
			t.CharLimit = 64
		case portField:
			t.Placeholder = "Port"
			t.Validate = func(s string) error {
				if _, err := strconv.Atoi(s); err != nil {
					return fmt.Errorf("invalid number '%s'", s)
				}
				return nil
			}
		case profileField:
			t.Placeholder = "Profile"
			t.ShowSuggestions = true
			t.SetSuggestions(profiles)
		case tagsField:
			t.Placeholder = "Tags (comma separated)"
		}

		m.inputFields[i] = t
	}

	m.initializeFields(conn)

	return m
}

func (m *Model) initializeFields(conn *sshconn.Connection) {
	if conn == nil {
		return
	}

	m.inputFields[labelField].SetValue(conn.Label)
	m.inputFields[hostField].SetValue(conn.Host)
	m.inputFields[portField].SetValue(strconv.Itoa(conn.Port))
	if conn.Profile != nil {
		m.inputFields[profileField].SetValue(*conn.Profile)
	}
	if conn.Tags != nil {
		m.inputFields[tagsField].SetValue(strings.Join(*conn.Tags, ","))
	}
}

func (m *Model) GetConnection() sshconn.Connection {
	port, _ := strconv.Atoi(m.inputFields[portField].Value())
	var profile *string
	if len(m.inputFields[profileField].Value()) > 0 {
		profile = new(m.inputFields[profileField].Value())
	}
	var tags *[]string
	if len(m.inputFields[tagsField].Value()) > 0 {
		data := strings.Split(m.inputFields[tagsField].Value(), ",")
		tags = &data
	}

	return sshconn.Connection{
		Label:   m.inputFields[labelField].Value(),
		Host:    m.inputFields[hostField].Value(),
		Port:    port,
		Profile: profile,
		Tags:    tags,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":

			return m, tea.Quit

		case "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputFields) {
				m.submitting = true
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputFields) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputFields)
			}

			cmds := make([]tea.Cmd, len(m.inputFields))
			for i := 0; i <= len(m.inputFields)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputFields[i].Focus()
					continue
				}
				m.inputFields[i].Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputFields))

	for i := range m.inputFields {
		m.inputFields[i], cmds[i] = m.inputFields[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	var b strings.Builder
	var c *tea.Cursor

	for i, in := range m.inputFields {
		b.WriteString(m.inputFields[i].View())
		if i < len(m.inputFields)-1 {
			b.WriteRune('\n')
		}
		if m.cursorMode != cursor.CursorHide && in.Focused() {
			c = in.Cursor()
			if c != nil {
				c.Y += i
			}
		}
	}
	fmt.Fprintf(&b, "%s\n", strings.Join(m.profiles, ","))

	button := &blurredButton
	if m.focusIndex == len(m.inputFields) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.submitting {
		b.WriteRune('\n')
	}

	v := tea.NewView(b.String())
	v.Cursor = c
	return v
}
