var assert = require("assert");

var codePath = "../wasm/source_trace.wasm";
var lang ="go"
var type="wasm"

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
    resp = c.Invoke("CreateGoods", { "id": "id1", "desc": "goods1" })
    assert.equal(resp.Message, "missing caller")
    resp = c.Invoke("CreateGoods", { "id": "id1", "desc": "goods1" }, { "account": "xchain" })
    console.log(resp.Message)
    assert.equal(resp.Body, "id1")
    resp = c.Invoke("CreateGoods", { "id": "id2", "desc": "goods2" }, { "account": "xchain" })
    assert.equal(resp.Body, "id2")
    {
        resp = c.Invoke("CreateGoods", { "id": "id2", "desc": "goods2" }, { "account": "xchain" })
        assert.equal(resp.Message, "goods type id2 aleready exists")
    }
    resp = c.Invoke("QueryRecords", { "id": "id1" })
    assert.equal(resp.Status, 200)
    assert.equal(resp.Body, "goodsId=id1,updateRecord=0,reason=CREATE")
    resp = c.Invoke("UpdateGoods",{"id":"id1","reason":"reason0"})
    assert.equal(resp.Message, "missing caller")
    resp = c.Invoke("UpdateGoods",{"id":"id1","reason":"reason0"},{"account":"xchain"})
    assert.equal(resp.Body,"1")
    {
        c.Invoke("UpdateGoods", { "id": "id1", "reason": "reason1" })
        c.Invoke("UpdateGoods", { "id": "id1", "reason": "reason2" })
        c.Invoke("UpdateGoods", { "id": "id1", "reason": "reason3" })
        c.Invoke("UpdateGoods", { "id": "id1", "reason": "reason4" })
        c.Invoke("UpdateGoods", { "id": "id1", "reason": "reason5" })
    }
    resp = c.Invoke("QueryRecords",{"id":"id1"})
    assert.equal(resp.Body, "goodsId=id1,updateRecord=0,reason=CREATEgoodsId=id1,updateRecord=1,reason=reason0")
}

Test("CreateGoods", CreateGoods)