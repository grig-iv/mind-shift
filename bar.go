package main

import (
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
	win *xproto.Window

	isVisible bool

	geom geometry
}

func newBar(x *xserver, screenMargin int) *bar {
	b := &bar{x: x}

	b.geom = geometry{
		x:      screenMargin,
		y:      screenMargin,
		width:  int(x.screen.WidthInPixels) - screenMargin*2,
		height: 20,
	}

	return b
}

func (b *bar) adjustScreenGeometry(screenGeom geometry) geometry {
	if b.win == nil || b.isVisible == false {
		return screenGeom
	}

	return screenGeom.shrinkTop(b.geom.y + b.geom.height)
}
