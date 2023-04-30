package log

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/amh11706/treasure-counter/config"

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

func FormatTreasure(treasure byte) string {
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

var MessageLog []string
var fileBuffer = bytes.NewBuffer(make([]byte, 0, 1024))

func AppendLog(message string) {
	if len(MessageLog) == 0 {
		newEntryMessage := "----- new entry -----"
		MessageLog = append(MessageLog, newEntryMessage)
		fileBuffer.WriteString(newEntryMessage + "\n")
	}
	MessageLog = append(MessageLog, message)
	timeString := time.Now().Format("2006-01-02 15:04:05")
	fileBuffer.WriteString(fmt.Sprintf("%s %s\n", timeString, message))
}

func ResetBuffer() {
	fileBuffer.Reset()
	MessageLog = MessageLog[:0]
}

func WriteLogToFile() {
	t := time.NewTicker(30 * time.Second)
	for range t.C {
		if !config.DefaultSettings.LogToFile || fileBuffer.Len() == 0 {
			continue
		}
		appendLogToFile()
	}
}

func appendLogToFile() {
	f, err := os.OpenFile(LogFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logger.Check(err) {
		return
	}
	defer f.Close()
	_, err = fileBuffer.WriteTo(f)
	logger.Check(err)
}
