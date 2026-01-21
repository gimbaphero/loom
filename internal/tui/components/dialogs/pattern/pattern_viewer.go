// Copyright 2026 Teradata
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package pattern

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/teradata-labs/loom/internal/tui/components/dialogs"
	"github.com/teradata-labs/loom/internal/tui/styles"
	"github.com/teradata-labs/loom/internal/tui/util"
)

const (
	patternViewerDialogID dialogs.DialogID = "pattern-viewer"
)

// PatternViewerDialog shows pattern file content in a scrollable viewer
type PatternViewerDialog interface {
	dialogs.DialogModel
}

type patternViewerDialogCmp struct {
	wWidth, wHeight int
	width, height   int

	filePath string
	fileName string
	content  string

	viewport viewport.Model
	keys     PatternViewerKeyMap
	help     help.Model

	positionRow int
	positionCol int
}

// PatternViewerKeyMap defines key bindings for pattern viewer dialog
type PatternViewerKeyMap struct {
	Close key.Binding
}

// DefaultPatternViewerKeyMap returns default key bindings
func DefaultPatternViewerKeyMap() PatternViewerKeyMap {
	return PatternViewerKeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "close"),
		),
	}
}

// ShortHelp returns key bindings for the short help view
func (k PatternViewerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Close}
}

// FullHelp returns key bindings for the full help view
func (k PatternViewerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Close},
	}
}

func NewPatternViewerDialog(filePath string) PatternViewerDialog {
	t := styles.CurrentTheme()
	h := help.New()
	h.Styles = t.S().Help

	// Read file content
	// #nosec G304 -- filePath comes from user selecting a pattern file in the sidebar
	content, err := os.ReadFile(filePath)
	if err != nil {
		content = []byte("Error reading file: " + err.Error())
	}

	return &patternViewerDialogCmp{
		filePath: filePath,
		fileName: filepath.Base(filePath),
		content:  string(content),
		viewport: viewport.New(),
		keys:     DefaultPatternViewerKeyMap(),
		help:     h,
	}
}

func (m *patternViewerDialogCmp) ID() dialogs.DialogID {
	return patternViewerDialogID
}

func (m *patternViewerDialogCmp) Position() (int, int) {
	return m.positionRow, m.positionCol
}

func (m *patternViewerDialogCmp) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m *patternViewerDialogCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.wWidth, m.wHeight = msg.Width, msg.Height
		return m, m.resize()

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Close):
			return m, util.CmdHandler(dialogs.CloseDialogMsg{})
		}
	}

	// Forward all other messages to viewport for scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *patternViewerDialogCmp) resize() tea.Cmd {
	t := styles.CurrentTheme()

	// Dialog should be 70% of screen width, 80% of height
	m.width = int(float64(m.wWidth) * 0.7)
	m.height = int(float64(m.wHeight) * 0.8)

	// Ensure minimum size
	if m.width < 60 {
		m.width = 60
	}
	if m.height < 20 {
		m.height = 20
	}

	// Account for border and padding
	contentWidth := m.width - 4   // 2 for border, 2 for padding
	contentHeight := m.height - 6 // 2 for border, 2 for padding, 2 for title/help

	// Set viewport size
	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(contentHeight)

	// Build content
	content := m.buildContent(t)
	m.viewport.SetContent(content)

	// Set position (centered)
	m.positionRow = m.wHeight/2 - m.height/2
	m.positionCol = m.wWidth/2 - m.width/2

	return nil
}

func (m *patternViewerDialogCmp) buildContent(t *styles.Theme) string {
	var parts []string

	// File path
	pathLabel := t.S().Base.Foreground(t.FgMuted).Render("File:")
	pathValue := t.S().Base.Foreground(t.FgSubtle).Render(m.filePath)
	parts = append(parts, pathLabel+" "+pathValue)
	parts = append(parts, "")

	// Content with line numbers and syntax highlighting for YAML
	lines := strings.Split(m.content, "\n")
	maxLineNum := len(lines)
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))

	for i, line := range lines {
		numStr := fmt.Sprintf("%d", i+1)
		padding := strings.Repeat(" ", lineNumWidth-len(numStr))
		lineNum := t.S().Base.Foreground(t.FgMuted).Render(padding + numStr + " │ ")

		// Basic YAML syntax highlighting
		styledLine := m.highlightYAML(line, t)
		parts = append(parts, lineNum+styledLine)
	}

	return strings.Join(parts, "\n")
}

// highlightYAML applies basic syntax highlighting for YAML content
func (m *patternViewerDialogCmp) highlightYAML(line string, t *styles.Theme) string {
	trimmed := strings.TrimSpace(line)

	// Comment
	if strings.HasPrefix(trimmed, "#") {
		return t.S().Base.Foreground(t.FgMuted).Render(line)
	}

	// Key-value pair
	if idx := strings.Index(line, ":"); idx != -1 {
		key := line[:idx+1]
		value := line[idx+1:]

		styledKey := t.S().Base.Foreground(t.Primary).Bold(true).Render(key)
		styledValue := t.S().Base.Foreground(t.FgBase).Render(value)

		return styledKey + styledValue
	}

	// List item
	if strings.HasPrefix(trimmed, "-") {
		return t.S().Base.Foreground(t.FgBase).Render(line)
	}

	// Default
	return t.S().Base.Foreground(t.FgBase).Render(line)
}

func (m *patternViewerDialogCmp) View() string {
	t := styles.CurrentTheme()

	// Title
	title := t.S().Base.
		Bold(true).
		Foreground(t.Primary).
		Render("Pattern: " + m.fileName)

	// Content (viewport)
	content := m.viewport.View()

	// Help
	helpView := m.help.View(m.keys)

	// Assemble dialog
	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		content,
		"",
		helpView,
	)

	// Border
	style := t.S().Base.
		Width(m.width).
		Height(m.height).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus)

	return style.Render(inner)
}

func (m *patternViewerDialogCmp) Cursor() *tea.Cursor {
	return nil
}
