package golik

import (
	"sort"
	"strconv"
)

func Pool(name string, size int, f func(CloveContext) *Clove) *Clove {
	return &Clove{
		Name: name,
		Sync: true,
		Behavior: func(ctx CloveContext, msg Message) {
			children := make([]CloveRef, len(ctx.Children()))
			children = append(children, ctx.Children()...)
			sort.Slice(children, func(i, j int) bool {
				return children[i].Length() < children[j].Length()
			})

			if len(children) > 0 {
				children[0].Forward(msg)
			}
		},
		PostStart: func(ctx CloveContext) error {
			for i := 0; i < size; i++ {
				c := f(ctx)
				c.Name = name + "-" + strconv.Itoa(i)
				if _, err := ctx.Execute(c); err != nil {
					return err
				}
			}
			return nil
		},
	}
}