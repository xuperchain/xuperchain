var assert = require("assert");

var codePath = "../wasm/award_manage.wasm";
var lang = "go"
var type = "wasm"

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "award_manage",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "total_supply": totalSupply },
        options: { "account": "xchain" }
    });
}

function beforeTest() {
    c = deploy("1000")
    resp = c.Invoke("Transfer", { "to": "user1", "token": "200" }, { "account": "xchain" })
    console.log(resp.Message)
    assert.equal(resp.Body, "ok")
    return c
}

function AddAward(t) {
    var c = beforeTest()
    resp = c.Invoke("AddAward", { "amount": "200" }, { "account": "user1" })
    assert.equal(resp.Message, "you do not have permission to call this method")
    var resp = c.Invoke("AddAward", { "amount": "0" }, { "account": "xchain" })
    assert.equal(resp.Status >= 500, true)
    // assert.equal(resp.Message, "amount must be greater than 0")
    var resp = c.Invoke("AddAward", { "amount": "200" }, { "account": "xchain" });
    assert.equal(resp.Body, "1200");
    resp = c.Invoke("TotalSupply", {})
    assert.equal(resp.Body, "1200")
}



function Balance(t) {
    var c = beforeTest()
    resp = c.Invoke("Balance", {"owner":"xchain"})
    assert.equal(resp.Body, "800")
    resp = c.Invoke("Balance", {"owner":"user1"})
    assert.equal(resp.Body, "200")
}

function Transfer() {
    c = beforeTest()
    resp = c.Invoke("Transfer", { "to": "user2", "token": "100" }, { "account": "user1" })
    console.log(resp.Message)
    assert.equal(resp.Body, "ok")

    resp = c.Invoke("Transfer", { "to": "user2", "token": "5000" }, { "account": "user1" })
    assert.equal(resp.Message, "balance not enough")

    resp = c.Invoke("Transfer", { "to": "user1", "token": "100" }, { "account": "user1" })
    assert.equal(resp.Message, "can not transfer to yourself")
}

function TransferFrom(t) {
    c = beforeTest()

    {
        resp = c.Invoke("TransferFrom", { "from": "xchain", "token": "200" }, { "account": "user2" })
        assert.equal(resp.Status, 500)
    }
    resp = c.Invoke("Approve", { "to": "user2", "token": "200" }, { "account": "xchain" })
    assert.equal(resp.Body, "ok")

    resp = c.Invoke("TransferFrom", { "from": "xchain", "token": "100" }, { "account": "user2" })
    assert.equal(resp.Body, "ok")

    resp = c.Invoke("TransferFrom", { "from": "xchain", "token": "300" }, { "account": "user2" })
    assert.equal(resp.Message, "allowance balance not enough")
}


Test("AddAward", AddAward)
Test("Balance", Balance)
Test("Transfer", Transfer)
Test("TransferFrom", TransferFrom)
