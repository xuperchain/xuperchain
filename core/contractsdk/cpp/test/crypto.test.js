var assert = require("assert");

var codePath = "../build/crypto.wasm";

function deploy() {
    return xchain.Deploy({
        name: "crypto",
        code: codePath,
        lang: "c",
        init_args: {}
    });
}

Test("sha256", function (t) {
    var c = deploy();
    var resp = c.Invoke("sha256", { "in": "hello" });
    assert.equal(resp.Body, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824");
})

Test("ecverify", function (t) {
    var c = deploy();
    var resp = c.Invoke("ecverify", {
        "hash": "9956888f6e56184a39fea3a326f90e6ef0e6f7562b8a28b3639b625f06c60d93",
        "sign": "3045022042037faff106c939edeb1d6ce13995c12053d2e546c2421c0d93b74f8fcc0d71022100869bd54d50b3b6156f8b826a522c3475386e416dfb1cee96a92e4292a208dc28",
        "pubkey": "{\"Curvname\":\"P-256\",\"X\":22906815976227871521975920646394124690640348668649121566560042772744241642538,\"Y\":108155450333976935776175836418819691128727398689901106691260413632387093305373}",
    });
    assert.equal(resp.Status, 200, resp.Message);
})