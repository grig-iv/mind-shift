package main

import (
	"strings"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type xserver struct {
	conn  *xgb.Conn
	setup *xproto.SetupInfo
}

func newXserver() (*xserver, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}

	setup := xproto.Setup(conn)

	return &xserver{conn, setup}, nil
}

func (x *xserver) root() xproto.Window {
	return x.setup.DefaultScreen(x.conn).Root
}

func (x xserver) instanceAndClass(win xproto.Window) (string, string) {
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

func (x *xserver) setBorder(win xproto.Window, width int) {
	xproto.ConfigureWindow(
		x.conn,
		win,
		xproto.ConfigWindowBorderWidth,
		[]uint32{uint32(width)},
	)
}

func (x *xserver) setBorderColor(win xproto.Window, color uint16) {
	xproto.ChangeWindowAttributes(
		x.conn,
		win,
		xproto.CwBorderPixel,
		[]uint32{uint32(color)},
	)
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
		x.root(),
		xproto.CwEventMask,
		values,
	).Check()
}
