package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var MainPages = tview.NewPages()

func init() {
	MainPages.AddPage("Connections", NewConnectionPages().Flex, true, true)
	MainPages.SetBackgroundColor(tcell.ColorDefault)
}
