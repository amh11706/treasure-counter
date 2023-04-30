package window

import (
	"image/color"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/TheTitanrain/w32"
)

var (
	wingdi       = syscall.NewLazyDLL("gdi32.dll")
	procGetPixel = wingdi.NewProc("GetPixel")

	modUser32           = syscall.NewLazyDLL("User32.dll")
	getForegroundWindow = modUser32.NewProc("GetForegroundWindow")
)

const ScansPerSecond = 30

type Window struct {
	Rect          w32.RECT
	Width, Height int
	Hwnd          w32.HWND
	Title         string
	Name          string
	Lock          chan struct{}
}

var ActiveWindow *Window

func (w *Window) RGBAAt(hdc w32.HDC, x, y int) color.RGBA {
	if w == nil || w.Hwnd == 0 {
		return color.RGBA{}
	}

	r1, _, _ := procGetPixel.Call(uintptr(hdc), uintptr(x), uintptr(y))
	return color.RGBA{
		R: uint8(r1 >> 16),
		G: uint8(r1 >> 8),
		B: uint8(r1),
		A: 255,
	}
}

func HexColorToRGB(c uintptr) color.RGBA {
	return color.RGBA{
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
		A: 255,
	}
}

// allow +/- error on each color channel
func FuzzyMatch(c1, c2 color.RGBA, fuzz uint8) bool {
	if c1.R > c2.R+fuzz || c1.R < c2.R-fuzz {
		return false
	}
	if c1.G > c2.G+fuzz || c1.G < c2.G-fuzz {
		return false
	}
	if c1.B > c2.B+fuzz || c1.B < c2.B-fuzz {
		return false
	}
	return true
}

func getWindowRect(hwnd w32.HWND) (rect w32.RECT) {
	attr, _ := w32.DwmGetWindowAttribute(hwnd, w32.DWMWA_EXTENDED_FRAME_BOUNDS)
	x, y := w32.ClientToScreen(hwnd, 0, 0)
	rect = *attr.(*w32.RECT)
	rect.Top = int32(y)
	rect.Left = int32(x)
	return
}

func getActiveWindow() (w *Window) {
	r1, _, _ := getForegroundWindow.Call()
	if r1 == 0 {
		return ActiveWindow
	}
	w = ActiveWindow
	if w == nil || w.Hwnd != w32.HWND(r1) {
		w = createWindow(w32.HWND(r1))
	}
	if w == nil {
		return
	}
	w.Rect = getWindowRect(w.Hwnd)

	w.Width = int(w.Rect.Right - w.Rect.Left)
	w.Height = int(w.Rect.Bottom - w.Rect.Top)
	ActiveWindow = w
	return
}

var nameRegex = regexp.MustCompile(`Puzzle Pirates - ([^ ]+)`)

func (w *Window) UpdateName() {
	if w == nil {
		return
	}
	title := w32.GetWindowText(w.Hwnd)
	w.Title = title
	if !strings.Contains(title, "Puzzle") {
		return
	}
	nameMatch := nameRegex.FindStringSubmatch(title)
	name := title
	if len(nameMatch) > 1 {
		name = nameMatch[1]
	}
	w.Name = name
}

var windowSwitchListeners = make([]func(*Window), 0)

func OnWindowSwitch(listener func(*Window)) {
	windowSwitchListeners = append(windowSwitchListeners, listener)
}

func createWindow(hWnd w32.HWND) *Window {
	title := w32.GetWindowText(hWnd)
	if !strings.Contains(title, "Puzzle") {
		return ActiveWindow
	}
	w := &Window{Hwnd: hWnd, Title: title, Lock: make(chan struct{}, 1)}
	w.UpdateName()
	for _, listener := range windowSwitchListeners {
		listener(w)
	}
	return w
}

var frameTimer *time.Ticker
var frameListeners = make([]func(time.Time), 0)

func OnFrame(listener func(time.Time)) {
	frameListeners = append(frameListeners, listener)
}

// scan for new puzzle windows
func ScanWindows() {
	frameTimer = time.NewTicker(time.Second / time.Duration(ScansPerSecond))
	// update window size at specified polling rate
	for range frameTimer.C {
		getActiveWindow()
		now := time.Now()
		for _, listener := range frameListeners {
			listener(now)
		}
	}
}
