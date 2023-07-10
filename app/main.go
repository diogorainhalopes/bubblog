package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	p        *tea.Program
	fp       filepicker.Model
	l        list.Model
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

const (
	menuView sessionState = iota
	fileSelectView
	logView
)

type sessionState int

type item struct {
	title, desc string
}

type clearErrorMsg struct{}

type filePickerModel struct {
	filepicker   *filepicker.Model
	selectedFile string
}

type menuModel struct {
	list   *list.Model
	option int
}

type model struct {
	state sessionState
	// menu Model
	menu menuModel
	// filepicker Model
	fp filePickerModel
	// other
	quitting bool
	err      error
}

func (m model) Init() tea.Cmd {
	return m.fp.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case clearErrorMsg:
		m.err = nil
	}

	switch m.state {
	case fileSelectView:
		return updateFilePicker(msg, m)
	}
	return updateMenu(msg, m)
}

// Update Sub-Functions

func updateMenu(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.menu.list.SelectedItem().(item)
			if ok && i.desc == "2" {
				m.state = fileSelectView
				m.menu.option, m.err = strconv.Atoi(i.desc)
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.menu.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd

	*m.menu.list, cmd = m.menu.list.Update(msg)
	return m, cmd
}

func updateFilePicker(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			m.state = menuView
			return m, nil
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	*m.fp.filepicker, cmd = m.fp.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.fp.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.fp.selectedFile = path
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.fp.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.fp.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

// Main View

func (m model) View() string {
	if m.quitting {
		return ""
	}
	switch m.state {
	case menuView:
		return m.viewMenu()
	case fileSelectView:
		return m.viewFilePicker()
	default:
		return "oopsie"
	}

}

// Sub-Views

func (m model) viewFilePicker() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.fp.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.fp.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.fp.filepicker.Styles.Selected.Render(m.fp.selectedFile))
	}
	s.WriteString("\n\n" + m.fp.filepicker.View() + "\n")
	return s.String()
}

func (m model) viewMenu() string {
	return docStyle.Render(m.menu.list.View())
}

// Helper Functions

// item implements ListableItem
func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func main() {

	items := []list.Item{
		item{title: "Open previously opened log file", desc: "1"},
		item{title: "Select log file", desc: "2"},
	}

	fp = filepicker.New()
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}
	//fp.CurrentDirectory, _ = os.UserHomeDir()

	l = list.New(items, list.NewDefaultDelegate(), 0, 0)

	m := model{
		menu: menuModel{
			list:   &l,
			option: 0,
		},
		fp: filePickerModel{
			filepicker:   &fp,
			selectedFile: "",
		},
		state: menuView,
	}
	m.menu.list.Title = "golog"

	p = tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
