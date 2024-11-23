package main

import (
	"log"

	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

func (wm *windowManager) enableFullscreen(client *client) {
	log.Println("Enabling fullscreen. client:", client.window)
	x.ChangeProperty32(
		client.window,
		x.AtomOrNone(x.NetWMState),
		xproto.AtomAtom,
		uint(x.AtomOrNone(x.NetWMFullscreen)),
	)
	client.isFullscreen = true
	x.ChangeBorderWidth(client.window, 0)
	x.Raise(client.window)
	wm.view(wm.currTag)
}

func (wm *windowManager) disableFullscreen(client *client) {
	log.Println("Disabling fullscreen. client:", client.window)
	x.ChangeProperty32(
		client.window,
		x.AtomOrNone(x.NetWMState),
		xproto.AtomAtom,
		0,
	)
	client.isFullscreen = false
	x.ChangeBorderWidth(client.window, borderWidth)
	wm.view(wm.currTag)
}
