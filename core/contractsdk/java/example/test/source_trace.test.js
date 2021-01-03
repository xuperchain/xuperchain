var assert = require("assert");

var codePath="../source_trace/target/source_trace-0.1.0-jar-with-dependencies.jar"
var lang="java"
var type="native"

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "source_trace",
        code: codePath,
        lang: lang,
        type:type,
        init_args: { "admin": "xchain" },
    });
}

function CreateGoods() {
    c = deploy()
var resp = c.Invoke("createGoods", { "id": "id1", "desc": "goods1" })
    assert.equal(resp.Message, "only the admin can create new goods")
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
    assert.equal(resp.Body, "goodsId=id1,updateRecord=0,reason=CREATE")
var resp = c.Invoke("updateGoods",{"id":"id1","reason":"reason0"})
    assert.equal(resp.Message, "missing caller")
var resp = c.Invoke("updateGoods",{"id":"id1","reason":"reason0"},{"account":"xchain"})
    assert.equal(resp.Body,"1")
    {
var resp = c.Invoke("updateGoods", { "id": "id1", "reason": "reason1" })
var resp = c.Invoke("updateGoods", { "id": "id1", "reason": "reason2" })
var resp = c.Invoke("updateGoods", { "id": "id1", "reason": "reason3" })
var resp = c.Invoke("updateGoods", { "id": "id1", "reason": "reason4" })
var resp = c.Invoke("updateGoods", { "id": "id1", "reason": "reason5" })
    }
var resp = c.Invoke("queryRecords",{"id":"id1"})
    assert.equal(resp.Body, "goodsId=id1,updateRecord=0,reason=CREATEgoodsId=id1,updateRecord=1,reason=reason0")
}

Test("CreateGoods", CreateGoods)