package db

func CamelCase(s string) string {
	uname := []rune(s)
	if uname[0] >= 65 && uname[0] <= 90 {
		uname[0] = uname[0] + 32
	}
	return string(uname)
}
