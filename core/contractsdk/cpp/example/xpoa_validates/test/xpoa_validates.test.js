
var assert = require("assert");

Test("xpoa_validates", function (t) {
    var contract;
    t.Run("deploy", function (tt) {
        contract = xchain.Deploy({
            name: "xpoa_validates",
            code: "../xpoa_validates.wasm",
            lang: "c",
            init_args: {}
        })
    });

    t.Run("TestAddValidate", function (tt) {
        resp = contract.Invoke("add_validate", { "address": "address1", "neturl": "neturl1" });
        assert.equal(resp.Status, 200)
        resp = contract.Invoke("add_validate", { "address": "address1", "neturl": "neturl1" });
        assert.equal(resp.Message, "this validate already exists")
    })

    t.Run("TestUpdateValidate", function (tt) {
        resp = contract.Invoke("add_validate", { "address": "address1", "neturl": "neturl1" });
        resp = contract.Invoke("update_validate", { "address": "address1", "neturl": "update neturl1" });
        assert.equal(resp.Status, 200)
    })

    t.Run("TestDelValidate", function (tt) {
        resp = contract.Invoke("add_validate", { "address": "address1", "neturl": "neturl1" });
        resp = contract.Invoke("del_validate", { "address": "address1" });
        assert.equal(resp.Status, 200)
    })

    t.Run("TestGetValidates", function (tt) {
        resp = contract.Invoke("add_validate", { "address": "address1", "neturl": "neturl1" });
        assert.equal(resp.Status, 200)
        resp = contract.Invoke("add_validate", { "address": "address2", "neturl": "neturl2" });
        assert.equal(resp.Status, 200)
        resp = contract.Invoke("add_validate", { "address": "address3", "neturl": "neturl3" });
        assert.equal(resp.Status, 200)
        resp = contract.Invoke("get_validates", {});
        validates = JSON.parse(resp.Body)["proposers"]
        assert.equal(validates[0].address, "address1")
        assert.equal(validates[0].neturl, "neturl1")
        assert.equal(validates[1].address, "address2")
        assert.equal(validates[1].neturl, "neturl2")
        assert.equal(validates[2].address, "address3")
        assert.equal(validates[2].neturl, "neturl3")
    })
})
