package golik

import (
	"strconv"
	"sort"
)

func Pool(name string, size int, f func() *Clove) *Clove {
	return &Clove{
		Name: name,
		Receive: func(ctx CloveContext) func(msg Message) {
			return func(msg Message) {
				children := make([]*CloveRef, len(ctx.Children()))
				copy(children, ctx.Children())
				sort.Slice(children, func(i, j int) bool {
					return children[i].Length() < children[j].Length()
				})

				if len(children) > 0 {
					children[0].Forward(msg)
				}
			}
		},
		Async: false,
		PostStart: func(ctx CloveContext) {
			for i := 0; i < size; i++ {
				c := f()
				c.Name = name + "-" + strconv.Itoa(i)
				ctx.Debug("Create worker: %v", i)
				ctx.Run(c)
			}
		},
	}
}
