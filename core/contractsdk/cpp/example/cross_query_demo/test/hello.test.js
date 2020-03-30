
var assert = require("assert");

Test("cross_query_demo", function (t) {
    var contract;
    t.Run("deploy", function (tt) {
        contract = xchain.Deploy({
            name: "cross_query_demo",
            code: "../cross_query_demo.wasm",
            lang: "c",
            init_args: {}
        })
    });
})

