var assert = require("assert");

var codePath = "../short_content/target/short_content-0.1.0-jar-with-dependencies.jar"

var lang = "java"
var type = "native"


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
    var resp = c.Invoke("storeShortContent", { "user_id": "user1", "topic": "topic1", "title": "title1" })
    assert.equal(resp.Message, "missing user_id or title of topic or content")
    var resp = c.Invoke("storeShortContent", { "user_id": "user1", "topic": "topic1", "title": "title1", "content": "content1" })
    assert.equal(resp.Body, "ok~")
}

function beforeTest() {
    var c = deploy()
    users = ["user1", "user2"]
    topics = ["topic1", "topic2"]
    titles = ["title1", "title2"]
    for (var i = 0; i < 2; i++) {
        for (var j = 0; j < 2; j++) {
            for (var k = 0; k < 2; k++) {
                var resp = c.Invoke("storeShortContent", { "user_id": users[i], "topic": topics[j], "title": titles[k], "content": "content" + i + j + k })
            }
        }
    }
    return c
}
function QueryByUser() {
    c = beforeTest()
    var resp = c.Invoke("queryByUser", { "user_id": "user1" })
    console.log(resp.Body, resp.Message)
}

function QueryByTitle() {
    var c = beforeTest()
    var resp = c.Invoke("queryByUser", { "user_id": "user1", "topic": "topic1", "title": "title1" })
    console.log(resp.Body, resp.Message)
}

function QueryByTopic() {
    var c = beforeTest()
    var resp = c.Invoke("queryByUser", { "user_id": "user1", "topic": "topic1" })
    console.log(resp.Body, resp.Message)
}

Test("StoreShortContent", StoreShortContent)
Test("QueryByUser", QueryByUser)
Test("QueryByTopic", QueryByTopic)
Test("QueryByTitle", QueryByTitle)