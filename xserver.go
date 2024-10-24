package main

import (
	"log"
	"strings"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type xserver struct {
	conn   *xgb.Conn
	setup  *xproto.SetupInfo
	screen *xproto.ScreenInfo
	root   xproto.Window

	atomMu    sync.Mutex
	atomCache map[atomName]xproto.Atom
}

func newXserver() *xserver {
	var err error = nil

	x := &xserver{}

	x.conn, err = xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	x.setup = xproto.Setup(x.conn)
	x.screen = x.setup.DefaultScreen(x.conn)
	x.root = x.screen.Root
	x.atomCache = make(map[atomName]xproto.Atom)

	return x
}

func (x *xserver) instanceAndClass(win xproto.Window) (string, string) {
	wmClass, _ := xproto.GetProperty(
		x.conn,
		false,
		win,
		xproto.AtomWmClass,
		xproto.AtomString,
		0,
		8,
	).Reply()

	instAndClass := strings.Split(string(wmClass.Value), "\x00")
	log.Println("Win:", win, "| inst and class:", instAndClass)
	switch len(instAndClass) {
	case 0:
		return "", ""
	case 1:
		return instAndClass[0], ""
	default:
		return instAndClass[0], instAndClass[1]
	}
}

func (x *xserver) checkOtherWm() error {
	values := []uint32{
		xproto.EventMaskSubstructureRedirect |
			xproto.EventMaskSubstructureNotify |
			xproto.EventMaskButtonPress |
			xproto.EventMaskPointerMotion |
			xproto.EventMaskEnterWindow |
			xproto.EventMaskLeaveWindow |
			xproto.EventMaskStructureNotify |
			xproto.EventMaskPropertyChange,
	}

	return xproto.ChangeWindowAttributesChecked(
		x.conn,
		x.root,
		xproto.CwEventMask,
		values,
	).Check()
}

func (x *xserver) deleteProperty(window xproto.Window, atomName atomName) {
	atom, err := x.atom(atomName)
	if err != nil {
		xproto.DeleteProperty(x.conn, window, atom)
	}
}
