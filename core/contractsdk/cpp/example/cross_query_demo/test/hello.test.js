
var assert = require("assert");

Test("hello", function (t) {
    var contract;
    t.Run("deploy", function (tt) {
        contract = xchain.Deploy({
            name: "hello",
            code: "../hello.wasm",
            lang: "c",
            init_args: {}
        })
    });

    t.Run("invoke", function (tt) {
        resp = contract.Invoke("hello", {});
        assert.equal(resp.Body, "hello world");
    })
})
