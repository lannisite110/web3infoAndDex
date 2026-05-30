// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";

/// @dev Sepolia 演示用 NFT：任何人可 mint，用于拍卖练习
contract TestNFT is ERC721 {
    uint256 private _nextTokenId;

    constructor() ERC721("Test NFT", "TNFT") {}

    function mint(address to) external returns (uint256 tokenId) {
        tokenId = ++_nextTokenId;
        _mint(to, tokenId);
    }
}
