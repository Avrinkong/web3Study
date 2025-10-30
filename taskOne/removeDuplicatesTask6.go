package main

import "sort"

func removeDuplicates(nums []int) int {
	hashmap := make(map[int]int)
	for _, v := range nums {
		hashmap[v]++
	}
	nums = nums[:0]
	for i, _ := range hashmap {
		nums = append(nums, i)
	}
	sort.Ints(nums)
	return len(hashmap)
	//可以考虑双指针
}
