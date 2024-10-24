package main

type tag struct {
	id        uint16
	currLaout layout
}

func newTagFromIndex(index uint, layout layout) tag {
	return tag{1 << index, layout}
}
