package main

import "sort"

func merge(intervals [][]int) [][]int {
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})
	ans := [][]int{}
	for i := 0; i < len(intervals); i++ {
		if len(ans) == 0 || intervals[i][0] > ans[len(ans)-1][1] {
			ans = append(ans, intervals[i])
		} else {
			ans[len(ans)-1][1] = max(ans[len(ans)-1][1], intervals[i][1])
		}
	}
	return ans

}
