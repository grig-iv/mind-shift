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

func (wm *wm) quit() {
	wm.isRunning = false
}

func (wm *wm) killClient() {
	if wm.focusedClient == nil {
		return
	}

	xproto.GrabServer(x.Conn)
	xproto.SetCloseDownMode(x.Conn, xproto.CloseDownDestroyAll)
	xproto.KillClient(x.Conn, uint32(wm.focusedClient.window))
	xproto.UngrabServer(x.Conn)
}

func (wm *wm) getPrevTag() *tag {
	for i := range wm.tags {
		i = len(wm.tags) - i - 1
		if wm.tags[i] == wm.currTag {
			return wm.tags[(i-1+len(wm.tags))%len(wm.tags)]
		}
	}
	return nil
}

func (wm *wm) getNextTag() *tag {
	for i := range wm.tags {
		if wm.tags[i] == wm.currTag {
			return wm.tags[(i+1)%len(wm.tags)]
		}
	}
	return nil
}

func (wm *wm) goToPrevTag() {
	wm.view(wm.getPrevTag())
}

func (wm *wm) goToNextTag() {
	wm.view(wm.getNextTag())
}

func (wm *wm) view(tag *tag) {
	log.Println("[wm.view] tag number:", tag.index()+1)

	wm.kbLayout.update(wm.currTag.id, tag.id)
	wm.currTag = tag

	tagClients := make([]*client, 0)
	tagFullscreenClients := make([]*client, 0)

	for _, c := range wm.clients {
		if !c.hasTag(wm.currTag.id) {
			xproto.ConfigureWindow(
				x.Conn,
				c.window,
				uint16(xproto.ConfigWindowX),
				[]uint32{uint32(c.geom.Width * -2)},
			)
		} else if c.isFullscreen {
			tagFullscreenClients = append(tagFullscreenClients, c)
		} else {
			tagClients = append(tagClients, c)
		}
	}

	screenGeom := domain.Geometry{
		X:      0,
		Y:      0,
		Width:  int(x.Screen.WidthInPixels),
		Height: int(x.Screen.HeightInPixels),
	}

	tagAreaGeom := wm.bar.adjustScreenGeometry(screenGeom)

	geoms := wm.currTag.currLaout.arrange(tagAreaGeom, len(tagClients))
	for i, c := range tagClients {
		c.geom = geoms[i]
		c.geom.Width -= borderWidth * 2
		c.geom.Height -= borderWidth * 2
		x.ChangeGeometry(c.window, c.geom)
	}

	for _, c := range tagFullscreenClients {
		x.ChangeGeometry(c.window, screenGeom)
	}

	if (wm.focusedClient == nil || wm.focusedClient.hasTag(tag.id) == false) && len(tagClients) > 0 {
		wm.unfocus(wm.focusedClient)

		for _, c := range wm.clients {
			if c.isFullscreen {
				wm.focus(c)
				return
			}
		}

		wm.focus(tagClients[0])
	}
}

func (wm *wm) moveToPrevTag() {
	wm.moveToTag(wm.getPrevTag())
}

func (wm *wm) moveToNextTag() {
	wm.moveToTag(wm.getNextTag())
}

func (wm *wm) moveToTag(tag *tag) {
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

func (wm *wm) gotoWindow(targetClass string) bool {
	client, ok := wm.findClientByClass(targetClass)
	if ok {
		tag, _ := wm.findTag(client.tagMask)
		if wm.currTag != tag {
			wm.view(tag)
		}
		if wm.focusedClient != client {
			wm.unfocus(wm.focusedClient)
			wm.focus(client)
		}
	}
	return ok
}

func (wm *wm) gotoWindowOrSpawn(targetClass string, command string, args ...string) {
	found := wm.gotoWindow(targetClass)
	if found {
		return
	}

	go func() {
		exec.Command(command, args...).Start()

		for range 20 {
			time.Sleep(time.Millisecond * 100)
			found = wm.gotoWindow(targetClass)
			if found {
				return
			}
		}
	}()
}
