pragma solidity >=0.0.0;
pragma experimental ABIEncoderV2;

contract StorageBasicData {
    bool storedBool;
    uint storedUint;
    address storedAddress;
    bytes32 storedBytes;
    string storedString;

    uint[] arrays;

//    string[] arraysStr;

    constructor() public{
        storedUint = 12345;
    }

    function setUints(uint[] memory uintArrays) public{
        uint len = uintArrays.length;
        uint i = 0;
        for (i=0; i < len;i++){
          arrays.push(uintArrays[i]);
        }
    }

    function getUints() view public returns(uint[] memory) {
        return arrays;
    }

//    function setStrings(string[] memory stringArrays) public{
//        uint len = stringArrays.length;
//        uint i = 0;
//        for (i=0; i < len;i++){
//          arraysStr.push(stringArrays[i]);
//        }
//    }
//
//    function getStrings() view public returns(string[] memory) {
//        return arraysStr;
//    }


    function setBool(bool x) public {
        storedBool = x;
    }

    function getBool() view public returns (bool retBool) {
        return storedBool;
    }

    function setUint(uint x) public payable{
        storedUint = x;
    }

    function getUint() view public returns (uint, uint, uint) {
        return (storedUint,storedUint+1,storedUint+2);
    }

    function setAddress(address x) public {
        storedAddress = x;
    }

    function getAddress() view public returns (address retAddress) {
        return storedAddress;
    }

    function setBytes(bytes32 x) public {
        storedBytes = x;
    }

    function getBytes() view public returns (bytes32 retBytes) {
        return storedBytes;
    }

    function setString(string memory x) public {
        storedString = x;
    }

    function getString() view public returns (string memory retString) {
        return storedString;
    }

    function getCurrentBlockInfo() public view returns(uint, uint) {
        return (block.number, block.timestamp);
    }

    function getHistoryBlockInfo(uint height) public view returns(string memory) {
        return Bytes32ToString(blockhash(height));
    }

    function getOwnerBalance() public view returns(uint){
        address owner = msg.sender;
        return owner.balance;
    }

    function getAccBalance(address acc) public view returns(uint){
        return acc.balance;
    }

    function send(address payable receiver, uint amount) public{
        receiver.transfer(amount);
    }

    function Bytes32ToString(bytes32 bname) public view returns(string memory){
        bytes memory bytesChar = new bytes(bname.length);
        uint charCount = 0;
        for(uint i = 0;i < bname.length; i++){
            bytes1 char = bname[i];

            if(char != 0){
                charCount++;
            }
        }

        bytes memory bytesName = new bytes(charCount);
        for(uint j = 0;j < charCount;j++){
            bytesName[j] = bname[j];
        }

        return string(bytesName);
    }

}
