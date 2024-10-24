package main

import (
	"fmt"
	"log"

	"github.com/jezek/xgb/xproto"
)

type atomName string

const (
	WMTransientName atomName = "WM_TRANSIENT_FOR"
	WMProtocols     atomName = "WM_PROTOCOLS"
	WMDelete        atomName = "WM_DELETE_WINDOW"
	WMState         atomName = "WMatomName_STATE"
	WMTakeFocus     atomName = "WM_TAKE_FOCUS"

	NetActiveWindow       atomName = "_NET_ACTIVE_WINDOW"
	NetSupported          atomName = "_NET_SUPPORTED"
	NetWMName             atomName = "_NET_WM_NAME"
	NetWMState            atomName = "_NET_WM_STATE"
	NetWMCheck            atomName = "_NET_SUPPORTING_WM_CHECK"
	NetWMFullscreen       atomName = "_NET_WM_STATE_FULLSCREEN"
	NetWMWindowType       atomName = "_NET_WM_WINDOW_TYPE"
	NetWMWindowTypeDialog atomName = "_NET_WM_WINDOW_TYPE_DIALOG"
	NetClientList         atomName = "_NET_CLIENT_LIST"
)

func (x *xserver) atomProperty(win xproto.Window, atomName atomName) (*xproto.GetPropertyReply, error) {
	atom, err := x.atom(atomName)
	if err != nil {
		return nil, err
	}

	return xproto.GetProperty(
		x.conn,
		false,
		win,
		atom,
		xproto.GetPropertyTypeAny,
		0,
		(1<<32)-1,
	).Reply()
}

func (x *xserver) atom(atomName atomName) (xproto.Atom, error) {
	atom, ok := x.findInCache(atomName)
	if ok {
		return atom, nil
	}

	reply, err := xproto.InternAtom(
		x.conn,
		false,
		uint16(len(atomName)),
		string(atomName),
	).Reply()

	if err != nil {
		return 0, fmt.Errorf("Atom: Error interning atom '%s': %s", atomName, err)
	}

	x.addToCache(atomName, reply.Atom)

	return reply.Atom, nil
}

func (x *xserver) findInCache(atomName atomName) (xproto.Atom, bool) {
	x.atomMu.Lock()
	defer x.atomMu.Unlock()

	atom, ok := x.atomCache[atomName]

	return atom, ok
}

func (x *xserver) addToCache(atomName atomName, atom xproto.Atom) {
	log.Printf("[x.addToCache] Name: %s, Value: %d", atomName, atom)

	x.atomMu.Lock()
	defer x.atomMu.Unlock()

	x.atomCache[atomName] = atom
}
