package main

import (
	"context"
	"syscall"
	"treasure-counter/config"
	"treasure-counter/sheets"
	"treasure-counter/treasure"
	"treasure-counter/window"

	g "github.com/AllenDang/giu"
	"github.com/amh11706/shutdown"
)

var (
	modShcore                  = syscall.NewLazyDLL("Shcore.dll")
	procSetProcessDpiAwareness = modShcore.NewProc("SetProcessDpiAwareness")

	mainWnd *g.MasterWindow
)

func init() {
	// We need to call SetProcessDpiAwareness so that Windows API calls will
	// tell us the scale factor for our monitor so that our screenshot works
	// on hi-res displays.
	_, _, _ = procSetProcessDpiAwareness.Call(uintptr(2)) // PROCESS_PER_MONITOR_DPI_AWARE
	config.Read()
	sheets.EncryptionSecret = encryptionSecret
}

func main() {
	println("started")
	go shutdown.Watch()
	go window.ScanWindows()

	mainWnd = g.NewMasterWindow("Treasure Counter", 350, 600, g.MasterWindowFlagsFloating)

	mainWnd.SetPos(config.DefaultSettings.PosX, config.DefaultSettings.PosY)
	shutdown.AddTask(func(_ context.Context) {
		if mainWnd != nil {
			mainWnd.Close()
			g.Update()
		}
	})
	defer func() {
		if !shutdown.Closing {
			shutdown.Trigger()
		}
	}()
	mainWnd.Run(mainLoop)
	mainWnd = nil
}

func mainLoop() {
	x, y := mainWnd.GetPos()
	if x != config.DefaultSettings.PosX || y != config.DefaultSettings.PosY {
		config.DefaultSettings.PosX = x
		config.DefaultSettings.PosY = y
		config.Save()
	}

	g.SingleWindow().Layout(treasure.BuildWidgets()...)
}
