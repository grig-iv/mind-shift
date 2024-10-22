package main

import (
	"github.com/BurntSushi/xgb/xproto"
)

type direction byte

const (
	Forward  direction = iota
	Backward direction = iota
)

func (wm *windowManager) quit() {
	wm.isRunning = false
}

func (wm *windowManager) killClient() {
	if wm.focusedClient == nil {
		return
	}

	xproto.GrabServer(wm.x.conn)
	xproto.SetCloseDownMode(wm.x.conn, xproto.CloseDownDestroyAll)
	xproto.KillClient(wm.x.conn, uint32(wm.focusedClient.window))
	xproto.UngrabServer(wm.x.conn)
}

func (wm *windowManager) getPrevTag() tag {
	for i := range wm.tags {
		i = len(wm.tags) - i - 1
		if wm.tags[i] == wm.currTag {
			return wm.tags[(i-1+len(wm.tags))%len(wm.tags)]
		}
	}
	return tag{}
}

func (wm *windowManager) getNextTag() tag {
	for i := range wm.tags {
		if wm.tags[i] == wm.currTag {
			return wm.tags[(i+1)%len(wm.tags)]
		}
	}
	return tag{}
}

func (wm *windowManager) gotoPrevTag() {
	wm.view(wm.getPrevTag())
}

func (wm *windowManager) gotoNextTag() {
	wm.view(wm.getNextTag())
}

func (wm *windowManager) view(tag tag) {
	wm.currTag = tag

	tagClients := make([]*client, 0)
	for _, c := range wm.clients {
		if c.isOnTag(wm.currTag.id) {
			tagClients = append(tagClients, c)
		} else {
			xproto.ConfigureWindow(
				c.conn,
				c.window,
				uint16(xproto.ConfigWindowX),
				[]uint32{uint32(c.geom.width * -2)},
			)
		}
	}

	screen := wm.setup.DefaultScreen(wm.x.conn)
	screenGeom := geometry{
		x:      0,
		y:      0,
		width:  int(screen.WidthInPixels),
		height: int(screen.HeightInPixels),
	}

	geoms := wm.currTag.currLaout.arrange(screenGeom, len(tagClients))
	for i, c := range tagClients {
		c.changeGeometry(geoms[i])
	}

	if len(tagClients) > 0 {
		wm.focus(tagClients[0])
	}
}

func (wm *windowManager) moveToPrevTag() {
	wm.moveToTag(wm.getPrevTag())
}

func (wm *windowManager) moveToNextTag() {
	wm.moveToTag(wm.getNextTag())
}

func (wm *windowManager) moveToTag(tag tag) {
	if wm.focusedClient == nil || wm.currTag.id == tag.id {
		return
	}

	wm.focusedClient.tagMask = wm.focusedClient.tagMask & ^wm.currTag.id
	wm.focusedClient.tagMask = wm.focusedClient.tagMask | tag.id

	for i, c := range wm.clients {
		if c == wm.focusedClient {
			wm.clients = append(wm.clients[:i], wm.clients[i+1:]...)
			wm.clients = append(wm.clients, c)
			break
		}
	}

	wm.view(tag)
}
