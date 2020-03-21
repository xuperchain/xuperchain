
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

    t.Run("Register", function (tt) {
        resp = contract.Invoke("RegisterChain", {"name":"mainnet.xuper","type":"xuper", "min_endorsor_num":"2"});
        resp2 = contract.Invoke("Resolve", {"name":"mainnet.xuper"})
        console.log(resp2.Body)
        meta = JSON.parse(resp2.Body)["chain_meta"]
        assert.equal(meta["type"], "xuper") 
        assert.equal(meta["min_endorsor_num"], 2)
        r3 = contract.Invoke("UpdateChain", {"name":"mainnet.xuper", "type":"xuper", "min_endorsor_num":"aaaa"})
        assert.equal(r3.Message, "invalid min_endorsor_num, it should be greater than 0")
    })

    t.Run("AddEndorsor", function(tt) {
        r0 = contract.Invoke("AddEndorsor", {"name":"testnet", "address":"foo", "host":"127.0.0.1", "pub_key":"xxx"})
        assert.equal(r0.Message, "chain name does not exist")
        r1 = contract.Invoke("AddEndorsor", {"name":"mainnet.xuper", "address":"bobfffff", "host":"192.168.8.8:37101", "pub_key":"xxxxx"})
        console.log(r1.Body)
        r2 = contract.Invoke("AddEndorsor", {"name":"mainnet.xuper", "address":"bobfffff", "host":"192.168.9.9:37101", "pub_key":"xxxxx"})
        assert.equal(r2.Message, "endorsor already exists")
        r3 = contract.Invoke("AddEndorsor", {"name":"mainnet.xuper", "address":"alicefff", "host":"192.168.9.9:37101", "pub_key":"xxxxx"})
        resp = contract.Invoke("Resolve", {"name":"mainnet.xuper"})
        console.log(resp.Body)
        nodes = JSON.parse(resp.Body)["endorsors"]
        assert.equal(nodes[0].address, "alicefff")
        assert.equal(nodes[1].address, "bobfffff")
    })
})
