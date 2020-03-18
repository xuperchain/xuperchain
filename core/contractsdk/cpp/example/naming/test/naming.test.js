
var assert = require("assert");

Test("hello", function (t) {
    var contract;
    t.Run("deploy", function (tt) {
        contract = xchain.Deploy({
            name: "naming",
            code: "../naming.wasm",
            lang: "c",
            init_args: {}
        })
    });

    t.Run("invoke", function (tt) {
        resp = contract.Invoke("RegisterChain", {});
        assert.equal(resp, "hello world");
    })
})
