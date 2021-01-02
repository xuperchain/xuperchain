var assert = require("assert");

var codePath = "../wasm/source_trace.wasm";

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "source_trace",
        code: codePath,
        lang: "go",
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
    c.Invoke("QueryRecords",{})
    assert.equal()

    c.Invoke("UpdateGods")
    assert.equal()
    c.Invoke("UpdateGods")
    c.Invoke("UpdateGods")
    c.Invoke("UpdateGods")
    c.Invoke("UpdateGods")
    c.Invoke("QueryRecords")
    assert.equal()
}

Test("CreateGoods", CreateGoods)