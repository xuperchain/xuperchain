var assert = require("assert");

var codePath = "../wasm/award_manage.wasm";

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "award_manage",
        code: codePath,
        lang: "go",
        init_args: {"totalSupply":totalSupply},
        options:{"account":"xchain"}
    });
}

function beforeTest(){
    c =  deploy("1000")
    c.Invoke("Transfer",{"from":"xchain","to":"user1","amount":"200"})
    return c
}

function  AddAward(t) {
    var c = beforeTest()
    resp = c.Invoke("AddAward",{"amount":"200"})
    assert.equal(resp.Message,"you do not have permission to call this method")
    var resp = c.Invoke("AddAward",{"amount":"200"},{"account":"xchain"});
    assert.equal(resp.Body, "1200");
}



function Balance(t){
    var c = beforeTest()
    resp = c.Invoke("Balance",{"caller":"xchain"})
    assert.equal(resp.Body,"1000")
    resp = c.Invoke("Balance",{"caller":"user1"})
    assert.equal(resp.Message,"200")
}

function Transfer(){
    c = beforeTest()
    c.Invoke("Transfer",{"from":"addr1","to":"addr2","token":"100"})
    assert.equal(resp.Body,"100")
    c.Invoke(Transfer,{"from":"addr1","to":"addr2","token":"5000"})
    assert.equal(resp.Message,"balance too low")
}

function TransferFrom(){
    c = berofeTest()
    c.Invoke("Appro",{"from":"addr1","to":"addr2"})
    assert()
    resp = c.Invoke("TransferFrom",{"from":"addr1","to":"addr2","amount":"100"})
    assert()

    resp = c.Invoke("Approve",{"from":"addr1","to":"addr2","amount":"200"})
    assert()
    resp = c.Invoke("TransferFrom",{"from":"addr1","to":"addr2","amount":"100"})
    // resp = c.Invoke("Balance",{""})

}
Test("AddAward",AddAward)
Test("Balance",Balance)
Test("Transfer",Transfer)
// Test("TransferFrom",TransferFrom)