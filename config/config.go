package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/amh11706/logger"

	"github.com/TheTitanrain/w32"
)

const (
	SettingFileName  = "settings.json"
	SaveDebounceTime = 5 * time.Second
)

type Settings struct {
	FrameRate   int32 `json:"frameRate"`
	PosX        int   `json:"posX"`
	PosY        int   `json:"posY"`
	PushToSheet bool  `json:"pushToSheet"`
	LogToFile   bool  `json:"logToFile"`
}

var DefaultSettings = Settings{
	FrameRate: 60,
	PosX:      1920 - 375,
	PosY:      230,
}

func Read() {
	bytes, err := os.ReadFile(SettingFileName)
	if err != nil {
		return
	}
	if logger.Check(json.Unmarshal(bytes, &DefaultSettings)) {
		logger.Info("Invalid " + SettingFileName + "! Default settings will be used.")
	}

	// set window position to right edge of screen if saved position was invalid
	width := w32.GetSystemMetrics(w32.SM_CXSCREEN)
	if DefaultSettings.PosX < 50 || DefaultSettings.PosX > width-50 {
		DefaultSettings.PosX = width - 375
	}
	height := w32.GetSystemMetrics(w32.SM_CYSCREEN)
	if DefaultSettings.PosY < 50 || DefaultSettings.PosY > height-50 {
		DefaultSettings.PosY = 230
	}
}

var saveTimer *time.Timer

func Save() {
	if saveTimer != nil {
		saveTimer.Stop()
	}
	saveTimer = time.AfterFunc(SaveDebounceTime, func() {
		bytes, _ := json.Marshal(DefaultSettings)
		logger.Check(os.WriteFile(SettingFileName, bytes, 0644))
	})
}
