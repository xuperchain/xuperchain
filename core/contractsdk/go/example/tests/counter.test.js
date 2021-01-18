var assert = require("assert");

var codePath = "../wasm/counter.wasm";

var lang = "go"
var type = "wasm"
function deploy() {
    return xchain.Deploy({
        name: "counter",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "creator": "xchain" }
    });
}

Test("Increase", function (t) {
    var c = deploy();
    var resp = c.Invoke("Increase", { "key": "xchain" }, { "name": "11111" });
    assert.equal(resp.Body, "1");
})

Test("Get", function (t) {
    var c = deploy()
    c.Invoke("Increase", { "key": "xchain" });
    var resp = c.Invoke("Get", { "key": "xchain" })
    assert.equal(resp.Body, "1")
})