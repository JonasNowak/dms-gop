package tui

import (
	"testing"

	"github.com/JonasNowak/dms-gop/gops"
	"github.com/JonasNowak/dms-gop/models"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTUIModel(procs []*models.ProcessInfo) *ResponsiveTUIModel {
	m := &ResponsiveTUIModel{
		keybinds:      defaultResolvedKeybinds(),
		sortBy:        gops.SortByCPU,
		selectedPID:   -1,
		mergeChildren: true,
		processTable: table.New(
			table.WithColumns([]table.Column{
				{Title: "PID", Width: 5},
				{Title: "USER", Width: 6},
				{Title: "CPU%", Width: 5},
				{Title: "MEM%", Width: 13},
				{Title: "COMMAND", Width: 20},
			}),
			table.WithHeight(10),
			table.WithFocused(true),
		),
	}
	m.processTable.SetStyles(m.tableStyles())
	if procs != nil {
		m.metrics = &models.SystemMetrics{Processes: procs}
		m.sortProcessesLocally()
		m.updateProcessTable()
	}
	return m
}

func sampleProcesses() []*models.ProcessInfo {
	return []*models.ProcessInfo{
		{PID: 30, Command: "zeta", CPU: 10.0, MemoryKB: 400},
		{PID: 10, Command: "Alpha", CPU: 50.0, MemoryKB: 100},
		{PID: 20, Command: "beta", CPU: 10.0, MemoryKB: 300},
		{PID: 40, Command: "gamma", CPU: 25.0, MemoryKB: 200},
	}
}

func pressRune(m *ResponsiveTUIModel, r rune) tea.Cmd {
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return cmd
}

func orderedPIDs(t *testing.T, m *ResponsiveTUIModel) []int32 {
	t.Helper()
	require.NotNil(t, m.metrics)
	pids := make([]int32, len(m.metrics.Processes))
	for i, p := range m.metrics.Processes {
		pids[i] = p.PID
	}
	return pids
}

func TestSortToggleDirectionByKey(t *testing.T) {
	m := newTestTUIModel(sampleProcesses())

	// Startup default: CPU descending, ties broken by PID ascending.
	assert.Equal(t, gops.SortByCPU, m.sortBy)
	assert.False(t, m.sortReversed)
	assert.Equal(t, " ↓CPU", m.sortIndicator())
	assert.Equal(t, []int32{10, 40, 20, 30}, orderedPIDs(t, m))

	// Pressing c while already sorted by CPU reverses to CPU ascending.
	cmd := pressRune(m, 'c')
	assert.Nil(t, cmd)
	assert.True(t, m.sortReversed)
	assert.Equal(t, " ↑CPU", m.sortIndicator())
	assert.Equal(t, []int32{20, 30, 40, 10}, orderedPIDs(t, m))

	// Pressing c again restores CPU descending.
	cmd = pressRune(m, 'c')
	assert.Nil(t, cmd)
	assert.False(t, m.sortReversed)
	assert.Equal(t, " ↓CPU", m.sortIndicator())
	assert.Equal(t, []int32{10, 40, 20, 30}, orderedPIDs(t, m))

	// Reversing CPU, then switching to memory resets direction to memory's default (descending).
	pressRune(m, 'c')
	require.True(t, m.sortReversed)

	cmd = pressRune(m, 'm')
	assert.NotNil(t, cmd)
	assert.Equal(t, gops.SortByMemory, m.sortBy)
	assert.False(t, m.sortReversed)
	assert.Equal(t, " ↓MEM", m.sortIndicator())
	assert.Equal(t, []int32{30, 20, 40, 10}, orderedPIDs(t, m))

	// Pressing m again reverses memory to ascending.
	cmd = pressRune(m, 'm')
	assert.Nil(t, cmd)
	assert.True(t, m.sortReversed)
	assert.Equal(t, " ↑MEM", m.sortIndicator())
	assert.Equal(t, []int32{10, 40, 20, 30}, orderedPIDs(t, m))

	// Switching to name defaults to ascending (case-insensitive A->Z), indicator ↑NAME.
	cmd = pressRune(m, 'n')
	assert.NotNil(t, cmd)
	assert.Equal(t, gops.SortByName, m.sortBy)
	assert.False(t, m.sortReversed)
	assert.Equal(t, " ↑NAME", m.sortIndicator())
	assert.Equal(t, []int32{10, 20, 40, 30}, orderedPIDs(t, m))

	// Pressing n again reverses name to descending (Z->A), indicator ↓NAME.
	cmd = pressRune(m, 'n')
	assert.Nil(t, cmd)
	assert.True(t, m.sortReversed)
	assert.Equal(t, " ↓NAME", m.sortIndicator())
	assert.Equal(t, []int32{30, 40, 20, 10}, orderedPIDs(t, m))

	// Switching to PID defaults to ascending, indicator ↑PID.
	cmd = pressRune(m, 'p')
	assert.NotNil(t, cmd)
	assert.Equal(t, gops.SortByPID, m.sortBy)
	assert.False(t, m.sortReversed)
	assert.Equal(t, " ↑PID", m.sortIndicator())
	assert.Equal(t, []int32{10, 20, 30, 40}, orderedPIDs(t, m))

	// Pressing p again reverses PID to descending, indicator ↓PID.
	cmd = pressRune(m, 'p')
	assert.Nil(t, cmd)
	assert.True(t, m.sortReversed)
	assert.Equal(t, " ↓PID", m.sortIndicator())
	assert.Equal(t, []int32{40, 30, 20, 10}, orderedPIDs(t, m))
}

func TestSortOrderSurvivesFetchUpdates(t *testing.T) {
	// Toggle before first metrics load arrives.
	m := newTestTUIModel(nil)
	pressRune(m, 'c')
	assert.True(t, m.sortReversed)

	// Incoming fetch with server-default descending order gets re-sorted to reversed order.
	m.Update(fetchDataMsg{
		generation: m.fetchGeneration,
		metrics: &models.SystemMetrics{
			Processes: []*models.ProcessInfo{
				{PID: 10, Command: "Alpha", CPU: 50.0, MemoryKB: 100},
				{PID: 40, Command: "gamma", CPU: 25.0, MemoryKB: 200},
				{PID: 20, Command: "beta", CPU: 10.0, MemoryKB: 300},
			},
		},
	})
	assert.Equal(t, []int32{20, 40, 10}, orderedPIDs(t, m))

	// Stale generation updates are still ignored.
	m.Update(fetchDataMsg{
		generation: m.fetchGeneration - 1,
		metrics: &models.SystemMetrics{
			Processes: []*models.ProcessInfo{{PID: 99, Command: "stale"}},
		},
	})
	assert.Equal(t, []int32{20, 40, 10}, orderedPIDs(t, m))
}

func TestRenderProcessPanelSortIndicator(t *testing.T) {
	m := newTestTUIModel(sampleProcesses())

	tests := []struct {
		key      rune
		expected string
	}{
		{0, "↓CPU"},
		{'c', "↑CPU"},
		{'m', "↓MEM"},
		{'m', "↑MEM"},
		{'n', "↑NAME"},
		{'n', "↓NAME"},
		{'p', "↑PID"},
		{'p', "↓PID"},
	}

	for _, tt := range tests {
		if tt.key != 0 {
			pressRune(m, tt.key)
		}
		rendered := m.renderProcessPanel(90, 20)
		assert.Contains(t, rendered, tt.expected)
	}
}

func TestSortKeysIgnoredInModalModes(t *testing.T) {
	m := newTestTUIModel(sampleProcesses())

	// Search input consumes 'c' instead of reversing sort.
	pressRune(m, '/')
	require.True(t, m.searchActive)
	pressRune(m, 'c')
	assert.False(t, m.sortReversed)
	assert.Equal(t, "c", m.searchInput)

	// Kill confirmation ignores 'c'.
	m.searchActive = false
	m.searchInput = ""
	m.killConfirmPID = 10
	pressRune(m, 'c')
	assert.False(t, m.sortReversed)
}
