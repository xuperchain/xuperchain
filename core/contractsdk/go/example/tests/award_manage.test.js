var assert = require("assert");

var codePath = "../wasm/award_manage.wasm";

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "award_manage",
        code: codePath,
        lang: "go",
        init_args: {"totalSupply":totalSupply},
        options:{"account":"xchain"}
    });
}
Test("AddAward", function (t) {
    var c = deploy("100");
    var resp = c.Invoke("AddAward",{"amount":"200"},{"account":"xchain"});
    assert.equal(resp.Body, "300");
    resp=c.Invoke("TotalSupply",{})
    assert.equal(resp.Body,"300")
    // Balance
    // Allowance
    // Transfer
    // TransferFrom
    // Approve
})