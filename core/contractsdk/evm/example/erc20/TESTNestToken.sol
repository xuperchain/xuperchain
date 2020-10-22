pragma solidity >=0.0.0;

// https://github.com/ethereum/EIPs/issues/20
interface Token {
    function totalSupply() external view returns (uint supply);
    function balanceOf(address _owner) external view returns (uint balance);
    function transfer(address _to, uint _value) external returns (bool success);
    function transferFrom(address _from, address _to, uint _value) external returns (bool success);
    function approve(address _spender, uint _value) external returns (bool success);
    function allowance(address _owner, address _spender) external view returns (uint remaining);
    function decimals() external view returns(uint digits);
}

contract TESTNestToken {
    Token testToken;

    constructor(Token token) public{
        testToken = token;
    }

    function getTokenAddress() public view returns(Token) {
        return testToken;
    }

    function balanceOf(address _owner) external view returns(uint) {
        return testToken.balanceOf(_owner);
    }

    function transfer(address _to, uint _value) external returns (bool) {
        return testToken.transfer(_to, _value);
    }

    function decimals() external view returns(uint) {
        return testToken.decimals();
    }

    function totalSupply() external view returns (uint) {
        return testToken.totalSupply();
    }
}