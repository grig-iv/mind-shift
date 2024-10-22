package main

type geometry struct {
	x, y          int
	width, height int
}

func (geom geometry) copy() geometry {
	return geometry{
		x:      geom.x,
		y:      geom.y,
		width:  geom.width,
		height: geom.height,
	}
}

func (geom geometry) shrink(n int) geometry {
	return geometry{
		x:      geom.x + n,
		y:      geom.y + n,
		width:  geom.width - n*2,
		height: geom.height - n*2,
	}
}

func (geom geometry) shrinkRight(n int) geometry {
	return geometry{
		x:      geom.x + n,
		y:      geom.y,
		width:  geom.width - n,
		height: geom.height,
	}
}

func (geom geometry) shrinkLeft(n int) geometry {
	return geometry{
		x:      geom.x,
		y:      geom.y,
		width:  geom.width - n,
		height: geom.height,
	}
}

func (geom geometry) withX(newX int) geometry {
	return geometry{
		x:      newX,
		y:      geom.y,
		width:  geom.width,
		height: geom.height,
	}
}

func (geom geometry) withY(newY int) geometry {
	return geometry{
		x:      geom.x,
		y:      newY,
		width:  geom.width,
		height: geom.height,
	}
}

func (geom geometry) withWidth(newWidth int) geometry {
	return geometry{
		x:      geom.x,
		y:      geom.y,
		width:  newWidth,
		height: geom.height,
	}
}

func (geom geometry) withHeight(newHeight int) geometry {
	return geometry{
		x:      geom.x,
		y:      geom.y,
		width:  geom.width,
		height: newHeight,
	}
}
