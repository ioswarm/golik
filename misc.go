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

