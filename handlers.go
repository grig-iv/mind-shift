package main

import (
	"fmt"
	"log"

	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

func (wm *windowManager) onMapRequest(event xproto.MapRequestEvent) {
	winAttrs, err := xproto.GetWindowAttributes(x.Conn, event.Window).Reply()
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

	_, class := x.InstanceAndClass(event.Window)

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
	xproto.MapWindow(x.Conn, event.Window)
}

func (wm *windowManager) onConfigureNotify(event xproto.ConfigureNotifyEvent) {
	if event.Window == x.Root {
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
			X:                int16(client.geom.X),
			Y:                int16(client.geom.Y),
			Width:            uint16(client.geom.Width),
			Height:           uint16(client.geom.Height),
			BorderWidth:      borderWidth,
			OverrideRedirect: false,
		}
		xproto.SendEvent(
			x.Conn,
			false,
			x.Root,
			xproto.EventMaskStructureNotify,
			string(newEvent.Bytes()),
		)
	} else {
		values := make([]uint32, 0)

		if event.ValueMask&xproto.ConfigWindowX != 0 {
			values = append(values, uint32(event.X))
		}

		if event.ValueMask&xproto.ConfigWindowY != 0 {
			values = append(values, uint32(event.Y))
		}

		if event.ValueMask&xproto.ConfigWindowWidth != 0 {
			values = append(values, uint32(event.Width))
		}

		if event.ValueMask&xproto.ConfigWindowHeight != 0 {
			values = append(values, uint32(event.Height))
		}

		if event.ValueMask&xproto.ConfigWindowBorderWidth != 0 {
			values = append(values, uint32(event.BorderWidth))
		}

		if event.ValueMask&xproto.ConfigWindowSibling != 0 {
			values = append(values, uint32(event.Sibling))
		}

		if event.ValueMask&xproto.ConfigWindowStackMode != 0 {
			values = append(values, uint32(event.StackMode))
		}

		xproto.ConfigureWindow(x.Conn, event.Window, event.ValueMask, values)
	}
}

func (wm *windowManager) onButtonPressEvent(event xproto.ButtonPressEvent) {
	if event.Event == event.Root && event.Detail == xproto.ButtonIndex1 {
		if event.RootX+20 > int16(x.Screen.WidthInPixels) {
			wm.goToNextTag()
			return
		}
		if event.RootX < 20 {
			wm.goToPrevTag()
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
		x.Conn,
		xproto.AllowReplayPointer,
		xproto.TimeCurrentTime,
	)
}

func (wm *windowManager) onClientMessageEvent(event xproto.ClientMessageEvent) {
	log.Printf("[wm.onClientMessageEvent] Message: %d\n", event.Type)

	client, ok := wm.windowToClient(event.Window)
	if !ok {
		log.Println("Client not found", event.Window)
		return
	}

	switch event.Type {
	case x.AtomOrNone(x.NetWMState):
		netWMFullscreen := x.AtomOrNone(x.NetWMFullscreen)
		if event.Data.Data32[0] == uint32(netWMFullscreen) ||
			event.Data.Data32[1] == uint32(netWMFullscreen) {
			// implement full screen logic
		}
	case x.AtomOrNone(x.NetActiveWindow):
		_, class := x.InstanceAndClass(event.Window)
		if class != "firefox" {
			return
		}

		tag, ok := wm.findTag(client.tagMask)
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
