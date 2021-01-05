var assert = require("assert");

var codePath = "../source-trace/target/source-trace-0.1.0-jar-with-dependencies.jar"
var lang = "java"
var type = "native"

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "source_trace",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "admin": "xchain" },
    });
}

function CreateGoods() {
    c = deploy()
    var resp = c.Invoke("createGoods", { "id": "id1", "desc": "goods1" })
    assert.equal(resp.Message, "missing caller")
    var resp = c.Invoke("createGoods", { "id": "id1", "desc": "goods1" }, { "account": "xchain" })
    console.log(resp.Message)
    assert.equal(resp.Body, "id1")
    var resp = c.Invoke("createGoods", { "id": "id2", "desc": "goods2" }, { "account": "xchain" })
    assert.equal(resp.Body, "id2")
    {
        var resp = c.Invoke("createGoods", { "id": "id2", "desc": "goods2" }, { "account": "xchain" })
        assert.equal(resp.Message, "goods type id2 already exists")
    }
    var resp = c.Invoke("queryRecords", { "id": "id1" })
    console.log(resp.Message)
    assert.equal(resp.Status, 200)
    assert.equal(resp.Body, "updateRecord=0,reason=CREATE\n")
    var resp = c.Invoke("updateGoods", { "id": "id1", "reason": "reason0" })
    assert.equal(resp.Message, "missing caller")
    var resp = c.Invoke("updateGoods", { "id": "id1", "reason": "reason0" }, { "account": "xchain" })
    console.log(resp.Body)
    assert.equal(resp.Body, "1")
    {
        var resp = c.Invoke("updateGoods", { "id": "id1", "reason": "reason1" }, { "account": "xchain" })
    }
    var resp = c.Invoke("queryRecords", { "id": "id1" })
    console.log(resp.Body)
    assert.equal(resp.Body, "updateRecord=0,reason=CREATE\nupdateRecord=1,reason=reason0\nupdateRecord=2,reason=reason1\n")
}

Test("CreateGoods", CreateGoods)