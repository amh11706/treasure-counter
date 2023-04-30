package treasure

import (
	"fmt"

	"github.com/amh11706/treasure-counter/config"
	"github.com/amh11706/treasure-counter/hyperlink"
	"github.com/amh11706/treasure-counter/log"
	"github.com/amh11706/treasure-counter/sheets"
	"github.com/amh11706/treasure-counter/window"

	g "github.com/AllenDang/giu"
)

func BuildWidgets() []g.Widget {
	var treasureLoaded [3]byte
	var timesSank byte
	for _, boat := range currentEntry.boats {
		for i, treasure := range boat.treasureLoaded {
			treasureLoaded[i] += treasure
		}
		timesSank += boat.timesSank
	}
	myBoat := getBoat(window.ActiveWindow, false)

	widgets := []g.Widget{
		g.Label("Treasure Counter v1.0.0"),
		g.Button("Reset").OnClick(func() {
			currentEntry.Reset()
		}),
		g.Row(
			g.Label("Time left:"),
			g.Label(turnsToTime(int(currentEntry.turnsLeft))),
		),
		g.SliderInt(&currentEntry.oozInTurns, 0, 5).Label("Ooz in"),

		g.Separator(),
		g.TreeNode("Stats").Flags(g.TreeNodeFlagsDefaultOpen).Layout(
			g.Table().Flags(g.TableFlagsSizingStretchProp).Columns(
				g.TableColumn("Treasure"),
				g.TableColumn("All Clients"),
				g.TableColumn("This Client"),
			).Rows(
				g.TableRow(
					g.Label("Eggs"),
					g.Label(fmt.Sprintf("%d", treasureLoaded[2])),
					g.Label(fmt.Sprintf("%d", myBoat.treasureLoaded[2])),
				),
				g.TableRow(
					g.Label("Pods"),
					g.Label(fmt.Sprintf("%d", treasureLoaded[1])),
					g.Label(fmt.Sprintf("%d", myBoat.treasureLoaded[1])),
				),
				g.TableRow(
					g.Label("Lockers"),
					g.Label(fmt.Sprintf("%d", treasureLoaded[0])),
					g.Label(fmt.Sprintf("%d", myBoat.treasureLoaded[0])),
				),
			),
			g.Label("Time alive: "+turnsToTime(int(myBoat.turnsAlive))),
			g.Row(
				g.Label(fmt.Sprintf("Times sank: %d", timesSank)),
				g.Label(fmt.Sprintf("This client: %d", myBoat.timesSank)),
			),
			g.Row(
				g.Checkbox("Push to sheet", &config.DefaultSettings.PushToSheet).OnChange(config.Save),
				g.Button("Open sheet").OnClick(func() {
					_ = hyperlink.Open(sheets.SheetUrl)
				}),
			),
		),

		g.Separator(),
		g.TreeNode("Clients").Layout(
			g.Table().Rows(
				buildBoatTable()...,
			).Flags(g.TableFlagsSizingStretchProp).Columns(
				g.TableColumn("Name"),
			),
		),

		g.Separator(),
		g.TreeNode("Log").Layout(
			g.Checkbox("Output to "+log.LogFileName, &config.DefaultSettings.LogToFile).OnChange(config.Save),
			g.ListBox("Log", log.MessageLog).Size(300, 150),
		),
	}
	return widgets
}
