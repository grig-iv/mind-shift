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

	eventCh chan xgb.Event
	errorCh chan xgb.Error
}

const (
	leftPtrCursor = 68
)

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

	x.eventCh = make(chan xgb.Event)
	x.errorCh = make(chan xgb.Error)

	return x
}

func (x *xserver) loop() {
	for {
		ev, xerr := x.conn.WaitForEvent()
		if ev == nil && xerr == nil {
			close(x.eventCh)
			close(x.errorCh)
			log.Println("Both event and error are nil. Exiting...")
			return
		}

		if xerr != nil {
			x.errorCh <- xerr
		}

		if ev != nil {
			x.eventCh <- ev
		}
	}
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
	log.Printf("Win: %d, inst and class: %v", win, instAndClass)
	switch len(instAndClass) {
	case 0:
		return "", ""
	case 1:
		return instAndClass[0], ""
	default:
		return instAndClass[0], instAndClass[1]
	}
}

func (x *xserver) checkOtherWm() {
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

	err := xproto.ChangeWindowAttributesChecked(
		x.conn,
		x.root,
		xproto.CwEventMask,
		values,
	).Check()

	if err != nil {
		log.Fatal("Another window manager might aleady be running:", err)
	}
}

func (x *xserver) deleteProperty(window xproto.Window, atomName atomName) {
	atom, err := x.atom(atomName)
	if err != nil {
		xproto.DeleteProperty(x.conn, window, atom)
	}
}

func (x *xserver) geometry(window xproto.Window) (geometry, error) {
	drw := xproto.Drawable(window)
	geom, err := xproto.GetGeometry(x.conn, drw).Reply()
	if err != nil {
		return geometry{}, err
	}

	return geometry{
		x:      int(geom.X),
		y:      int(geom.Y),
		width:  int(geom.Width),
		height: int(geom.Height),
	}, nil
}

func (x *xserver) changeGeometry(win xproto.Window, geom geometry) {
	geom = geometry{
		x:      max(geom.x, 0),
		y:      max(geom.y, 0),
		width:  max(geom.width, 1),
		height: max(geom.height, 1),
	}

	vals := []uint32{
		uint32(geom.x),
		uint32(geom.y),
		uint32(geom.width),
		uint32(geom.height),
	}

	mask := xproto.ConfigWindowX | xproto.ConfigWindowY |
		xproto.ConfigWindowWidth | xproto.ConfigWindowHeight

	xproto.ConfigureWindow(x.conn, win, uint16(mask), vals)
}
