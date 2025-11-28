// 反转一个字符串。输入 "abcde"，输出 "edcba"

// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

contract ReverseString {
    function reverse(string memory str) public pure returns (string memory) {
        bytes memory s = bytes(str);
        bytes memory reversed = new bytes(s.length);
        for (uint i = 0; i < s.length; i++) {
            reversed[i] = s[s.length - 1 - i];
        }
        return string(reversed);
    }
}
 
