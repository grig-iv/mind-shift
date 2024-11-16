package main

import (
	"log"
	"os/exec"
	"time"

	"github.com/grig-iv/mind-shift/domain"
	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
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

	xproto.GrabServer(x.Conn)
	xproto.SetCloseDownMode(x.Conn, xproto.CloseDownDestroyAll)
	xproto.KillClient(x.Conn, uint32(wm.focusedClient.window))
	xproto.UngrabServer(x.Conn)
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
	log.Println("[wm.view] tag number:", tag.index()+1)

	wm.currTag = tag

	tagClients := make([]*client, 0)
	for _, c := range wm.clients {
		if c.hasTag(wm.currTag.id) {
			tagClients = append(tagClients, c)
		} else {
			xproto.ConfigureWindow(
				c.conn,
				c.window,
				uint16(xproto.ConfigWindowX),
				[]uint32{uint32(c.geom.Width * -2)},
			)
		}
	}

	screenGeom := domain.Geometry{
		X:      0,
		Y:      0,
		Width:  int(x.Screen.WidthInPixels),
		Height: int(x.Screen.HeightInPixels),
	}

	screenGeom = wm.bar.adjustScreenGeometry(screenGeom)

	geoms := wm.currTag.currLaout.arrange(screenGeom, len(tagClients))
	for i, c := range tagClients {
		c.geom = geoms[i]
		c.geom.Width -= borderWidth * 2
		c.geom.Height -= borderWidth * 2
		x.ChangeGeometry(c.window, c.geom)
	}

	if (wm.focusedClient == nil || wm.focusedClient.hasTag(tag.id) == false) && len(tagClients) > 0 {
		wm.unfocus(wm.focusedClient)
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

	wm.focusedClient.replaceTag(wm.currTag.id, tag.id)

	for i, c := range wm.clients {
		if c == wm.focusedClient {
			wm.clients = append(wm.clients[:i], wm.clients[i+1:]...)
			wm.clients = append(wm.clients, c)
			break
		}
	}

	wm.view(tag)
}

func (wm *windowManager) gotoWindow(targetClass string) bool {
	for _, c := range wm.clients {
		_, clientClass := x.InstanceAndClass(c.window)
		if targetClass == clientClass {
			tag, _ := wm.findTag(c.tagMask)
			if wm.currTag != tag {
				wm.view(tag)
			}
			if wm.focusedClient != c {
				wm.unfocus(wm.focusedClient)
				wm.focus(c)
			}
			return true
		}
	}

	return false
}

func (wm *windowManager) gotoWindowOrCreate(targetClass string, command string, args ...string) {
	found := wm.gotoWindow(targetClass)
	if found {
		return
	}

	go func() {
		cmd := exec.Command(command, args...)
		cmd.Start()

		for range 20 {
			time.Sleep(time.Millisecond * 100)
			found = wm.gotoWindow(targetClass)
			if found {
				return
			}
		}
	}()
}
