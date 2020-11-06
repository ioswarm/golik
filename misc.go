package golik

import (
	"time"
	"strconv"

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

func (v Values) GetString(key string) (string, bool) {
	if v != nil {
		return v[key], true
	}
	return "",false
}

func (v Values) GetInt64(key string) (int64, bool) {
	if s, ok := v.GetString(key); ok {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			return v, true
		}
	}
	return 0, false
}

func (v Values) GetInt(key string) (int, bool) {
	if i, ok := v.GetInt64(key); ok {
		return int(i), ok
	}
	return 0, false
}

func (v Values) GetInt8(key string) (int8, bool) {
	if i, ok := v.GetInt64(key); ok {
		return int8(i), ok
	}
	return 0, false
}

func (v Values) GetInt16(key string) (int16, bool) {
	if i, ok := v.GetInt64(key); ok {
		return int16(i), ok
	}
	return 0, false
}

func (v Values) GetInt32(key string) (int32, bool) {
	if i, ok := v.GetInt64(key); ok {
		return int32(i), ok
	}
	return 0, false
}


func (v Values) GetUint64(key string) (uint64, bool) {
	if s, ok := v.GetString(key); ok {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			return v, true
		}
	}
	return 0, false
}

func (v Values) GetUint(key string) (uint, bool) {
	if i, ok := v.GetUint64(key); ok {
		return uint(i), ok
	}
	return 0, false
}

func (v Values) GetUint8(key string) (uint8, bool) {
	if i, ok := v.GetUint64(key); ok {
		return uint8(i), ok
	}
	return 0, false
}

func (v Values) GetUint16(key string) (uint16, bool) {
	if i, ok := v.GetUint64(key); ok {
		return uint16(i), ok
	}
	return 0, false
}

func (v Values) GetUint32(key string) (uint32, bool) {
	if i, ok := v.GetUint64(key); ok {
		return uint32(i), ok
	}
	return 0, false
}

func (v Values) GetBool(key string) (bool, bool) {
	if s, ok := v.GetString(key); ok {
		if b, err := strconv.ParseBool(s); err == nil {
			return b, true
		}
	}
	return false, false
}

func (v Values) GetFloat64(key string) (float64, bool) {
	if s, ok := v.GetString(key); ok {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, true
		}
	}
	return 0.0, false
}

func (v Values) GetFloat32(key string) (float32, bool) {
	if s, ok := v.GetString(key); ok {
		if f, err := strconv.ParseFloat(s, 32); err == nil {
			return float32(f), true
		}
	}
	return 0.0, false
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
