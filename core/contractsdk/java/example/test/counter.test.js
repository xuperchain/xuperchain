var assert = require("assert");

var codePath="../counter/target/counter-0.1.0-jar-with-dependencies.jar"

var lang="java"
var type="native"
function deploy() {
    return xchain.Deploy({
        name: "counter",
        code: codePath,
        lang: lang,
        type:type,
        init_args: {"creator":"xchain"}
    });
}

Test("Increase", function (t) {
    var c = deploy();
var resp = c.Invoke("increase",{"key":"xchain"},{"name":"11111"});
    assert.equal(resp.Body, "1");
})

Test("Get",function (t) {
    var c = deploy()
var resp = c.Invoke("increase",{"key":"xchain"});
var resp = c.Invoke("get",{"key":"xchain"})
    assert.equal(resp.Body,"1")
})