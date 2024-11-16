package x

import (
	"log"
	"strings"
	"sync"

	"github.com/grig-iv/mind-shift/domain"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

var (
	Conn   *xgb.Conn
	Setup  *xproto.SetupInfo
	Screen *xproto.ScreenInfo
	Root   xproto.Window

	EventCh = make(chan xgb.Event)
	ErrorCh = make(chan xgb.Error)

	atomMu    sync.Mutex
	atomCache = make(map[atomName]xproto.Atom)
)

func Initialize() {
	var err error = nil

	Conn, err = xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	Setup = xproto.Setup(Conn)
	Screen = Setup.DefaultScreen(Conn)
	Root = Screen.Root
}

func Loop() {
	for {
		ev, xerr := Conn.WaitForEvent()
		if ev == nil && xerr == nil {
			close(EventCh)
			close(ErrorCh)
			log.Println("Both event and error are nil. Exiting...")
			return
		}

		if xerr != nil {
			ErrorCh <- xerr
		}

		if ev != nil {
			EventCh <- ev
		}
	}
}

func InstanceAndClass(win xproto.Window) (string, string) {
	wmClass, _ := xproto.GetProperty(
		Conn,
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

func CheckOtherWm() {
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
		Conn,
		Root,
		xproto.CwEventMask,
		values,
	).Check()

	if err != nil {
		log.Fatal("Another window manager might aleady be running:", err)
	}
}

func DeleteProperty(window xproto.Window, atomName atomName) {
	atom, err := Atom(atomName)
	if err != nil {
		xproto.DeleteProperty(Conn, window, atom)
	}
}

func Geometry(window xproto.Window) (domain.Geometry, error) {
	drw := xproto.Drawable(window)
	geom, err := xproto.GetGeometry(Conn, drw).Reply()
	if err != nil {
		return domain.Geometry{}, err
	}

	return domain.Geometry{
		X:      int(geom.X),
		Y:      int(geom.Y),
		Width:  int(geom.Width),
		Height: int(geom.Height),
	}, nil
}

func ChangeGeometry(win xproto.Window, geom domain.Geometry) {
	geom = domain.Geometry{
		X:      max(geom.X, 0),
		Y:      max(geom.Y, 0),
		Width:  max(geom.Width, 1),
		Height: max(geom.Height, 1),
	}

	vals := []uint32{
		uint32(geom.X),
		uint32(geom.Y),
		uint32(geom.Width),
		uint32(geom.Height),
	}

	mask := xproto.ConfigWindowX | xproto.ConfigWindowY |
		xproto.ConfigWindowWidth | xproto.ConfigWindowHeight

	xproto.ConfigureWindow(Conn, win, uint16(mask), vals)
}
