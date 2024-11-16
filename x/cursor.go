// taken from https://github.com/BurntSushi/xgbutil/blob/0d543e91747d289a63a8698e021fdef1512fe766/xcursor/xcursor.go
package x

import "github.com/jezek/xgb/xproto"

const (
	LeftPtrCursor = 68
)

func CreateCursor(cursor uint16) (xproto.Cursor, error) {
	return CreateCursorExtra(cursor, 0, 0, 0, 0xffff, 0xffff, 0xffff)
}

func CreateCursorExtra(cursor, foreRed, foreGreen,
	foreBlue, backRed, backGreen, backBlue uint16) (xproto.Cursor, error) {

	fontId, err := xproto.NewFontId(Conn)
	if err != nil {
		return 0, err
	}

	cursorId, err := xproto.NewCursorId(Conn)
	if err != nil {
		return 0, err
	}

	err = xproto.OpenFontChecked(Conn, fontId,
		uint16(len("cursor")), "cursor").Check()
	if err != nil {
		return 0, err
	}

	err = xproto.CreateGlyphCursorChecked(Conn, cursorId, fontId, fontId,
		cursor, cursor+1,
		foreRed, foreGreen, foreBlue,
		backRed, backGreen, backBlue).Check()
	if err != nil {
		return 0, err
	}

	err = xproto.CloseFontChecked(Conn, fontId).Check()
	if err != nil {
		return 0, err
	}

	return cursorId, nil
}
