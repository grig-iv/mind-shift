package main

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"
)

func (wm *windowManager) onMapRequest(event xproto.MapRequestEvent) {
	winAttrs, err := xproto.GetWindowAttributes(wm.x.conn, event.Window).Reply()
	if err != nil {
		log.Println(err)
		return
	}

	if winAttrs.OverrideRedirect {
		return
	}

	if _, ok := wm.windowToClient(event.Window); ok {
		return
	}

	client := wm.addClient(event.Window)
	wm.view(wm.currTag)

	if client != nil {
		wm.focus(client)
	}

	xproto.MapWindow(wm.x.conn, event.Window)
}

func (wm *windowManager) onConfigureNotify(event xproto.ConfigureNotifyEvent) {
	if event.Window == wm.x.root() {
		log.Println("TODO: add handler for when root geometry changed")
	}
}

func (wm *windowManager) onDestroyNotify(event xproto.DestroyNotifyEvent) {
	if client, ok := wm.windowToClient(event.Window); ok {
		wm.removeClient(client)
		wm.view(wm.currTag)
	}
}

func (wm *windowManager) onConfigureRequest(event xproto.ConfigureRequestEvent) {
	client, ok := wm.windowToClient(event.Window)

	if ok {
		newEvent := xproto.ConfigureNotifyEvent{
			Event:            client.window,
			Window:           client.window,
			AboveSibling:     0,
			X:                int16(client.geom.x),
			Y:                int16(client.geom.y),
			Width:            uint16(client.geom.width),
			Height:           uint16(client.geom.height),
			BorderWidth:      1,
			OverrideRedirect: false,
		}
		xproto.SendEvent(
			wm.x.conn,
			false,
			wm.x.root(),
			xproto.EventMaskStructureNotify,
			string(newEvent.Bytes()),
		)
	} else {
		values := []uint32{
			uint32(event.X),
			uint32(event.Y),
			uint32(event.Width),
			uint32(event.Height),
			uint32(event.BorderWidth),
			uint32(event.Sibling),
			uint32(event.StackMode),
		}
		xproto.ConfigureWindow(wm.x.conn, event.Window, event.ValueMask, values)
	}
}

func (wm *windowManager) onButtonPressEvent(event xproto.ButtonPressEvent) {
	var client *client = nil
	for _, c := range wm.clients {
		if c.window == event.Child {
			client = c
			break
		}
	}

	if client == nil {
		return
	}

	wm.focus(client)

	xproto.AllowEvents(
		wm.x.conn,
		xproto.AllowReplayPointer,
		xproto.TimeCurrentTime,
	)
}
