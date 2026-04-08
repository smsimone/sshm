package profile

import (
	"fmt"
	"os"
	"slices"
	sshconn "sshm/internal/ssh_conn"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/x/term"
)

const (
	labelField = iota
	usernameField
	passwordField
	keyPassphraseField
)

const (
	privateKeyField = 4
	submitField     = 5
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type model struct {
	focusIndex  int
	inputFields []textinput.Model
	pvtKeyField textarea.Model
	profiles    []string
	cursorMode  cursor.Mode
	quitting    bool
}

type pvtKeyRequest struct {
	key        string
	passphrase string
}

var _ tea.Model = (*model)(nil)

func (m *model) GetProfile() (sshconn.Profile, *pvtKeyRequest) {
	prof := sshconn.Profile{
		Label:    m.inputFields[labelField].Value(),
		Username: m.inputFields[usernameField].Value(),
	}
	psw := m.inputFields[passwordField].Value()
	if len(psw) > 0 {
		prof.Password = new(psw)
	}

	pvtKey := m.pvtKeyField.Value()
	pvtPassphrase := m.inputFields[keyPassphraseField].Value()

	if len(pvtKey) > 0 {
		req := pvtKeyRequest{
			key:        pvtKey,
			passphrase: pvtPassphrase,
		}
		return prof, &req
	}

	return prof, nil
}

func initialModel() *model {

	width, _, err := term.GetSize(os.Stdin.Fd())
	if err != nil {
		width = 90
	}

	m := model{inputFields: make([]textinput.Model, 4)}

	p, _ := sshconn.LoadProfiles()
	profiles := make([]string, len(*p))
	for i := range *p {
		profiles[i] = (*p)[i].Label
	}
	m.profiles = profiles

	var t textinput.Model
	for i := range m.inputFields {
		t = textinput.New()
		t.SetWidth(width)
		t.CharLimit = 32

		s := t.Styles()
		s.Cursor.Color = lipgloss.Color("205")
		s.Focused.Prompt = focusedStyle
		s.Focused.Text = focusedStyle
		s.Blurred.Prompt = blurredStyle
		s.Focused.Text = focusedStyle
		t.SetStyles(s)

		switch i {
		case 0:
			t.Placeholder = "Label"
			t.CharLimit = 64
			t.Validate = func(s string) error {
				if idx := slices.IndexFunc(m.profiles, func(p string) bool {
					return p == s
				}); idx == -1 {
					return nil
				}
				return fmt.Errorf("Label already used")
			}
			t.Focus()
		case 1:
			t.Placeholder = "Username"
			t.CharLimit = 64
		case 2:
			t.Placeholder = "Password"
			t.EchoCharacter = '*'
			t.EchoMode = textinput.EchoPassword
		case 3:
			t.Placeholder = "Passphrase"
			t.EchoCharacter = '*'
			t.EchoMode = textinput.EchoPassword
		}
		m.inputFields[i] = t
	}

	pvtKeyArea := textarea.New()
	pvtKeyArea.Placeholder = "Private key content or file path"
	pvtKeyArea.SetWidth(width)
	pvtKeyArea.SetHeight(5)

	m.pvtKeyField = pvtKeyArea

	return &m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, textarea.Blink)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.focusIndex == privateKeyField {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c", "esc":
				m.quitting = true
				return m, tea.Quit
			case "tab", "shift+tab":
				if keyMsg.String() == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > submitField {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = submitField
				}

				cmds := m.setFocus(m.focusIndex)
				return m, tea.Batch(cmds...)
			}
		}

		var cmd tea.Cmd
		m.pvtKeyField, cmd = m.pvtKeyField.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == submitField {
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > submitField {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = submitField
			}

			cmds := m.setFocus(m.focusIndex)
			return m, tea.Batch(cmds...)
		}
	}

	return m, m.updateInputs(msg)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	if m.focusIndex < 0 || m.focusIndex >= len(m.inputFields) {
		return nil
	}

	var cmd tea.Cmd
	m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
	return cmd
}

func (m *model) setFocus(focusIndex int) []tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputFields)+1)

	for i := range m.inputFields {
		if i == focusIndex {
			cmds[i] = m.inputFields[i].Focus()
			continue
		}
		m.inputFields[i].Blur()
	}

	if focusIndex == privateKeyField {
		cmds[len(m.inputFields)] = m.pvtKeyField.Focus()
	} else {
		m.pvtKeyField.Blur()
	}

	return cmds
}

func (m *model) View() tea.View {
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

	b.WriteRune('\n')
	b.WriteString(m.pvtKeyField.View())
	if m.cursorMode != cursor.CursorHide && m.pvtKeyField.Focused() {
		c = m.pvtKeyField.Cursor()
		if c != nil {
			c.Y += len(m.inputFields)
		}
	}

	button := &blurredButton
	if m.focusIndex == submitField {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.quitting {
		b.WriteRune('\n')
	}

	v := tea.NewView(b.String())
	v.Cursor = c
	return v
}
