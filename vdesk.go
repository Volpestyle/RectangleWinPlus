package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ahmetb/RectangleWin/w32ex"
	"github.com/gonutz/w32/v2"
)

type desktopKey struct {
	Monitor w32.HMONITOR
	Desktop int // desktop ID
}

type windowState struct {
	HWND          w32.HWND
	UsedCloak     bool
	HadToolWindow bool // true if window already had WS_EX_TOOLWINDOW
}

type monitorDesktops struct {
	desktops    []int // ordered list of desktop IDs
	activeIndex int
}

type vdeskManager struct {
	monitors  map[w32.HMONITOR]*monitorDesktops
	hidden    map[desktopKey][]windowState
	stateFile string
	nextID    int // auto-incrementing desktop ID generator
}

func newVdeskManager() *vdeskManager {
	stateDir := os.Getenv("APPDATA")
	if stateDir == "" {
		stateDir = "."
	}
	dir := filepath.Join(stateDir, "RectangleWin")
	os.MkdirAll(dir, 0755)

	return &vdeskManager{
		monitors:  make(map[w32.HMONITOR]*monitorDesktops),
		hidden:    make(map[desktopKey][]windowState),
		stateFile: filepath.Join(dir, "vdesk-state.json"),
		nextID:    1, // 0 is the default desktop
	}
}

func (m *vdeskManager) getOrInitMonitor(monitor w32.HMONITOR) *monitorDesktops {
	md, ok := m.monitors[monitor]
	if !ok {
		md = &monitorDesktops{desktops: []int{0}, activeIndex: 0}
		m.monitors[monitor] = md
	}
	return md
}

func (m *vdeskManager) switchDesktop(delta int) {
	hwnd := w32.GetForegroundWindow()
	if hwnd == 0 {
		return
	}
	monitor := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	md := m.getOrInitMonitor(monitor)

	if len(md.desktops) <= 1 {
		return
	}

	target := ((md.activeIndex + delta) % len(md.desktops) + len(md.desktops)) % len(md.desktops)
	if target == md.activeIndex {
		return
	}

	m.doSwitch(monitor, md, target)
}

func (m *vdeskManager) doSwitch(monitor w32.HMONITOR, md *monitorDesktops, targetIndex int) {
	currentID := md.desktops[md.activeIndex]
	targetID := md.desktops[targetIndex]

	fmt.Printf("vdesk: monitor 0x%X: desktop %d -> %d (%d/%d)\n",
		monitor, currentID, targetID, targetIndex+1, len(md.desktops))

	// Phase 1: hide all visible windows on this monitor's current desktop
	visible := enumerateWindowsOnMonitor(monitor)
	var states []windowState
	for _, w := range visible {
		states = append(states, hideFromDesktop(w))
	}
	if len(states) > 0 {
		key := desktopKey{Monitor: monitor, Desktop: currentID}
		m.hidden[key] = append(m.hidden[key], states...)
	}

	// Phase 2: restore windows on the target desktop
	targetKey := desktopKey{Monitor: monitor, Desktop: targetID}
	var focusTarget w32.HWND
	if toRestore, ok := m.hidden[targetKey]; ok {
		for _, ws := range toRestore {
			if !w32.IsWindow(ws.HWND) {
				continue
			}
			restoreToDesktop(ws)
			focusTarget = ws.HWND
		}
		delete(m.hidden, targetKey)
	}

	if focusTarget != 0 {
		w32.SetForegroundWindow(focusTarget)
	}

	md.activeIndex = targetIndex
	m.saveState()
	showHUD(monitor, fmt.Sprintf("Desktop %d / %d", targetIndex+1, len(md.desktops)))
	fmt.Printf("vdesk: hid %d, restored %d\n", len(visible), targetID)
}

func (m *vdeskManager) createDesktop() {
	hwnd := w32.GetForegroundWindow()
	if hwnd == 0 {
		return
	}
	monitor := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	md := m.getOrInitMonitor(monitor)

	newID := m.nextID
	m.nextID++

	// Insert after current position
	insertIdx := md.activeIndex + 1
	md.desktops = append(md.desktops, 0)
	copy(md.desktops[insertIdx+1:], md.desktops[insertIdx:])
	md.desktops[insertIdx] = newID

	fmt.Printf("vdesk: created desktop %d on monitor 0x%X (total: %d)\n",
		newID, monitor, len(md.desktops))

	// doSwitch will show the "Desktop X / Y" HUD after switching
	m.doSwitch(monitor, md, insertIdx)
}

func (m *vdeskManager) destroyDesktop() {
	hwnd := w32.GetForegroundWindow()
	if hwnd == 0 {
		return
	}
	monitor := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	md := m.getOrInitMonitor(monitor)

	if len(md.desktops) <= 1 {
		fmt.Println("vdesk: can't close the last desktop")
		return
	}

	closingID := md.desktops[md.activeIndex]

	// Pick neighbor to switch to (prefer previous, fall back to next)
	targetIndex := md.activeIndex - 1
	if targetIndex < 0 {
		targetIndex = 1
	}
	targetID := md.desktops[targetIndex]

	fmt.Printf("vdesk: closing desktop %d on monitor 0x%X, moving to %d\n",
		closingID, monitor, targetID)

	// Restore any hidden windows that belonged to the target desktop
	targetKey := desktopKey{Monitor: monitor, Desktop: targetID}
	if toRestore, ok := m.hidden[targetKey]; ok {
		for _, ws := range toRestore {
			if w32.IsWindow(ws.HWND) {
				restoreToDesktop(ws)
			}
		}
		delete(m.hidden, targetKey)
	}

	// Discard any hidden windows from the closing desktop (show them —
	// they merge into the target desktop along with currently visible windows)
	closingKey := desktopKey{Monitor: monitor, Desktop: closingID}
	if leftover, ok := m.hidden[closingKey]; ok {
		for _, ws := range leftover {
			if w32.IsWindow(ws.HWND) {
				restoreToDesktop(ws)
			}
		}
		delete(m.hidden, closingKey)
	}

	// Remove the desktop from the list
	md.desktops = append(md.desktops[:md.activeIndex], md.desktops[md.activeIndex+1:]...)
	if targetIndex > md.activeIndex {
		targetIndex-- // adjust for removal
	}
	md.activeIndex = targetIndex

	m.saveState()
	showHUD(monitor, fmt.Sprintf("Closed  (%d / %d)", md.activeIndex+1, len(md.desktops)))
	fmt.Printf("vdesk: closed, %d desktops remain on monitor 0x%X\n",
		len(md.desktops), monitor)
}

