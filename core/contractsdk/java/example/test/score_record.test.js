var assert = require("assert");

var codePath="../score_record/target/score_record-0.1.0-jar-with-dependencies.jar"

var lang="java"
var type="native"

function deploy() {
    return xchain.Deploy({
        name: "award_manage",
        code: codePath,
        lang: lang,
        type: type,
        init_args: {"owner": "xchain"}
    });
}


function AddScore(t) {
    var c = deploy()
var resp = c.Invoke("addScore", {"user_id": "user1"})
    assert.equal(resp.Message, "missing caller")
var resp = c.Invoke("addScore", {"user_id": "user1", "data": "data1"}, {"account": "xchain"})
    assert.equal(resp.Body, "user1")
}

function QueryScore(t) {
    var c = deploy()
var resp = c.Invoke("addScore", {"user_id": "user1", "data": "data1"}, {"account": "xchain"})
    assert.equal(resp.Body, "user1")
var resp = c.Invoke("addScore", {"user_id": "user2", "data": "data2"}, {"account": "xchain"})
    assert.equal(resp.Body, "user2")

var resp = c.Invoke("addScore", {"user_id": "user3"})
    assert.equal(resp.Message, "missing caller")


var resp = c.Invoke("addScore", {"user_id": "user3"},{"account":"xchain"})
    assert.equal(resp.Message, "missing data")


var resp = c.Invoke("queryScore", {"user_id": "user1"})
    assert.equal(resp.Body, "data1")
}

function QueryOwner(t) {
    var c = deploy()
var resp = c.Invoke("queryOwner", {})
    assert.equal(resp.Body, "xchain")
}


Test("QueryOwner",QueryOwner)
Test("QueryScore", QueryScore)
Test("AddScore",AddScore)