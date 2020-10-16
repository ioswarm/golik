package golik

import (
	"time"

	"encoding/hex"
	"hash/fnv"
)

var (
	title = `
 ________  ________  ___       ___  ___  __       
|\   ____\|\   __  \|\  \     |\  \|\  \|\  \     
\ \  \___|\ \  \|\  \ \  \    \ \  \ \  \/  /|_   
 \ \  \  __\ \  \\\  \ \  \    \ \  \ \   ___  \  
  \ \  \|\  \ \  \\\  \ \  \____\ \  \ \  \\ \  \ 
   \ \_______\ \_______\ \_______\ \__\ \__\\ \__\
    \|_______|\|_______|\|_______|\|__|\|__| \|__|			
`
)

type Values map[string]string

func (v Values) Add(key string, value string) {
	v[key] = value
}

func (v Values) Get(key string) string {
	if v == nil {
		return ""
	}
	return v[key]
}

func hash() string {
	h := fnv.New64a()
	h.Write([]byte(time.Now().UTC().String()))
	return hex.EncodeToString(h.Sum(nil))
}

func CamelCase(s string) string {
	uname := []rune(s)
	if uname[0] >= 65 && uname[0] <= 90 {
		uname[0] = uname[0] + 32
	}
	return string(uname)
}
