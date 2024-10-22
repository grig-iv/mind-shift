package main

type layout interface {
	arrange(screenGeom geometry, totalClients int) []geometry
}

type masterStack struct {
	screenPadding int
	clientPadding int
	masterRatio   float32
}

func (lt masterStack) arrange(screenGeom geometry, totalClients int) []geometry {
	if totalClients == 0 {
		return []geometry{}
	}

	screenGeom = screenGeom.shrink(lt.screenPadding)

	if totalClients == 1 {
		return []geometry{screenGeom}
	}

	layout := make([]geometry, totalClients)

	masterWidth := int(float32(screenGeom.width-lt.clientPadding) * lt.masterRatio)
	masterGeom := screenGeom.withWidth(masterWidth)

	layout[0] = masterGeom

	stackX := masterGeom.x + masterGeom.width + lt.clientPadding
	stackWidth := screenGeom.width - masterGeom.width - lt.clientPadding
	stackGeom := screenGeom.withX(stackX).withWidth(stackWidth)

	totalStackClients := totalClients - 1
	clientHeight := (screenGeom.height / totalStackClients) - (lt.clientPadding * (totalStackClients - 1))
	for i := range totalStackClients {
		clientY := screenGeom.y + clientHeight*i + lt.clientPadding*i
		clientGeom := stackGeom.withHeight(clientHeight).withY(clientY)
		layout[i+1] = clientGeom
	}

	return layout
}
