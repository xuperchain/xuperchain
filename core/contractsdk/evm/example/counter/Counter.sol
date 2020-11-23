pragma solidity >=0.0.0;

contract Counter {
    address owner;
    mapping (string => uint256) values;

    constructor() public{
        owner = msg.sender;
    }

    function increase(string memory key) public payable{
        values[key] = values[key] + 1;
    }

    function get(string memory key) view public returns (uint) {
        return values[key];
    }

    function getOwner() view public returns (address) {
        return owner;
    }

}
