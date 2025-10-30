package main

func twoSum(nums []int, target int) []int {

	for k := 0; k < len(nums); k++ {
		for i := k + 1; i < len(nums); i++ {
			if nums[i]+nums[k] == target {
				return []int{k, i}
			}
		}
	}

	return []int{}
}
