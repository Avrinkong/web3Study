const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("MyNFT", function () {
  let MyNFT;
  let myNFT;
  let owner;
  let addr1;
  let addr2;

  beforeEach(async function () {
    [owner, addr1, addr2] = await ethers.getSigners();
    
    MyNFT = await ethers.getContractFactory("MyNFT");
    myNFT = await MyNFT.deploy();
    await myNFT.waitForDeployment();
  });

  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      expect(await myNFT.owner()).to.equal(owner.address);
    });

    it("Should have correct name and symbol", async function () {
      expect(await myNFT.name()).to.equal("MyNFT");
      expect(await myNFT.symbol()).to.equal("MNFT");
    });
  });

  describe("Minting", function () {
    it("Should allow owner to mint NFTs", async function () {
      const tokenURI = "ipfs://test-uri-1";
      
      await expect(myNFT.safeMint(addr1.address, tokenURI))
        .to.emit(myNFT, "Transfer")
        .withArgs(ethers.ZeroAddress, addr1.address, 0);
      
      expect(await myNFT.ownerOf(0)).to.equal(addr1.address);
      expect(await myNFT.tokenURI(0)).to.equal(tokenURI);
    });

    it("Should not allow non-owner to mint", async function () {
      const tokenURI = "ipfs://test-uri-1";
      
      await expect(
        myNFT.connect(addr1).safeMint(addr1.address, tokenURI)
      ).to.be.revertedWithCustomError(myNFT, "OwnableUnauthorizedAccount");
    });

    it("Should allow batch minting", async function () {
      const recipients = [addr1.address, addr2.address];
      const uris = ["ipfs://test-uri-1", "ipfs://test-uri-2"];
      
      await myNFT.batchMint(recipients, uris);
      
      expect(await myNFT.ownerOf(0)).to.equal(addr1.address);
      expect(await myNFT.ownerOf(1)).to.equal(addr2.address);
      expect(await myNFT.tokenURI(0)).to.equal(uris[0]);
      expect(await myNFT.tokenURI(1)).to.equal(uris[1]);
    });

    it("Should track total supply", async function () {
      expect(await myNFT.totalSupply()).to.equal(0);
      
      await myNFT.safeMint(addr1.address, "ipfs://test-uri-1");
      expect(await myNFT.totalSupply()).to.equal(1);
      
      await myNFT.safeMint(addr2.address, "ipfs://test-uri-2");
      expect(await myNFT.totalSupply()).to.equal(2);
    });
  });
});