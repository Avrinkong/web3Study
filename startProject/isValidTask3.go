package main

func isValid(s string) bool {
	hashMap := make(map[byte]byte)
	hashMap['('] = ')'
	hashMap['{'] = '}'
	hashMap['['] = ']'

	b := []byte(s)
	stack := []byte{}
	for i := 0; i < len(b); i++ {
		if b[i] == '(' || b[i] == '[' || b[i] == '{' {
			stack = append(stack, hashMap[b[i]])
		} else {
			if len(stack) == 0 {
				return false
			}
			if stack[len(stack)-1] != b[i] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) != 0 {
		return false
	}
	return true
}
