package golik

type Service interface {
	CreateInstance(system Golik) *Clove
}