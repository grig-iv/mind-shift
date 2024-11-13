package main

import (
	"fmt"
	"log"

	"github.com/jezek/xgb/xproto"
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

	_, class := wm.x.instanceAndClass(event.Window)

	if wm.bar.hasBar == false && wm.bar.isBar(class) {
		wm.bar.register(event.Window)
		wm.bar.onMapRequest(event)
		return
	}

	client := wm.manageClient(event.Window, class)
	if client != nil {
		wm.unfocus(wm.focusedClient)
		wm.focus(client)
	}

	wm.view(wm.currTag)
	xproto.MapWindow(wm.x.conn, event.Window)
}

func (wm *windowManager) onConfigureNotify(event xproto.ConfigureNotifyEvent) {
	if event.Window == wm.x.root {
		log.Println("TODO: add handler for when root geometry changed")
	}
}

func (wm *windowManager) onDestroyNotify(event xproto.DestroyNotifyEvent) {
	if wm.bar.win == event.Window {
		wm.bar.onDestroyNotify(event)
		wm.view(wm.currTag)
		return
	}

	if client, ok := wm.windowToClient(event.Window); ok {
		wm.removeClient(client)
		wm.view(wm.currTag)
	}
}

func (wm *windowManager) onConfigureRequest(event xproto.ConfigureRequestEvent) {
	if wm.bar.win == event.Window {
		wm.bar.onConfigureRequest(event)
		return
	}

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
			BorderWidth:      borderWidth,
			OverrideRedirect: false,
		}
		xproto.SendEvent(
			wm.x.conn,
			false,
			wm.x.root,
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
	if event.Event == event.Root && event.Detail == xproto.ButtonIndex1 {
		if event.RootX+20 > int16(wm.x.screen.WidthInPixels) {
			wm.gotoNextTag()
			return
		}
		if event.RootX < 20 {
			wm.gotoPrevTag()
			return
		}
	}

	client, ok := wm.windowToClient(event.Event)
	if !ok {
		return
	}

	wm.unfocus(wm.focusedClient)
	wm.focus(client)

	xproto.AllowEvents(
		wm.x.conn,
		xproto.AllowReplayPointer,
		xproto.TimeCurrentTime,
	)
}

func (wm *windowManager) onClientMessageEvent(event xproto.ClientMessageEvent) {
	log.Printf("[wm.onClientMessageEvent] Message: %d\n", event.Type)
	netActiveAtom, err := wm.x.atom(NetActiveWindow)
	if err == nil && event.Type == netActiveAtom {
		_, class := wm.x.instanceAndClass(event.Window)
		fmt.Print(class)
		if class != "firefox" {
			return
		}

		client, ok := wm.windowToClient(event.Window)
		fmt.Print(ok)
		if !ok {
			return
		}

		tag, ok := wm.findTag(client.tagMask)
		fmt.Print(ok)
		if !ok {
			return
		}

		if wm.currTag.id == tag.id {
			fmt.Print("already here")
			return
		}

		wm.view(tag)
		wm.unfocus(wm.focusedClient)
		wm.focus(client)
	}
}

func (wm *windowManager) onMapNotifyEvent(event xproto.MapNotifyEvent) {
	if wm.bar.win == event.Window {
		wm.view(wm.currTag)
	}
}
