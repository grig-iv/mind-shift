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

func (c *client) hasTag(tagId uint16) bool {
	return c.tagMask&tagId != 0
}

func (c *client) replaceTag(oldTagMask, newTagMask uint16) {
	c.tagMask = c.tagMask & ^oldTagMask
	c.tagMask = c.tagMask | newTagMask
}
