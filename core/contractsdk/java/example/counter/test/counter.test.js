
var assert = require("assert");

Test("countr-demo", function (t) {
    var contract;
    t.Run("deploy", function (tt) {
        contract = xchain.Deploy({
            name: "counter",
            code: "../target/counter-0.1.0-jar-with-dependencies.jar",
            lang: "java",
            type:"native",
            init_args: {}
        })
    });
})