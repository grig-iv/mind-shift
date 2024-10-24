package main

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type client struct {
	conn    *xgb.Conn
	window  xproto.Window
	geom    geometry
	tagMask uint16
}

func (c *client) changeGeometry(newGeom geometry) {
	newGeom = geometry{
		x:      max(newGeom.x, 0),
		y:      max(newGeom.y, 0),
		width:  max(newGeom.width, 1),
		height: max(newGeom.height, 1),
	}

	vals := []uint32{
		uint32(newGeom.x),
		uint32(newGeom.y),
		uint32(newGeom.width),
		uint32(newGeom.height),
	}

	mask := xproto.ConfigWindowX | xproto.ConfigWindowY |
		xproto.ConfigWindowWidth | xproto.ConfigWindowHeight

	xproto.ConfigureWindow(c.conn, c.window, uint16(mask), vals)

	c.geom = newGeom
}

func (c *client) hasTag(tagId uint16) bool {
	return c.tagMask&tagId != 0
}

func (c *client) replaceTag(oldTagMask, newTagMask uint16) {
	c.tagMask = c.tagMask & ^oldTagMask
	c.tagMask = c.tagMask | newTagMask
}
