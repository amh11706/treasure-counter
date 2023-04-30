package treasure

import (
	"fmt"
	"os"
	"time"
	"treasure-counter/config"

	"github.com/amh11706/logger"
)

const LogFileName = "treasure_log.txt"

var vowels = map[byte]struct{}{
	'A': {}, 'E': {}, 'I': {}, 'O': {}, 'U': {},
}

var treasureNames = []string{
	"Unknown",
	"Cuttle",
	"Locker",
	"Pod",
	"Egg",
}

func formatTreasure(treasure byte) string {
	if treasure == 0 {
		return "nothing"
	}
	prefix := "a "
	name := treasureNames[treasure]
	if _, ok := vowels[name[0]]; ok {
		prefix = "an "
	}
	return prefix + name
}

var messageLog []string
var fileLogIndex int
var lastWriteTime *time.Timer

func WriteLogToFile() {
	if !config.DefaultSettings.LogToFile {
		return
	}
	if lastWriteTime != nil {
		lastWriteTime.Stop()
	}
	lastWriteTime = time.AfterFunc(5*time.Second, appendLogToFile)
}

func appendLogToFile() {
	if len(messageLog) <= fileLogIndex {
		return
	}
	if fileLogIndex == 0 {
		messageLog = append([]string{"----- new entry -----"}, messageLog...)
	}
	f, err := os.OpenFile(LogFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error(err)
		return
	}
	defer f.Close()
	timeString := time.Now().Format("2006-01-02 15:04:05")
	for _, line := range messageLog {
		if _, err = f.WriteString(fmt.Sprintf("%s %s\n", timeString, line)); err != nil {
			logger.Error(err)
		}
		fileLogIndex++
	}
}
