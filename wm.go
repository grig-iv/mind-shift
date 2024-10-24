package main

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"
)

type windowManager struct {
	x *xserver

	tags    []tag
	clients []*client

	currTag       tag
	focusedClient *client

	isRunning bool

	colorTable colorTable
}

func newWindowManager() (*windowManager, error) {
	const screenMargin = 16

	x := newXserver()

	masterStack := masterStack{screenMargin, 8, 0.5}

	tags := []tag{
		{1 << 0, masterStack},
		{1 << 1, masterStack},
		{1 << 2, masterStack},
		{1 << 3, masterStack},
	}

	colormap := x.setup.DefaultScreen(x.conn).DefaultColormap
	colorTable, _ := createColorTable(x.conn, colormap)

	wr := &windowManager{
		x,
		tags,
		make([]*client, 0),
		tags[0],
		nil,
		false,
		colorTable,
	}

	return wr, nil
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

		transient, err := getAtomProperty(wm.x.conn, win, WMTransientName)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(transient.Value) != 0 {
			log.Println("Fount transient window: ", transient)
			continue
		}

		_, class := wm.x.instanceAndClass(win)
		if class == "bar" {
			xproto.ChangeWindowAttributes(
				wm.x.conn,
				win,
				xproto.CwBackPixel,
				[]uint32{0x00ff00ff},
			)
		}

		wm.manageClient(win)

	}
}

func (wm *windowManager) manageClient(win xproto.Window) *client {
	drw := xproto.Drawable(win)
	geom, err := xproto.GetGeometry(wm.x.conn, drw).Reply()
	if err != nil {
		log.Println("GetGeometry error: ", err)
		return nil
	}

	_, class := wm.x.instanceAndClass(win)

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
		geometry{
			x:      int(geom.X),
			y:      int(geom.Y),
			width:  int(geom.Width),
			height: int(geom.Height),
		},
		tag.id,
	}

	xproto.ChangeWindowAttributes(
		wm.x.conn,
		client.window,
		xproto.CwBorderPixel,
		[]uint32{uint32(wm.colorTable[normBorder])},
	)

	wm.clients = append(wm.clients, client)

	return client
}

func (wm *windowManager) removeClient(oldClient *client) {
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
		if c.isOnTag(tagId) {
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
	return client.isOnTag(wm.currTag.id)
}

func (wm *windowManager) focus(client *client) {
	log.Println("[wm.focuse]", client)

	if client == nil || !wm.isClientVisible(client) {
		log.Println("Cant focuse client")
		return
	}

	if wm.focusedClient != client {
		wm.unfocus(wm.focusedClient)
	}

	wm.focusedClient = client

	xproto.SetInputFocus(
		wm.x.conn,
		xproto.InputFocusPointerRoot,
		client.window,
		xproto.TimeCurrentTime,
	)

	wm.grabButtons(client)

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

	wm.grabButtons(client)

	xproto.ChangeWindowAttributes(
		wm.x.conn,
		client.window,
		xproto.CwBorderPixel,
		[]uint32{uint32(wm.colorTable[normBorder])},
	)
}

func (wm *windowManager) grabButtons(client *client) {
	xproto.UngrabButton(
		wm.x.conn,
		xproto.ButtonIndexAny,
		client.window,
		xproto.ButtonMaskAny,
	)

	if client == wm.focusedClient {

	} else {
		xproto.GrabButton(
			wm.x.conn,
			false,
			client.window,
			xproto.ButtonPress|xproto.ButtonRelease,
			xproto.GrabModeSync,
			xproto.GrabModeSync,
			xproto.WindowNone,
			xproto.CursorNone,
			xproto.ButtonIndexAny,
			xproto.ButtonMaskAny,
		)
	}
}
