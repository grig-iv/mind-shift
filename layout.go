package main

import "github.com/grig-iv/mind-shift/domain"

type layout interface {
	arrange(screenGeom domain.Geometry, totalClients int) []domain.Geometry
}

type masterStack struct {
	screenPadding int
	clientPadding int
	masterRatio   float32
}

func (lt masterStack) arrange(screenGeom domain.Geometry, totalClients int) []domain.Geometry {
	if totalClients == 0 {
		return []domain.Geometry{}
	}

	screenGeom = screenGeom.Shrink(lt.screenPadding)

	if totalClients == 1 {
		return []domain.Geometry{screenGeom}
	}

	layout := make([]domain.Geometry, totalClients)

	masterWidth := int(float32(screenGeom.Width-lt.clientPadding) * lt.masterRatio)
	masterGeom := screenGeom.WithWidth(masterWidth)

	layout[0] = masterGeom

	stackX := masterGeom.X + masterGeom.Width + lt.clientPadding
	stackWidth := screenGeom.Width - masterGeom.Width - lt.clientPadding
	stackGeom := screenGeom.WithX(stackX).WithWidth(stackWidth)

	totalStackClients := totalClients - 1
	clientHeight := (screenGeom.Height / totalStackClients) - (lt.clientPadding * (totalStackClients - 1))
	for i := range totalStackClients {
		clientY := screenGeom.Y + clientHeight*i + lt.clientPadding*i
		clientGeom := stackGeom.WithHeight(clientHeight).WithY(clientY)
		layout[i+1] = clientGeom
	}

	return layout
}
