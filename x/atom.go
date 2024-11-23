package x

import (
	"fmt"
	"log"

	"github.com/jezek/xgb/xproto"
)

type AtomName string

const (
	// default atoms
	WMTransientName AtomName = "WM_TRANSIENT_FOR"
	WMProtocols     AtomName = "WM_PROTOCOLS"
	WMDelete        AtomName = "WM_DELETE_WINDOW"
	WMState         AtomName = "WM_STATE"
	WMTakeFocus     AtomName = "WM_TAKE_FOCUS"

	// EWMH atoms
	NetSupported          AtomName = "_NET_SUPPORTED"
	NetWMName             AtomName = "_NET_WM_NAME"
	NetWMState            AtomName = "_NET_WM_STATE"
	NetWMCheck            AtomName = "_NET_SUPPORTING_WM_CHECK"
	NetWMFullscreen       AtomName = "_NET_WM_STATE_FULLSCREEN"
	NetActiveWindow       AtomName = "_NET_ACTIVE_WINDOW"
	NetWMWindowType       AtomName = "_NET_WM_WINDOW_TYPE"
	NetWMWindowTypeDialog AtomName = "_NET_WM_WINDOW_TYPE_DIALOG"
	NetClientList         AtomName = "_NET_CLIENT_LIST"
)

func AtomProperty(win xproto.Window, atomName AtomName) (*xproto.GetPropertyReply, error) {
	atom, err := Atom(atomName)
	if err != nil {
		return nil, err
	}

	return xproto.GetProperty(
		Conn,
		false,
		win,
		atom,
		xproto.GetPropertyTypeAny,
		0,
		(1<<32)-1,
	).Reply()
}

func AtomOrNone(atomName AtomName) xproto.Atom {
	atom, err := Atom(atomName)
	if err != nil {
		return xproto.AtomNone
	}
	return atom
}

func Atom(atomName AtomName) (xproto.Atom, error) {
	atom, ok := findInCache(atomName)
	if ok {
		return atom, nil
	}

	reply, err := xproto.InternAtom(
		Conn,
		false,
		uint16(len(atomName)),
		string(atomName),
	).Reply()

	if err != nil {
		return 0, fmt.Errorf("Atom: Error interning atom '%s': %s", atomName, err)
	}

	addToCache(atomName, reply.Atom)

	return reply.Atom, nil
}

func findInCache(atomName AtomName) (xproto.Atom, bool) {
	atomMu.Lock()
	defer atomMu.Unlock()

	atom, ok := atomCache[atomName]

	return atom, ok
}

func addToCache(atomName AtomName, atom xproto.Atom) {
	log.Printf("[x.addToCache] Name: %s, Value: %d", atomName, atom)

	atomMu.Lock()
	defer atomMu.Unlock()

	atomCache[atomName] = atom
}
