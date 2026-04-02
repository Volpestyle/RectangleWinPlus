package main

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

const (
	hudTimerID = 1
	hudTimeout = 800 // milliseconds
	hudWidth   = 200
	hudHeight  = 56
	hudAlpha   = 210 // ~82% opaque
)

var (
	hudWindow  w32.HWND
	hudText    string
	hudFont    w32.HFONT
	hudBgBrush w32.HBRUSH
)

func initHUD() {
	className, _ := syscall.UTF16PtrFromString("RectangleWinHUD")
	wndProc := syscall.NewCallback(hudWndProc)
	inst := w32.GetModuleHandle("")

	wc := w32.WNDCLASSEX{
		Style:      w32.CS_HREDRAW | w32.CS_VREDRAW,
		WndProc:    wndProc,
		Instance:   inst,
		ClassName:  className,
		Background: 0,
	}
	w32.RegisterClassEx(&wc)

	windowName, _ := syscall.UTF16PtrFromString("")
	hudWindow = w32.CreateWindowEx(
		w32.WS_EX_LAYERED|w32.WS_EX_TOPMOST|w32.WS_EX_TOOLWINDOW|w32.WS_EX_NOACTIVATE|w32.WS_EX_TRANSPARENT,
		className, windowName,
		w32.WS_POPUP,
		0, 0, hudWidth, hudHeight,
		0, 0, inst, nil,
	)

	w32.SetLayeredWindowAttributes(hudWindow, 0, hudAlpha, w32.LWA_ALPHA)

	lf := w32.LOGFONT{
		Height:  -22,
		Weight:  600, // semi-bold
		Quality: 5,   // CLEARTYPE_QUALITY
	}
	lf.SetFaceName("Segoe UI")
	hudFont = w32.CreateFontIndirect(&lf)

	hudBgBrush = w32.CreateSolidBrush(0x261B1A) // BGR: #1A1B26
}

func showHUD(monitor w32.HMONITOR, text string) {
	if hudWindow == 0 {
		return
	}
	hudText = text

	var monInfo w32.MONITORINFO
	if !w32.GetMonitorInfo(monitor, &monInfo) {
		return
	}
	work := monInfo.RcWork
	x := int(work.Left) + (int(work.Width())-hudWidth)/2
	y := int(work.Top) + int(float64(work.Height())*0.15) // upper area, not dead center

	w32.SetWindowPos(hudWindow, w32.HWND_TOPMOST,
		x, y, hudWidth, hudHeight,
		w32.SWP_NOACTIVATE|w32.SWP_SHOWWINDOW)
	w32.InvalidateRect(hudWindow, nil, true)
	w32.UpdateWindow(hudWindow)

	w32.SetTimer(hudWindow, hudTimerID, hudTimeout, 0)
}

func hudWndProc(hwnd w32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case w32.WM_PAINT:
		var ps w32.PAINTSTRUCT
		hdc := w32.BeginPaint(hwnd, &ps)

		rect := &w32.RECT{Left: 0, Top: 0, Right: hudWidth, Bottom: hudHeight}
		w32.FillRect(hdc, rect, hudBgBrush)

		oldFont := w32.SelectObject(hdc, w32.HGDIOBJ(hudFont))
		w32.SetBkMode(hdc, w32.TRANSPARENT)
		w32.SetTextColor(hdc, 0xF5CAC0) // BGR: #C0CAF5
		w32.DrawText(hdc, hudText, rect, w32.DT_CENTER|w32.DT_VCENTER|w32.DT_SINGLELINE|w32.DT_NOPREFIX)
		w32.SelectObject(hdc, oldFont)

		w32.EndPaint(hwnd, &ps)
		return 0

	case w32.WM_TIMER:
		if uintptr(wParam) == hudTimerID {
			w32.ShowWindow(hudWindow, w32.SW_HIDE)
			killTimer(hudWindow, hudTimerID)
		}
		return 0

	case w32.WM_ERASEBKGND:
		return 1 // prevent flicker
	}

	return w32.DefWindowProc(hwnd, msg, wParam, lParam)
}

var procKillTimer = syscall.NewLazyDLL("user32.dll").NewProc("KillTimer")

func killTimer(hwnd w32.HWND, id uintptr) {
	procKillTimer.Call(uintptr(hwnd), id)
}

// hudWndProcPtr is needed to prevent garbage collection of the callback.
var _ = unsafe.Pointer(nil)
