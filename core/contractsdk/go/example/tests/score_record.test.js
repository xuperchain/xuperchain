var assert = require("assert");

var codePath = "../wasm/score_record.wasm";

function deploy() {
    return xchain.Deploy({
        name: "award_manage",
        code: codePath,
        lang: "go",
        init_args: {"owner": "xchain"}
    });
}


function AddScore(t) {
    var c = deploy()
    var resp = c.Invoke("AddScore", {"user_id": "user1"})
    assert.equal(resp.Message, "missing caller")
    var resp = c.Invoke("AddScore", {"user_id": "user1", "data": "data1"}, {"account": "xchain"})
    assert.equal(resp.Body, "user1")
}

function QueryScore(t) {
    var c = deploy()
    resp = c.Invoke("AddScore", {"user_id": "user1", "data": "data1"}, {"account": "xchain"})
    assert.equal(resp.Body, "user1")
    resp = c.Invoke("AddScore", {"user_id": "user2", "data": "data2"}, {"account": "xchain"})
    assert.equal(resp.Body, "user2")

    resp = c.Invoke("AddScore", {"user_id": "user3"})
    assert.equal(resp.Message, "missing caller")


    resp = c.Invoke("AddScore", {"user_id": "user3"},{"account":"xchain"})
    assert.equal(resp.Message, "missing data")


    resp = c.Invoke("QueryScore", {"user_id": "user1"})
    assert.equal(resp.Body, "data1")
}

function QueryOwner(t) {
    var c = deploy()
    var resp = c.Invoke("QueryOwner", {})
    assert.equal(resp.Body, "xchain")
}


Test("QueryOwner",QueryOwner)
Test("QueryScore", QueryScore)
Test("AddScore",AddScore)