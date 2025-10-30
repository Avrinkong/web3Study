package main

func longestCommonPrefix(strs []string) string {
	res := []byte{}
	for i, str := range strs {
		if i == 0 {
			res = []byte(str)
			continue
		}
		b := []byte(str)
		length := len(res)
		if length == 0 {
			return ""
		}
		if length > len(b) {
			res = res[:len(b)]
		}
		for i2, b2 := range b {
			if len(res)-1 < i2 || res[i2] != b2 {
				res = res[:i2]
				break
			}
		}
	}
	return string(res)
}
