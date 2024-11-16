package main

import (
	"log"

	"github.com/grig-iv/mind-shift/domain"
	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

type ctxId int

const (
	emptyTag ctxId = iota
	nonEmptyTag
	activeTag
	tagSurface
)

type bar struct {
	win xproto.Window

	isVisible bool
	hasBar    bool

	geom domain.Geometry
}

func newBar() *bar {
	b := &bar{}

	b.geom = domain.Geometry{
		X:      0,
		Y:      0,
		Width:  int(x.Screen.WidthInPixels),
		Height: 20,
	}

	return b
}

func (b *bar) isBar(class string) bool {
	return class == "mind-shift-st"
}

func (b *bar) register(win xproto.Window) {
	log.Printf("[b.register] win: %d", win)
	b.win = win
	b.hasBar = true
	b.isVisible = true
}

func (b *bar) adjustScreenGeometry(screenGeom domain.Geometry) domain.Geometry {
	if b.hasBar == false || b.isVisible == false {
		return screenGeom
	}

	return screenGeom.ShrinkTop(b.geom.Y + b.geom.Height)
}

func (b *bar) onMapRequest(event xproto.MapRequestEvent) {
	if b.win != event.Window {
		log.Fatal("not a bar")
	}

	x.ChangeGeometry(b.win, b.geom)

	xproto.MapWindow(x.Conn, event.Window)
}

func (b *bar) onConfigureRequest(event xproto.ConfigureRequestEvent) {
	if b.win != event.Window {
		log.Fatal("not a bar")
	}

	b.geom.Height = int(event.Height)

	newEvent := xproto.ConfigureNotifyEvent{
		Event:            b.win,
		Window:           b.win,
		AboveSibling:     0,
		X:                int16(b.geom.X),
		Y:                int16(b.geom.Y),
		Width:            uint16(b.geom.Width),
		Height:           uint16(b.geom.Height),
		BorderWidth:      1,
		OverrideRedirect: false,
	}

	xproto.SendEvent(
		x.Conn,
		false,
		x.Root,
		xproto.EventMaskStructureNotify,
		string(newEvent.Bytes()),
	)
}

func (b *bar) onDestroyNotify(_ xproto.DestroyNotifyEvent) {
	b.hasBar = false
	b.isVisible = false
	b.win = 0
}
