package domain

type Geometry struct {
	X, Y          int
	Width, Height int
}

func (geom Geometry) Copy() Geometry {
	return Geometry{
		X:      geom.X,
		Y:      geom.Y,
		Width:  geom.Width,
		Height: geom.Height,
	}
}

func (geom Geometry) Shrink(n int) Geometry {
	return Geometry{
		X:      geom.X + n,
		Y:      geom.Y + n,
		Width:  geom.Width - n*2,
		Height: geom.Height - n*2,
	}
}

func (geom Geometry) ShrinkTop(n int) Geometry {
	return Geometry{
		X:      geom.X,
		Y:      geom.Y + n,
		Width:  geom.Width,
		Height: geom.Height - n,
	}
}

func (geom Geometry) WithX(newX int) Geometry {
	return Geometry{
		X:      newX,
		Y:      geom.Y,
		Width:  geom.Width,
		Height: geom.Height,
	}
}

func (geom Geometry) WithY(newY int) Geometry {
	return Geometry{
		X:      geom.X,
		Y:      newY,
		Width:  geom.Width,
		Height: geom.Height,
	}
}

func (geom Geometry) WithWidth(newWidth int) Geometry {
	return Geometry{
		X:      geom.X,
		Y:      geom.Y,
		Width:  newWidth,
		Height: geom.Height,
	}
}

func (geom Geometry) WithHeight(newHeight int) Geometry {
	return Geometry{
		X:      geom.X,
		Y:      geom.Y,
		Width:  geom.Width,
		Height: newHeight,
	}
}
