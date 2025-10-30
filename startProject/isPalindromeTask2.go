package main

import "strconv"

func isPalindrome(x int) bool {
	as := strconv.Itoa(x)
	for i := 0; i < len(as)/2; i++ {
		if as[i] != as[len(as)-i-1] {
			return false
		}
	}
	return true
}
