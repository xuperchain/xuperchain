
var assert = require("assert");

Test("naming", function (t) {
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
        resp = contract.Invoke("RegisterChain", {"name":"mainnet.xuper","type":"xuper", "min_endorsor_num":"2"});
	resp2 = contract.Invoke("GetChainMeta", {"name":"mainnet.xuper"})
	console.log(resp2.Body)
	obj = JSON.parse(resp2.Body)
        assert.equal(obj["type"], "xuper") 
	assert.equal(obj["min_endorsor_num"], "2")
    })
})
