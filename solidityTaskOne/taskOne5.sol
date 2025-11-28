// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract TaskOne5 {
    /**
     * @dev 合并两个有序数组
     * @param nums1 第一个有序数组
     * @param nums2 第二个有序数组
     * @return merged 合并后的有序数组
     */
    function mergeSortedArrays(int256[] memory nums1, int256[] memory nums2) 
        public pure returns (int256[] memory) 
    {
        uint256 m = nums1.length;
        uint256 n = nums2.length;
        int256[] memory merged = new int256[](m + n);
        
        uint256 i = 0; // nums1的索引
        uint256 j = 0; // nums2的索引
        uint256 k = 0; // merged的索引
        
        // 比较两个数组的元素，将较小的元素放入结果数组
        while (i < m && j < n) {
            if (nums1[i] <= nums2[j]) {
                merged[k] = nums1[i];
                i++;
            } else {
                merged[k] = nums2[j];
                j++;
            }
            k++;
        }
        
        // 将剩余元素复制到结果数组
        while (i < m) {
            merged[k] = nums1[i];
            i++;
            k++;
        }
        
        while (j < n) {
            merged[k] = nums2[j];
            j++;
            k++;
        }
        
        return merged;
    }
}