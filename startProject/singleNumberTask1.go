package main

func singleNumber(nums []int) int {
	//singleMap := make(map[int]int)
	//// 遍历数组
	//for _, v := range nums {
	//	value, exists := singleMap[v]
	//	if !exists {
	//		singleMap[v] = 1
	//	} else {
	//		singleMap[v] = value + 1
	//	}
	//}
	//for k, v := range singleMap {
	//	if v == 1 {
	//		return k
	//	}
	//}
	//return 0
	ans := 0
	//异或
	for _, num := range nums {
		ans ^= num
	}
	return ans
}
