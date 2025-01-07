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
	atomCache = make(map[AtomName]xproto.Atom)
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
	eventMask := xproto.EventMaskSubstructureRedirect |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskButtonPress |
		xproto.EventMaskPointerMotion |
		xproto.EventMaskEnterWindow |
		xproto.EventMaskLeaveWindow |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskPropertyChange

	err := xproto.ChangeWindowAttributesChecked(
		Conn,
		Root,
		xproto.CwEventMask,
		[]uint32{uint32(eventMask)},
	).Check()

	if err != nil {
		log.Fatal("Another window manager might aleady be running:", err)
	}
}

func DeleteProperty(window xproto.Window, atomName AtomName) {
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

func ChangeBorderWidth(win xproto.Window, width int) {
	xproto.ConfigureWindow(
		Conn,
		win,
		xproto.ConfigWindowBorderWidth,
		[]uint32{uint32(width)},
	)
}

func ChangeBorderColor(win xproto.Window, pixel uint32) {
	xproto.ChangeWindowAttributes(
		Conn,
		win,
		xproto.CwBorderPixel,
		[]uint32{uint32(pixel)},
	)
}

func Raise(win xproto.Window) {
	xproto.ConfigureWindow(
		Conn,
		win,
		xproto.ConfigWindowStackMode,
		[]uint32{xproto.StackModeAbove},
	)
}

func SetInputFocus(win xproto.Window) {
	xproto.SetInputFocus(
		Conn,
		xproto.InputFocusPointerRoot,
		win,
		xproto.TimeCurrentTime,
	)
}

func ChangeProperty32(
	win xproto.Window,
	property xproto.Atom,
	propertyType xproto.Atom,
	data ...uint) {
	buf := toBuf32(data...)
	xproto.ChangeProperty(
		Conn,
		xproto.PropModeReplace,
		Root,
		property,
		propertyType,
		32,
		uint32(len(buf)/(int(32)/8)),
		buf,
	)
}

func GetTransientFor(win xproto.Window) (xproto.Window, bool) {
	transient, err := xproto.GetProperty(
		Conn,
		false,
		win,
		AtomOrNone(WMTransientName),
		xproto.AtomWindow,
		0,
		1,
	).Reply()

	if err != nil {
		log.Println(err)
		return 0, false
	}

	if transient.ValueLen == 0 || transient.Format != 32 {
		return 0, false
	}

	return xproto.Window(xgb.Get32(transient.Value)), true
}

func toBuf32(data ...uint) []byte {
	buf := make([]byte, len(data)*4)
	for i, datum := range data {
		xgb.Put32(buf[(i*4):], uint32(datum))
	}
	return buf
}

func SendEvent(win xproto.Window, protoName AtomName) bool {
	wmProtocolsAtom := AtomOrNone(WMProtocols)

	proto, err := Atom(protoName)
	if err != nil {
		log.Println(err)
		return false
	}

	reply, err := xproto.GetProperty(
		Conn, false, win,
		wmProtocolsAtom, xproto.AtomAtom,
		0, (1<<32)-1,
	).Reply()

	if err != nil {
		log.Println(err)
		return false
	}

	exists := false
	if reply != nil && reply.Format == 32 {
		for i := 0; i < int(reply.ValueLen); i++ {
			if xproto.Atom(reply.Value[i*4]) == proto {
				exists = true
				break
			}
		}
	}

	if exists {
		ev := xproto.ClientMessageEvent{
			Format: 32,
			Window: win,
			Type:   wmProtocolsAtom,
			Data: xproto.ClientMessageDataUnionData32New([]uint32{
				uint32(proto),
				uint32(xproto.TimeCurrentTime),
				0, 0, 0,
			}),
		}
		err = xproto.SendEventChecked(Conn, false, win, xproto.EventMaskNoEvent, string(ev.Bytes())).Check()
		if err != nil {
			log.Println(err)
			return false
		}
	}

	return exists
}
