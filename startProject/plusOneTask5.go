package main

func plusOne(digits []int) []int {
	for i := len(digits) - 1; i >= 0; i-- {
		if i == 0 && digits[i] == 9 {
			digits[i] = 0
			digits = append([]int{1}, digits...)
			return digits
		}
		if digits[i] == 9 {
			digits[i] = 0
		} else {
			digits[i]++
			return digits
		}
	}
	return digits
}
