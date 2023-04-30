package treasure

import (
	"image/color"
	"sort"
	"time"

	"github.com/amh11706/treasure-counter/window"

	g "github.com/AllenDang/giu"
	"github.com/TheTitanrain/w32"
	"github.com/google/uuid"
)

type boatStatus struct {
	id             uuid.UUID
	moves          [4]byte
	treasureLoaded [3]byte
	window         *window.Window
	freshTurnTime  time.Time
	turnsAlive     int8
	timesSank      byte
	Treasure       byte
	Damage         byte
	spawnVerified  bool
	freshTurn      bool
	checkRespawn   bool
}

func getBoat(w *window.Window, create bool) *boatStatus {
	if w == nil {
		return &boatStatus{}
	}
	currentEntry.boatLock <- struct{}{}
	defer func() { <-currentEntry.boatLock }()
	for _, boat := range currentEntry.boats {
		if boat.window.Hwnd == w.Hwnd {
			return boat
		}
	}
	if !create {
		return &boatStatus{}
	}
	boat := &boatStatus{window: w, id: uuid.New()}
	currentEntry.boats = append(currentEntry.boats, boat)
	return boat
}

func (boat *boatStatus) Respawn(findRoute bool) {
	boat.turnsAlive = -1
	boat.Treasure = 0
	boat.Damage = 0
	boat.moves = [4]byte{}
	boat.spawnVerified = true
	go sortBoats()
}

func (boat *boatStatus) Reset() {
	boat.Respawn(false)
	boat.treasureLoaded = [3]byte{}
	boat.timesSank = 0
	boat.spawnVerified = false
}

func sortBoats() {
	currentEntry.boatLock <- struct{}{}
	defer func() { <-currentEntry.boatLock }()
	sort.SliceStable(currentEntry.boats, func(i, j int) bool {
		a, b := currentEntry.boats[i], currentEntry.boats[j]
		if a.Treasure != b.Treasure {
			return a.Treasure > b.Treasure
		}
		return a.turnsAlive > b.turnsAlive
	})
}

func buildBoatTable() []*g.TableRowWidget {
	rows := make([]*g.TableRowWidget, 0, len(currentEntry.boats))
	for _, boat := range currentEntry.boats {
		rows = append(rows, g.TableRow(
			g.Label(boat.window.Name),
		))
	}
	return rows
}

var timerColor = window.HexColorToRGB(170<<16 | 243<<8 | 250)
var timerBgColor = window.HexColorToRGB(0x6f5d49)
var damageColor = window.HexColorToRGB(0x0e0b9d)

func (boat *boatStatus) updateFreshTurn(dcSrc w32.HDC, now time.Time) (freshTurn, timerSeen bool) {
	if now.Sub(boat.freshTurnTime) < 10*time.Second {
		return
	}
	freshTurn = boat.checkReset(dcSrc, boat.window.Rect)
	if freshTurn {
		boat.checkRespawn = true
	} else {
		c := boat.window.RGBAAt(dcSrc, 300, boat.window.Height-84)
		if window.FuzzyMatch(c, timerColor, 5) {
			freshTurn = true
		} else if !window.FuzzyMatch(c, timerBgColor, 5) {
			return false, false
		}
	}
	timerSeen = true
	if boat.freshTurn == freshTurn {
		return false, timerSeen
	}
	boat.freshTurn = freshTurn
	if !freshTurn {
		return
	}
	boat.freshTurnTime = time.Now()
	return
}

func (boat *boatStatus) updateDamage(dcSrc w32.HDC) {
	var damage byte
	for step := 0; step < 2; step++ {
		c := boat.window.RGBAAt(dcSrc, 294, boat.window.Height-143+step*12)
		if window.FuzzyMatch(c, damageColor, 3) {
			damage = 2 - byte(step)
			break
		}
	}
	boat.Damage = damage
}

var resetColor = window.HexColorToRGB(0x9c9279)

func (boat *boatStatus) checkReset(hdc w32.HDC, rect w32.RECT) bool {
	c := boat.window.RGBAAt(hdc, 275, int(rect.Bottom-rect.Top)-150)
	return window.FuzzyMatch(c, resetColor, 3)
}

func (boat *boatStatus) checkMaxDamage(hdc w32.HDC, rect w32.RECT) bool {
	c := boat.window.RGBAAt(hdc, 301, int(rect.Bottom-rect.Top-158))
	return window.FuzzyMatch(c, damageColor, 3)
}

var treasureColors = map[color.RGBA]byte{
	window.HexColorToRGB(0x755e41): 0,
	window.HexColorToRGB(0x1e5b6f): 1,
	window.HexColorToRGB(0x17819d): 2,
	window.HexColorToRGB(0x2fbdda): 3,
	window.HexColorToRGB(0x0a84b3): 4,
}

func (boat *boatStatus) scanTreasure(hdc w32.HDC, rect w32.RECT) byte {
	c := boat.window.RGBAAt(hdc, 150, int(rect.Bottom-rect.Top-40))
	for c2, t := range treasureColors {
		if window.FuzzyMatch(c, c2, 10) {
			return t
		}
	}
	return 0
}

var moveColors = map[color.RGBA]byte{
	window.HexColorToRGB(161<<16 | 110<<8 | 23): 1,
	window.HexColorToRGB(101<<16 | 137<<8 | 46): 2,
	window.HexColorToRGB(46<<16 | 114<<8 | 138): 3,
}

func (boat *boatStatus) readMoves(dcSrc w32.HDC) [4]byte {
	moveRead := [4]byte{}
	for i := range moveRead {
		c := boat.window.RGBAAt(dcSrc, 217, boat.window.Height-130+34*i)
		for c2, move := range moveColors {
			if window.FuzzyMatch(c, c2, 3) {
				moveRead[i] = move
				break
			}
		}
	}
	return moveRead
}
