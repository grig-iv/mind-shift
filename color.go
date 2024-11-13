package main

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type color struct {
	r, g, b uint8
}

type colorTable map[color]uint32

var (
	normBorder  = color{80, 80, 80}
	focusBorder = color{20, 90, 160}
)

func createColorTable(conn *xgb.Conn, colormap xproto.Colormap) (map[color]uint32, error) {
	table := make(map[color]uint32)

	colors := []color{normBorder, focusBorder}

	colorToCookie := make(map[color]xproto.AllocColorCookie)
	for _, color := range colors {
		colorToCookie[color] = xproto.AllocColor(
			conn,
			colormap,
			uint16(color.r)*255,
			uint16(color.g)*255,
			uint16(color.b)*255,
		)
	}

	for color, cookie := range colorToCookie {
		replay, err := cookie.Reply()
		if err != nil {
			return nil, err
		}

		table[color] = replay.Pixel
	}

	return table, nil
}
