//SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.19;

abstract contract ERC20 {
    string public name = "";
    uint8 public decimals = 0;
    string public symbol = "";

    function totalSupply() public view virtual returns (uint256);

    function balanceOf(
        address tokenOwner
    ) public view virtual returns (uint256 balance);

    function allowance(
        address tokenOwner,
        address spender
    ) public view virtual returns (uint256 remaining);

    function transfer(
        address to,
        uint256 tokens
    ) public virtual returns (bool success);

    function approve(
        address spender,
        uint256 tokens
    ) public virtual returns (bool success);

    function transferFrom(
        address from,
        address to,
        uint256 tokens
    ) public virtual returns (bool success);

    function deposit() public payable virtual;

    function withdraw(uint256 tokens) public payable virtual;

    event Transfer(address indexed from, address indexed to, uint256 tokens);
    event Approval(
        address indexed tokenOwner,
        address indexed spender,
        uint256 tokens
    );
    event Deposit(address indexed to, uint256 tokens);
    event Withdrawal(address indexed from, uint256 tokens);
}
