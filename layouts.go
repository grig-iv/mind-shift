package main

import (
	"os/exec"
	"strings"
)

type kbLayout struct {
	layouts map[uint16]string
}

func newKbLayout() *kbLayout {
	l := &kbLayout{}
	l.layouts = make(map[uint16]string)
	return l
}

func (k *kbLayout) update(prevTagId, newTagId uint16) {
	if prevTagId == newTagId {
		return
	}

	go func() {
		prevLayout, err := exec.Command("xkb-switch").Output()
		if err != nil {
			return
		}

		k.layouts[prevTagId] = strings.TrimSpace(string(prevLayout))

		if newLayout, ok := k.layouts[newTagId]; ok {
			exec.Command("xkb-switch", "-s", newLayout).Output()
		}
	}()
}
