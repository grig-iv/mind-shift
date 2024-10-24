package main

import (
	"log"
	"strings"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type xserver struct {
	conn   *xgb.Conn
	setup  *xproto.SetupInfo
	screen *xproto.ScreenInfo
	root   xproto.Window
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
