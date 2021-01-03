var assert = require("assert");

var codePath="../award_manage/target/award_manage-0.1.0-jar-with-dependencies.jar"
var lang="java"
var type="native"

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "award_manage",
        code: codePath,
        lang: lang,
        type: type,
        init_args: {"totalSupply":totalSupply},
        options:{"account":"xchain"}
    });
}

function beforeTest(){
    c =  deploy("1000")
var resp = c.Invoke("transfer",{"to":"user1","token":"200"},{"account":"xchain"})
    assert.equal(resp.Body,"ok~")
    return c
}

function  AddAward(t) {
    var c = beforeTest()
var resp = c.Invoke("addAward",{"amount":"200"},{"account":"user1"})
    assert.equal(resp.Message,"you do not have permission to call this method")
var resp = c.Invoke("addAward",{"amount":"0"},{"account":"xchain"})
    assert.equal(resp.Message,"amount must be greater than 0")
var resp = c.Invoke("addAward",{"amount":"200"},{"account":"xchain"});
    assert.equal(resp.Body, "1200");
var resp = c.Invoke("totalSupply",{})
    assert.equal(resp.Body,"1200")
}



function Balance(t){
    var c = beforeTest()
var resp = c.Invoke("balance",{},{"account":"xchain"})
    assert.equal(resp.Body,"800")
var resp = c.Invoke("balance",{},{"account":"user1"})
    assert.equal(resp.Body,"200")
}

function Transfer(){
    c = beforeTest()
var resp = c.Invoke("transfer",{"to":"user2","token":"100"},{"account":"user1"})
    console.log(resp.Message)
    assert.equal(resp.Body,"ok~")
    
var resp = c.Invoke("transfer",{"to":"user2","token":"5000"},{"account":"user1"})
    assert.equal(resp.Message,"balance not enough")

var resp = c.Invoke("transfer",{"to":"user1","token":"100"},{"account":"user1"})
    assert.equal(resp.Message,"can not transfer to yourself")
}

function TransferFrom(t){
    c = beforeTest()

    {
var resp = c.Invoke("transferFrom",{"from":"xchain","token":"200"},{"account":"user2"})
        assert.equal(resp.Status,500)
    }
var resp = c.Invoke("approve",{"to":"user2","token":"200"},{"account":"xchain"})
    assert.equal(resp.Body,"ok~")

var resp = c.Invoke("transferFrom",{"from":"xchain","token":"100"},{"account":"user2"})
    assert.equal(resp.Body,"ok~")

var resp = c.Invoke("transferFrom",{"from":"xchain","token":"300"},{"account":"user2"})
    assert.equal(resp.Message,"allowance balance not enough")
}


Test("AddAward",AddAward)
Test("Balance",Balance)
Test("Transfer",Transfer)
Test("TransferFrom",TransferFrom)