func (m *vdeskManager) showAllWindows() {
	count := 0
	for key, states := range m.hidden {
		for _, ws := range states {
			if w32.IsWindow(ws.HWND) {
				restoreToDesktop(ws)
				count++
			}
		}
		delete(m.hidden, key)
	}
	m.monitors = make(map[w32.HMONITOR]*monitorDesktops)
	os.Remove(m.stateFile)
	fmt.Printf("vdesk: showed all windows (%d restored)\n", count)
}

// saveState writes hidden window state to disk for crash recovery.
func (m *vdeskManager) saveState() {
	if len(m.hidden) == 0 {
		os.Remove(m.stateFile)
		return
	}

	type stateEntry struct {
		HWND          uintptr `json:"hwnd"`
		UsedCloak     bool    `json:"usedCloak"`
		HadToolWindow bool    `json:"hadToolWindow"`
	}
	state := make(map[string][]stateEntry)
	for key, windows := range m.hidden {
		k := fmt.Sprintf("%d:%d", key.Monitor, key.Desktop)
		for _, ws := range windows {
			state[k] = append(state[k], stateEntry{
				HWND:          uintptr(ws.HWND),
				UsedCloak:     ws.UsedCloak,
				HadToolWindow: ws.HadToolWindow,
			})
		}
	}

	data, err := json.Marshal(state)
	if err != nil {
		fmt.Printf("vdesk: warn: marshal: %v\n", err)
		return
	}
	if err := os.WriteFile(m.stateFile, data, 0644); err != nil {
		fmt.Printf("vdesk: warn: write state: %v\n", err)
	}
}

// loadAndRecover uncloaks windows left hidden by a previous crash.
func (m *vdeskManager) loadAndRecover() {
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		return
	}

	type stateEntry struct {
		HWND          uintptr `json:"hwnd"`
		UsedCloak     bool    `json:"usedCloak"`
		HadToolWindow bool    `json:"hadToolWindow"`
	}
	var state map[string][]stateEntry
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Printf("vdesk: warn: parse state: %v\n", err)
		os.Remove(m.stateFile)
		return
	}

	count := 0
	for _, entries := range state {
		for _, e := range entries {
			hwnd := w32.HWND(e.HWND)
			if !w32.IsWindow(hwnd) {
				continue
			}
			restoreToDesktop(windowState{
				HWND:          hwnd,
				UsedCloak:     e.UsedCloak,
				HadToolWindow: e.HadToolWindow,
			})
			count++
		}
	}

	os.Remove(m.stateFile)
	if count > 0 {
		fmt.Printf("vdesk: recovered %d windows from previous crash\n", count)
	}
}

func enumerateWindowsOnMonitor(monitor w32.HMONITOR) []w32.HWND {
	var result []w32.HWND
	w32.EnumWindows(func(hwnd w32.HWND) bool {
		if !isZonableWindow(hwnd) {
			return true
		}
		if w32ex.IsWindowCloaked(hwnd) {
			return true
		}
		if w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST) == monitor {
			result = append(result, hwnd)
		}
		return true
	})
	return result
}

// hideFromDesktop cloaks a window and hides it from the taskbar/alt-tab.
func hideFromDesktop(hwnd w32.HWND) windowState {
	exStyle := w32.GetWindowLong(hwnd, GWL_EXSTYLE)
	hadToolWindow := exStyle&w32.WS_EX_TOOLWINDOW != 0

	// Cloak first so the style change is invisible
	usedCloak := w32ex.CloakWindow(hwnd, true)
	if !usedCloak {
		w32.ShowWindow(hwnd, w32.SW_HIDE)
	}

	// Add WS_EX_TOOLWINDOW to hide from taskbar and alt-tab
	if !hadToolWindow {
		w32.SetWindowLong(hwnd, GWL_EXSTYLE, exStyle|int32(w32.WS_EX_TOOLWINDOW))
	}

	return windowState{HWND: hwnd, UsedCloak: usedCloak, HadToolWindow: hadToolWindow}
}

// restoreToDesktop uncloaks a window and restores its taskbar presence.
func restoreToDesktop(ws windowState) {
	// Remove WS_EX_TOOLWINDOW before uncloaking so taskbar button reappears
	if !ws.HadToolWindow {
		exStyle := w32.GetWindowLong(ws.HWND, GWL_EXSTYLE)
		w32.SetWindowLong(ws.HWND, GWL_EXSTYLE, exStyle&^int32(w32.WS_EX_TOOLWINDOW))
	}

	if ws.UsedCloak {
		w32ex.CloakWindow(ws.HWND, false)
	} else {
		w32.ShowWindow(ws.HWND, w32.SW_SHOW)
	}
}
