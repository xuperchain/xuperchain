var assert = require("assert");

var codePath = "../luck_draw/target/luck_draw-0.1.0-jar-with-dependencies.jar"
var lang = "java"
var type = "native"

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "luck_draw",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "admin": "xchain" },
    });
}


Test("LuckDraw", function (t) {
    c = deploy()
    var resp = c.Invoke("getLuckId", {}, { "account": "user1" })
    var resp = c.Invoke("getLuckId", {}, { "account": "user2" })
    var resp = c.Invoke("getLuckId", {}, { "account": "user3" })
    var resp = c.Invoke("getLuckId", {}, { "account": "user4" })
    var resp = c.Invoke("getLuckId", {}, { "account": "user5" })
    console.log(resp.Message)
    assert.equal(resp.Body, "5")

    var resp = c.Invoke("getLuckId", {}, { "account": "user1" })
    assert.equal(resp.Body, "1")

    var resp = c.Invoke("startLuckDraw", {}, { "account": "nobody" })
    assert.equal(resp.Message, "you do not have permission to call this method")

    var resp = c.Invoke("startLuckDraw", { "seed": "100" }, { "account": "xchain" })
    assert.equal(resp.Message, "")
    assert.equal(resp.Status, 200)
    var resp = c.Invoke("getResult", {})
    assert.equal(resp.Message, "")
    assert.equal(resp.Status, 200)

    var resp = c.Invoke("getLuckId", {}, { "account": "user5" })
    assert.equal(resp.Message, "the luck draw has finished")
})