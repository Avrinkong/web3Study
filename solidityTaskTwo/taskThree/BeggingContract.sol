// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

contract BeggingContract {
    // 合约所有者，在构造函数中设置
    address public owner;
    
    // 使用 mapping 记录每个地址的捐赠总额[citation:1]
    mapping(address => uint256) public donations;
    
    // 捐赠排行榜数组，用于记录前3名捐赠者地址
    address[3] public topDonors;
    
    // 事件：记录每次捐赠[citation:5]
    event DonationMade(address indexed donor, uint256 amount);
    
    // 修饰符：限制只有合约所有者可以调用[citation:1]
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this");
        _;
    }

    // 构造函数，将部署者设置为合约所有者[citation:1]
    constructor() {
        owner = msg.sender;
    }

    // 核心捐赠函数：接收以太币并更新记录[citation:3]
    function donate() external payable {
        require(msg.value > 0, "Donation must be greater than 0");
        
        // 更新该地址的捐赠总额
        donations[msg.sender] += msg.value;
        
        // 更新捐赠排行榜
        updateTopDonors(msg.sender);
        
        // 触发捐赠事件
        emit DonationMade(msg.sender, msg.value);
    }

    // 备用接收以太币的函数：允许直接转账[citation:3]
    receive() external payable {
        require(msg.value > 0, "Donation must be greater than 0");
        donations[msg.sender] += msg.value;
        updateTopDonors(msg.sender);
        emit DonationMade(msg.sender, msg.value);
    }

    // 提款函数：仅所有者可调用，提取合约全部余额[citation:1]
    function withdraw() external onlyOwner {
        // 获取合约当前余额
        uint256 balance = address(this).balance;
        require(balance > 0, "No funds to withdraw");
        
        // 使用 transfer 发送资金[citation:1]
        payable(owner).transfer(balance);
    }

    // 查询特定地址的捐赠金额（public mapping已自动生成getter，此函数便于特定查询）[citation:1]
    function getDonation(address donor) external view returns (uint256) {
        return donations[donor];
    }

    // 获取当前合约总余额
    function getContractBalance() external view returns (uint256) {
        return address(this).balance;
    }

    // 内部函数：更新捐赠排行榜前3名
    function updateTopDonors(address donor) internal {
        uint256 donorAmount = donations[donor];
        
        // 检查并更新前三名
        for (uint256 i = 0; i < 3; i++) {
            address currentTop = topDonors[i];
            // 如果排行榜当前位置为空，或当前捐赠者金额更高，则插入
            if (currentTop == address(0) || donorAmount > donations[currentTop]) {
                // 将后面的名次后移
                for (uint256 j = 2; j > i; j--) {
                    topDonors[j] = topDonors[j - 1];
                }
                // 插入新的捐赠者
                topDonors[i] = donor;
                break;
            }
        }
    }

    // 获取完整的排行榜信息
    function getTopDonors() external view returns (address[3] memory, uint256[3] memory) {
        uint256[3] memory amounts;
        for (uint256 i = 0; i < 3; i++) {
            if (topDonors[i] != address(0)) {
                amounts[i] = donations[topDonors[i]];
            }
        }
        return (topDonors, amounts);
    }
}