package main

import (
	"fmt"
	"log"

	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

type wm struct {
	bar      *bar
	kbLayout *kbLayout

	tags    []*tag
	currTag *tag

	clients       []*client
	focusedClient *client

	isRunning bool

	colorTable x.ColorTable
}

func newWindowManager() *wm {
	var err error = nil
	const screenMargin = 16

	x.Initialize()

	wm := &wm{}
	wm.bar = newBar()
	wm.kbLayout = newKbLayout()

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

func (wm *wm) dispose() {
	x.Conn.Close()
}

func (wm *wm) manage(win xproto.Window, class string) *client {
	log.Println("[wm.manage]", win)

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

	transient, ok := x.GetTransientFor(win)
	if ok {
		log.Println("Fount transient window: ", transient)
	}

	x.ChangeBorderColor(client.window, wm.colorTable[x.NormBorder])
	x.ChangeBorderWidth(client.window, borderWidth)

	wm.applyRules(client, class)
	wm.updateWindowType(client)
	wm.grabAnyButton(client)

	wm.clients = append(wm.clients, client)

	return client
}

func (wm *wm) applyRules(client *client, class string) {
	for _, r := range rules {
		if r.class != class {
			continue
		}

		if _, ok := wm.findTag(r.tagId); ok {
			client.tagMask = r.tagId
		}
	}
}

func (wm *wm) updateWindowType(client *client) {
	state, err := x.AtomPropertyAsAtom(client.window, x.NetWMState)
	if err == nil && state == x.AtomOrNone(x.NetWMFullscreen) {
		wm.enableFullscreen(client)
	}

	wtype, err := x.AtomPropertyAsAtom(client.window, x.NetWMWindowType)
	if err == nil && wtype == x.AtomOrNone(x.NetWMWindowTypeDialog) {
		fmt.Println("found dialog")
		// handle dialogs
	}
}

func (wm *wm) removeClient(oldClient *client) {

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

func (wm *wm) findClients(tagId uint16) []*client {
	clients := make([]*client, 0)

	for _, c := range wm.clients {
		if c.hasTag(tagId) {
			clients = append(clients, c)
		}
	}

	return clients
}

func (wm *wm) windowToClient(window xproto.Window) (*client, bool) {
	for _, c := range wm.clients {
		if c.window == window {
			return c, true
		}
	}

	return nil, false
}

func (wm *wm) isClientVisible(client *client) bool {
	return client.hasTag(wm.currTag.id)
}

func (wm *wm) focus(client *client) {
	if wm.focusedClient == client {
		return
	}

	if client == nil || !wm.isClientVisible(client) {
		log.Println("Cant focuse client")
		return
	}

	log.Printf("[wm.focus] %d", client.window)

	wm.focusedClient = client
	wm.ungrabAllButtons(client)

	x.ChangeBorderColor(client.window, wm.colorTable[x.FocusBorder])
	x.SetInputFocus(client.window)
	x.ChangeProperty32(
		x.Root,
		x.AtomOrNone(x.NetActiveWindow),
		xproto.AtomWindow,
		uint(client.window),
	)
}

func (wm *wm) unfocus(client *client) {
	if client == nil {
		return
	}

	log.Printf("[wm.unfocus] win: %d", client.window)

	wm.grabAnyButton(client)

	x.ChangeBorderColor(client.window, wm.colorTable[x.NormBorder])
}

func (wm *wm) ungrabAllButtons(client *client) {
	log.Println("[wm.ungrabAllButtons]", client.window)
	xproto.UngrabButton(
		x.Conn,
		xproto.ButtonIndexAny,
		client.window,
		xproto.ButtonMaskAny,
	)
}

func (wm *wm) grabAnyButton(client *client) {
	log.Println("[wm.grabAnyButton]", client.window)

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

func (wm *wm) findTag(tagId uint16) (*tag, bool) {
	for _, tag := range wm.tags {
		if tag.id == tagId {
			return tag, true
		}
	}

	return nil, false
}

func (wm *wm) cleanup() {
	x.DeleteProperty(x.Root, x.NetActiveWindow)
}
