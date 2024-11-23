package main

import (
	"log"

	"github.com/grig-iv/mind-shift/socket"
	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

type windowManager struct {
	bar *bar

	tags    []*tag
	currTag *tag

	clients       []*client
	focusedClient *client

	isRunning bool

	colorTable x.ColorTable
}

func newWindowManager() *windowManager {
	var err error = nil
	const screenMargin = 16

	x.Initialize()

	wm := &windowManager{}
	wm.bar = newBar()

	masterStack := masterStack{screenMargin, 8, 0.5}
	wm.tags = []*tag{
		newTagFromIndex(0, masterStack),
		newTagFromIndex(1, masterStack),
		newTagFromIndex(2, masterStack),
		newTagFromIndex(3, masterStack),
	}
	wm.currTag = wm.tags[0]

	wm.clients = make([]*client, 0)

	colormap := x.Setup.DefaultScreen(x.Conn).DefaultColormap
	wm.colorTable, err = x.CreateColorTable(x.Conn, colormap)
	if err != nil {
		log.Fatal(err)
	}

	return wm
}

func (wm *windowManager) dispose() {
	x.Conn.Close()
}

func (wm *windowManager) manageClient(win xproto.Window, class string) *client {
	geom, err := x.Geometry(win)
	if err != nil {
		log.Println("GetGeometry error: ", err)
		return nil
	}

	tag := wm.currTag
	if wm.isRunning == false {
		tag = wm.tags[3]
	}

	client := &client{
		win,
		geom,
		tag.id,
		false,
	}

	transient, err := x.AtomProperty(win, x.WMTransientName)
	if err != nil {
		log.Println(err)
	}
	if len(transient.Value) != 0 {
		log.Println("Fount transient window: ", transient)
	}

	wm.applyRules(client, class)
	wm.updateWindowType(client)
	wm.grabButtons(client)

	x.ChangeBorderColor(client.window, wm.colorTable[x.NormBorder])
	x.ChangeBorderWidth(client.window, borderWidth)

	wm.clients = append(wm.clients, client)

	return client
}

func (wm *windowManager) applyRules(client *client, class string) {
	for _, r := range rules {
		if r.class != class {
			continue
		}

		if _, ok := wm.findTag(r.tagId); ok {
			client.tagMask = r.tagId
		}
	}
}

func (wm *windowManager) updateWindowType(client *client) {
	state, err := x.AtomProperty(client.window, x.NetWMState)
	if err == nil && state.Type == x.AtomOrNone(x.NetWMFullscreen) {
		wm.enableFullscreen(client)
	}

	wtype, err := x.AtomProperty(client.window, x.NetWMWindowType)
	if err == nil && wtype.Type == x.AtomOrNone(x.NetWMWindowTypeDialog) {
		// handle dialogs
	}
}

func (wm *windowManager) removeClient(oldClient *client) {

	if oldClient == wm.focusedClient {
		wm.focusedClient = nil
	}

	for i, c := range wm.clients {
		if c == oldClient {
			wm.clients = append(wm.clients[:i], wm.clients[i+1:]...)
			break
		}
	}
}

func (wm *windowManager) findClients(tagId uint16) []*client {
	clients := make([]*client, 0)

	for _, c := range wm.clients {
		if c.hasTag(tagId) {
			clients = append(clients, c)
		}
	}

	return clients
}

func (wm *windowManager) windowToClient(window xproto.Window) (*client, bool) {
	for _, c := range wm.clients {
		if c.window == window {
			return c, true
		}
	}

	return nil, false
}

func (wm *windowManager) isClientVisible(client *client) bool {
	return client.hasTag(wm.currTag.id)
}

func (wm *windowManager) focus(client *client) {
	if wm.focusedClient == client {
		return
	}

	if client == nil || !wm.isClientVisible(client) {
		log.Println("Cant focuse client")
		return
	}

	log.Printf("[wm.focus] win: %d", client.window)

	wm.focusedClient = client
	wm.ungrabButtons(client)

	x.ChangeBorderColor(client.window, wm.colorTable[x.FocusBorder])
	x.SetInputFocus(client.window)
	x.ChangeProperty32(
		x.Root,
		x.AtomOrNone(x.NetActiveWindow),
		xproto.AtomWindow,
		uint(client.window),
	)
}

func (wm *windowManager) unfocus(client *client) {
	if client == nil {
		return
	}

	log.Printf("[wm.unfocus] win: %d", client.window)

	wm.grabButtons(client)

	x.ChangeBorderColor(client.window, wm.colorTable[x.NormBorder])
}

func (wm *windowManager) ungrabButtons(client *client) {
	log.Println("[wm.ungrabButtons]", client.window)
	xproto.UngrabButton(
		x.Conn,
		xproto.ButtonIndexAny,
		client.window,
		xproto.ButtonMaskAny,
	)
}

func (wm *windowManager) grabButtons(client *client) {
	log.Println("[wm.grabButtons]", client.window)

	xproto.UngrabButton(
		x.Conn,
		xproto.ButtonIndexAny,
		client.window,
		xproto.ButtonMaskAny,
	)

	xproto.GrabButton(
		x.Conn,
		false,
		client.window,
		xproto.EventMaskButtonPress|xproto.EventMaskButtonRelease,
		xproto.GrabModeSync,
		xproto.GrabModeSync,
		xproto.WindowNone,
		xproto.CursorNone,
		xproto.ButtonIndexAny,
		xproto.ButtonMaskAny,
	)
}

func (wm *windowManager) findTag(tagId uint16) (*tag, bool) {
	for _, tag := range wm.tags {
		if tag.id == tagId {
			return tag, true
		}
	}

	return nil, false
}

func (wm *windowManager) loop() {
	go x.Loop()
	go socket.Listen()

	wm.isRunning = true
	for wm.isRunning {
		select {
		case ev, ok := <-x.EventCh:
			if !ok {
				return
			}

			switch v := ev.(type) {
			case xproto.MapRequestEvent:
				log.Println("-> MapRequestEvent")
				wm.onMapRequest(v)
			case xproto.ConfigureNotifyEvent:
				wm.onConfigureNotify(v)
			case xproto.ConfigureRequestEvent:
				log.Println("-> ConfigureRequestEvent")
				wm.onConfigureRequest(v)
			case xproto.DestroyNotifyEvent:
				log.Println("-> DestroyNotifyEvent")
				wm.onDestroyNotify(v)
			case xproto.ButtonPressEvent:
				log.Println("-> ButtonPressEvent")
				wm.onButtonPressEvent(v)
			case xproto.ClientMessageEvent:
				wm.onClientMessageEvent(v)
			case xproto.MapNotifyEvent:
				wm.onMapNotifyEvent(v)
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

func (wm *windowManager) cleanup() {
	x.DeleteProperty(x.Root, x.NetActiveWindow)
}
