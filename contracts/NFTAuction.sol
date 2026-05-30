// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";

contract NFTAuction is ERC721Holder {
    struct Auction {
        address seller;
        address nftContract;
        uint256 tokenId;
        uint256 startPrice;
        uint256 startTime;
        uint256 duration;
        address highestBidder;
        uint256 highestBid;
        bool ended;
    }

    uint256 public auctionCount;
    mapping(uint256 => Auction) public auctions;

    event AuctionCreated(uint256 auctionId, address seller, address nft, uint256 tokenId);
    event Bid(uint256 auctionId, address bidder, uint256 amount);
    event AuctionEnded(uint256 auctionId, address winner, uint256 bid);

    function createAuction(
        address _nftContract,
        uint256 _tokenId,
        uint256 _startPrice,
        uint256 _duration
    ) external {
        IERC721(_nftContract).transferFrom(msg.sender, address(this), _tokenId);

        auctionCount++;
        auctions[auctionCount] = Auction({
            seller: msg.sender,
            nftContract: _nftContract,
            tokenId: _tokenId,
            startPrice: _startPrice,
            startTime: block.timestamp,
            duration: _duration,
            highestBidder: address(0),
            highestBid: 0,
            ended: false
        });

        emit AuctionCreated(auctionCount, msg.sender, _nftContract, _tokenId);
    }

    function bid(uint256 _auctionId) external payable {
        Auction storage auction = auctions[_auctionId];
        require(!auction.ended, "auction ended");
        require(block.timestamp >= auction.startTime, "not started");
        require(block.timestamp < auction.startTime + auction.duration, "timeout");
        require(msg.value > auction.highestBid, "bid too low");
        require(msg.value >= auction.startPrice, "below start price");

        if (auction.highestBidder != address(0)) {
            payable(auction.highestBidder).transfer(auction.highestBid);
        }

        auction.highestBidder = msg.sender;
        auction.highestBid = msg.value;
        emit Bid(_auctionId, msg.sender, msg.value);
    }

    function endAuction(uint256 _auctionId) external {
        Auction storage auction = auctions[_auctionId];
        require(!auction.ended, "already ended");
        require(block.timestamp >= auction.startTime + auction.duration, "not ended yet");

        auction.ended = true;

        if (auction.highestBidder != address(0)) {
            payable(auction.seller).transfer(auction.highestBid);
            IERC721(auction.nftContract).transferFrom(address(this), auction.highestBidder, auction.tokenId);
        } else {
            IERC721(auction.nftContract).transferFrom(address(this), auction.seller, auction.tokenId);
        }

        emit AuctionEnded(_auctionId, auction.highestBidder, auction.highestBid);
    }

    function getAuction(uint256 id) external view returns (Auction memory) {
        return auctions[id];
    }
}