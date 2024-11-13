package main

import (
	"log"

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
	x   *xserver
	win xproto.Window

	isVisible bool
	hasBar    bool

	geom geometry
}

func newBar(x *xserver) *bar {
	b := &bar{x: x}

	b.geom = geometry{
		x:      0,
		y:      0,
		width:  int(x.screen.WidthInPixels),
		height: 20,
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

func (b *bar) adjustScreenGeometry(screenGeom geometry) geometry {
	if b.hasBar == false || b.isVisible == false {
		return screenGeom
	}

	return screenGeom.shrinkTop(b.geom.y + b.geom.height)
}

func (b *bar) onMapRequest(event xproto.MapRequestEvent) {
	if b.win != event.Window {
		log.Fatal("not a bar")
	}

	b.x.changeGeometry(b.win, b.geom)

	xproto.MapWindow(b.x.conn, event.Window)
}

func (b *bar) onConfigureRequest(event xproto.ConfigureRequestEvent) {
	if b.win != event.Window {
		log.Fatal("not a bar")
	}

	b.geom.height = int(event.Height)

	newEvent := xproto.ConfigureNotifyEvent{
		Event:            b.win,
		Window:           b.win,
		AboveSibling:     0,
		X:                int16(b.geom.x),
		Y:                int16(b.geom.y),
		Width:            uint16(b.geom.width),
		Height:           uint16(b.geom.height),
		BorderWidth:      1,
		OverrideRedirect: false,
	}

	xproto.SendEvent(
		b.x.conn,
		false,
		b.x.root,
		xproto.EventMaskStructureNotify,
		string(newEvent.Bytes()),
	)
}

func (b *bar) onDestroyNotify(_ xproto.DestroyNotifyEvent) {
	b.hasBar = false
	b.isVisible = false
	b.win = 0
}
