package main

import (
	"fmt"
	"log"

	"github.com/grig-iv/mind-shift/socket"
	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

func (wm *wm) loop() {
	go x.Loop()
	go socket.Listen()

	wm.isRunning = true
	for wm.isRunning {
		select {
		case ev, ok := <-x.EventCh:
			if !ok {
				return
			}

			switch ev := ev.(type) {
			case xproto.MapRequestEvent:
				log.Println("-> MapRequestEvent")
				wm.onMapRequest(ev)
			case xproto.ConfigureNotifyEvent:
				wm.onConfigureNotify(ev)
			case xproto.ConfigureRequestEvent:
				log.Println("-> ConfigureRequestEvent", ev.Window)
				wm.onConfigureRequest(ev)
			case xproto.DestroyNotifyEvent:
				log.Println("-> DestroyNotifyEvent")
				wm.onDestroyNotify(ev)
			case xproto.ButtonPressEvent:
				log.Println("-> ButtonPressEvent")
				wm.onButtonPressEvent(ev)
			case xproto.ClientMessageEvent:
				wm.onClientMessageEvent(ev)
			case xproto.MapNotifyEvent:
				wm.onMapNotifyEvent(ev)
			case xproto.UnmapNotifyEvent:
				wm.unmapNotifyEvent(ev)
			case xproto.PropertyNotifyEvent:
				log.Println("-> PropertyNotifyEvent")
				wm.propertyNotifyEvent(ev)
			case xproto.CreateNotifyEvent:
			case xproto.MotionNotifyEvent:
				continue
			default:
				// log.Printf("-> [skip] %T\n", v)
			}

		case err, ok := <-x.ErrorCh:
			if !ok {
				return
			}

			log.Printf("Error: %s\n", err)

		case cmd := <-socket.CommandCh:
			wm.eval(cmd)
		}
	}
}

func (wm *wm) onMapRequest(event xproto.MapRequestEvent) {
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

	client := wm.manage(event.Window, class)

	wm.view(wm.currTag)
	xproto.MapWindow(x.Conn, event.Window)

	if client != nil {
		wm.unfocus(wm.focusedClient)
		wm.focus(client)
	}
}

func (wm *wm) onConfigureNotify(event xproto.ConfigureNotifyEvent) {
	if event.Window == x.Root {
		log.Println("TODO: add handler for when root geometry changed")
	}
}

func (wm *wm) onDestroyNotify(event xproto.DestroyNotifyEvent) {
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

func (wm *wm) onConfigureRequest(event xproto.ConfigureRequestEvent) {
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

func (wm *wm) onButtonPressEvent(event xproto.ButtonPressEvent) {
	if event.Event == event.Root && event.Detail == xproto.ButtonIndex1 {
		if event.RootX+20 > int16(x.Screen.WidthInPixels) {
			wm.goToNextTag()
			return
		}
		if event.RootX < 20 {
			wm.goToPrevTag()
			return
		}
		if event.RootY >= int16(x.Screen.HeightInPixels)-20 &&
			event.RootX > int16(x.Screen.WidthInPixels)/3 &&
			event.RootX < int16(x.Screen.WidthInPixels)*2/3 {
			if wm.focusedClient != nil && !wm.focusedClient.isFullscreen {
				wm.enableFullscreen(wm.focusedClient)
			}
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

func (wm *wm) onClientMessageEvent(event xproto.ClientMessageEvent) {
	log.Printf("[wm.onClientMessageEvent] msg: %d; win: %d\n", event.Type, event.Window)

	client, ok := wm.windowToClient(event.Window)
	if !ok {
		log.Println("Client not found", event.Window)
		return
	}

	switch event.Type {

	case x.AtomOrNone(x.NetWMState):
		netWMFullscreen := x.AtomOrNone(x.NetWMFullscreen)
		if event.Data.Data32[1] == uint32(netWMFullscreen) ||
			event.Data.Data32[2] == uint32(netWMFullscreen) {

			if event.Data.Data32[0] == 1 /* _NET_WM_STATE_ADD */ {
				wm.enableFullscreen(client)
			}

			if event.Data.Data32[0] == 2 /* _NET_WM_STATE_TOGGLE */ {
				if client.isFullscreen {
					wm.disableFullscreen(client)
				} else {
					wm.enableFullscreen(client)
				}
			}
		}

	case x.AtomOrNone(x.NetActiveWindow):
		_, class := x.InstanceAndClass(event.Window)
		if class != firefoxClass {
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

func (wm *wm) onMapNotifyEvent(event xproto.MapNotifyEvent) {
	if wm.bar.win == event.Window {
		wm.view(wm.currTag)
	}
}

func (wm *wm) unmapNotifyEvent(event xproto.UnmapNotifyEvent) {
	if client, ok := wm.windowToClient(event.Window); ok {
		wm.removeClient(client)
		wm.view(wm.currTag)
	}
}

func (wm *wm) propertyNotifyEvent(event xproto.PropertyNotifyEvent) {
	if event.State == xproto.PropertyDelete {
		return
	}

	client, ok := wm.windowToClient(event.Window)
	if !ok {
		return
	}

	switch event.Atom {
	case xproto.AtomWmHints:
		fmt.Println("TODO: add handaling for AtomWmHints")
	case xproto.AtomWmTransientFor:
		transient, ok := x.GetTransientFor(event.Window)
		if ok {
			fmt.Println("TODO: add handaling for AtomWmTransientFor", transient)
		}
	case x.AtomOrNone(x.NetWMWindowType):
		wm.updateWindowType(client)
	}
}
