package main

import (
	"github.com/grig-iv/mind-shift/domain"
	"github.com/jezek/xgb/xproto"
)

type client struct {
	window  xproto.Window
	geom    domain.Geometry
	tagMask uint16

	isFullscreen bool
}

func (c *client) hasTag(tagId uint16) bool {
	return c.tagMask&tagId != 0
}

func (c *client) replaceTag(oldTagMask, newTagMask uint16) {
	c.tagMask = c.tagMask & ^oldTagMask
	c.tagMask = c.tagMask | newTagMask
}
