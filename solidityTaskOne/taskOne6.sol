//  二分查找 (Binary Search)
// 题目描述：在一个有序数组中查找目标值。
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract TaskOne6 {
    /**
     * @dev 在有序数组中使用二分查找算法查找目标值
     * @param nums 有序数组（升序排列）
     * @param target 要查找的目标值
     * @return index 目标值在数组中的索引，如果不存在则返回-1
     */
    function binarySearch(int256[] memory nums, int256 target) 
        public pure returns (int256) 
    {
        uint256 left = 0;
        uint256 right = nums.length;
        
        while (left < right) {
            uint256 mid = left + (right - left) / 2;
            
            if (nums[mid] == target) {
                return int256(mid);
            } else if (nums[mid] < target) {
                left = mid + 1;
            } else {
                right = mid;
            }
        }
        
        return -1; // 未找到目标值
    }
}