package treasure

import (
	"fmt"
	"time"

	"github.com/amh11706/treasure-counter/config"
	"github.com/amh11706/treasure-counter/log"
	"github.com/amh11706/treasure-counter/sheets"
	"github.com/amh11706/treasure-counter/window"

	g "github.com/AllenDang/giu"
	"github.com/TheTitanrain/w32"
	"github.com/amh11706/logger"
	"github.com/google/uuid"
)

func init() {
	currentEntry.Reset()

	window.OnWindowSwitch(func(w *window.Window) {
		getBoat(w, true)
	})
	window.OnFrame(scanBoats)
}

func turnsToTime(turns int) string {
	if turns < 0 {
		return "0:00"
	}
	seconds := turns * 20
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

type KhEntry struct {
	id            uuid.UUID
	oozInTurns    int32
	turnsLeft     int32
	timerLastSeen time.Time
	boats         []*boatStatus
	boatLock      chan struct{}
}

var currentEntry = &KhEntry{boatLock: make(chan struct{}, 1)}

func (e *KhEntry) Reset() {
	e.id = uuid.New()
	e.oozInTurns = 5
	e.turnsLeft = 91
	log.ResetBuffer()

	e.boatLock <- struct{}{}
	defer func() { <-e.boatLock }()
	for _, boat := range e.boats {
		boat.Reset()
	}
}

var freshTurnTime = time.Now().Add(-time.Minute)

func incrementTurn(now time.Time) {
	if now.Sub(freshTurnTime) < 10*time.Second {
		return
	}
	if now.Sub(currentEntry.timerLastSeen) > 10*time.Second || currentEntry.turnsLeft < -1 {
		// if we haven't seen the timer in a while or we've seen too many turns, assume we're in a new entry
		currentEntry.Reset()
	}
	freshTurnTime = now
	currentEntry.turnsLeft--
	if currentEntry.oozInTurns == 0 {
		currentEntry.oozInTurns = 4
	} else if currentEntry.oozInTurns < 5 {
		currentEntry.oozInTurns--
	}
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func (boat *boatStatus) scanHud(now time.Time) (treasureRow []interface{}) {
	w := boat.window
	if w == nil {
		return
	}
	w.Lock <- struct{}{}
	defer func() { <-w.Lock }()
	hdc := w32.GetDC(w.Hwnd)
	defer w32.ReleaseDC(w.Hwnd, hdc)
	if hdc == 0 {
		return
	}

	freshTurn, timerSeen := boat.updateFreshTurn(hdc, now)
	if timerSeen {
		if now.Sub(currentEntry.timerLastSeen) > 15*time.Second {
			currentEntry.Reset()
		}
		currentEntry.timerLastSeen = now
	}
	if freshTurn {
		if boat.turnsAlive < 0 {
			boat.turnsAlive = 0
		}
		incrementTurn(now)
		if boat.turnsAlive > 0 || boat.moves != [4]byte{} {
			boat.turnsAlive++
		}
	}
	if !boat.checkReset(hdc, w.Rect) {
		boat.Treasure = boat.scanTreasure(hdc, w.Rect)
		// we only need to update moves and damage if we didn't just reset the boat
		if now.Sub(boat.freshTurnTime) > 2300*time.Millisecond {
			boat.updateDamage(hdc)
			boat.moves = boat.readMoves(hdc)
		}
		return
	}
	if boat.turnsAlive <= 0 || !boat.checkRespawn || now.Sub(boat.freshTurnTime) < 500*time.Millisecond {
		return
	}
	boat.checkRespawn = false
	defer boat.Respawn(true)
	title := boat.window.Title
	boat.window.UpdateName()
	if title != boat.window.Title {
		boat.spawnVerified = false
	}
	if !boat.freshTurn {
		boat.turnsAlive++
	}
	if boat.checkMaxDamage(hdc, w.Rect) {
		boat.timesSank++
		log.AppendLog(fmt.Sprintf("%s sank in %s", boat.window.Name, turnsToTime(int(boat.turnsAlive))))
		return
	}
	if boat.Treasure == 0 {
		return
	}
	log.AppendLog(fmt.Sprintf("%s loaded %s in %s", boat.window.Name, log.FormatTreasure(boat.Treasure), turnsToTime(int(boat.turnsAlive))))
	if boat.Treasure > 1 {
		boat.treasureLoaded[boat.Treasure-2]++
	}
	if boat.turnsAlive < 4 || !boat.spawnVerified {
		boat.turnsAlive = 90
	}

	return []interface{}{
		currentEntry.id.String(), boat.id.String(), boat.window.Name,
		boolToByte(boat.Treasure == 2), boolToByte(boat.Treasure == 3), boolToByte(boat.Treasure == 4),
		boat.turnsAlive, boat.Damage, time.Now().Unix(), boolToByte(boat.Treasure == 1),
	}
}

func scanBoats(now time.Time) {
	rowsToPush := make([][]interface{}, 0)
	for _, boat := range currentEntry.boats {
		row := boat.scanHud(now)
		if len(row) > 0 {
			rowsToPush = append(rowsToPush, row)
		}
	}

	if len(rowsToPush) > 0 {
		if config.DefaultSettings.PushToSheet {
			go func() {
				logger.Check(sheets.PushRowsToSheet(rowsToPush))
			}()
		}
	}
	g.Update()
}
