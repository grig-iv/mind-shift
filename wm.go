package main

import (
	"log"

	"github.com/jezek/xgb/xproto"
)

type windowManager struct {
	x *xserver

	bar *bar

	tags    []tag
	currTag tag

	clients       []*client
	focusedClient *client

	isRunning bool

	colorTable colorTable
}

func newWindowManager() *windowManager {
	var err error = nil
	const screenMargin = 16

	wm := &windowManager{}
	wm.x = newXserver()
	wm.bar = newBar(wm.x)

	masterStack := masterStack{screenMargin, 8, 0.5}
	wm.tags = []tag{
		newTagFromIndex(0, masterStack),
		newTagFromIndex(1, masterStack),
		newTagFromIndex(2, masterStack),
		newTagFromIndex(3, masterStack),
	}
	wm.currTag = wm.tags[0]

	wm.clients = make([]*client, 0)

	colormap := wm.x.setup.DefaultScreen(wm.x.conn).DefaultColormap
	wm.colorTable, err = createColorTable(wm.x.conn, colormap)
	if err != nil {
		log.Fatal(err)
	}

	return wm
}

func (wm *windowManager) dispose() {
	wm.x.conn.Close()
}

func (wm *windowManager) scan() {
	queryTree, err := xproto.QueryTree(wm.x.conn, wm.x.root).Reply()
	if err != nil {
		log.Println(err)
	}

	for _, win := range queryTree.Children {
		winAttrs, err := xproto.GetWindowAttributes(wm.x.conn, win).Reply()
		if err != nil {
			log.Println(err)
			continue
		}

		if winAttrs.OverrideRedirect {
			continue
		}

		if winAttrs.MapState == xproto.MapStateUnmapped {
			continue
		}

		transient, err := wm.x.atomProperty(win, WMTransientName)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(transient.Value) != 0 {
			log.Println("Fount transient window: ", transient)
			continue
		}

		_, class := wm.x.instanceAndClass(win)

		if wm.bar.isBar(class) {
			wm.bar.register(win)
			wm.x.changeGeometry(win, wm.bar.geom)
			continue
		}

		wm.manageClient(win, class)
	}
}

func (wm *windowManager) manageClient(win xproto.Window, class string) *client {
	geom, err := wm.x.geometry(win)
	if err != nil {
		log.Println("GetGeometry error: ", err)
		return nil
	}

	tag := wm.currTag
	if wm.isRunning == false {
		tag = wm.tags[3]
	}

	switch class {
	case "org.wezfu":
		tag = wm.tags[0]
	case "firefox":
		tag = wm.tags[1]
	case "TelegramDesktop":
		tag = wm.tags[2]
	}

	client := &client{
		wm.x.conn,
		win,
		geom,
		tag.id,
	}

	wm.grabButtons(client)

	xproto.ChangeWindowAttributes(
		wm.x.conn,
		client.window,
		xproto.CwBorderPixel,
		[]uint32{uint32(wm.colorTable[normBorder])},
	)

	xproto.ConfigureWindow(
		wm.x.conn,
		client.window,
		xproto.ConfigWindowBorderWidth,
		[]uint32{uint32(borderWidth)},
	)

	wm.clients = append(wm.clients, client)

	return client
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

	xproto.SetInputFocus(
		wm.x.conn,
		xproto.InputFocusPointerRoot,
		client.window,
		xproto.TimeCurrentTime,
	)

	xproto.ChangeWindowAttributes(
		wm.x.conn,
		client.window,
		xproto.CwBorderPixel,
		[]uint32{uint32(wm.colorTable[focusBorder])},
	)
}

func (wm *windowManager) unfocus(client *client) {
	if client == nil {
		return
	}

	log.Printf("[wm.unfocus] win: %d", client.window)

	wm.grabButtons(client)

	xproto.ChangeWindowAttributes(
		wm.x.conn,
		client.window,
		xproto.CwBorderPixel,
		[]uint32{uint32(wm.colorTable[normBorder])},
	)
}

func (wm *windowManager) ungrabButtons(client *client) {
	log.Println("[wm.ungrabButtons]", client.window)
	xproto.UngrabButton(
		wm.x.conn,
		xproto.ButtonIndexAny,
		client.window,
		xproto.ButtonMaskAny,
	)
}

func (wm *windowManager) grabButtons(client *client) {
	log.Println("[wm.grabButtons]", client.window)

	xproto.UngrabButton(
		wm.x.conn,
		xproto.ButtonIndexAny,
		client.window,
		xproto.ButtonMaskAny,
	)

	xproto.GrabButton(
		wm.x.conn,
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

func (wm *windowManager) findTag(tagId uint16) (tag, bool) {
	for _, tag := range wm.tags {
		if tag.id == tagId {
			return tag, true
		}
	}

	return tag{}, false
}

func (wm *windowManager) loop(kbm *keyboardManager) {
	go wm.x.loop()

	wm.isRunning = true
	for wm.isRunning {
		select {
		case ev, ok := <-wm.x.eventCh:
			if !ok {
				return
			}

			switch v := ev.(type) {
			case xproto.KeyPressEvent:
				log.Println("-> KeyPressEvent")
				kbm.onKeyPress(v)
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

		case err, ok := <-wm.x.errorCh:
			if !ok {
				return
			}

			log.Printf("Error: %s\n", err)
		}
	}
}

func (wm *windowManager) cleanup() {
	wm.x.deleteProperty(wm.x.root, NetActiveWindow)
}
