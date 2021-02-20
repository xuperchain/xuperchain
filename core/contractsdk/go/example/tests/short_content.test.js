var assert = require("assert");

var codePath = "../wasm/short_content.wasm";

var lang = "go"
var type = "wasm"


function deploy() {
    return xchain.Deploy({
        name: "short_content",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "owner": "xchain" }
    });
}

function StoreShortContent() {
    var c = deploy()
    resp = c.Invoke("StoreShortContent", { "user_id": "user1", "topic": "topic1", "title": "title1" })
    assert.equal(resp.Status >= 500, true)
    // assert.equal(resp.Message, "missing content")
    resp = c.Invoke("StoreShortContent", {
        "user_id": "user1",
        "topic": "topic1",
        "title": "title1",
        "content": "content1"
    })
    assert.equal(resp.Body, "ok")
}

function beforeTest() {
    var c = deploy()
    users = ["user1", "user2"]
    topics = ["topic1", "topic2"]
    titles = ["title1", "title2"]
    for (var i = 0; i < 2; i++) {
        for (var j = 0; j < 2; j++) {
            for (var k = 0; k < 2; k++) {
                resp = c.Invoke("StoreShortContent", {
                    "user_id": users[i],
                    "topic": topics[j],
                    "title": titles[k],
                    "content": "content" + i + j + k
                })
            }
        }
    }
    return c
}

function QueryByUser() {
    c = beforeTest()
    resp = c.Invoke("QueryByUser", { "user_id": "user1" })
    console.log(resp.Body)
    console.log(resp.Message)
    assert.equal(resp.Status, 200)
}

function QueryByTitle() {
    var c = beforeTest()
    resp = c.Invoke("QueryByUser", { "user_id": "user1", "topic": "topic1", "title": "title1" })
    assert.equal(resp.Status, 200)
}

function QueryByTopic() {
    var c = beforeTest()
    resp = c.Invoke("QueryByTopic", { "user_id": "user1", "topic": "topic1" })
    console.log(resp.Body)
    assert.equal(resp.Status, 200)
}

Test("StoreShortContent", StoreShortContent)
Test("QueryByUser", QueryByUser)
Test("QueryByTopic", QueryByTopic)
Test("QueryByTitle", QueryByTitle)