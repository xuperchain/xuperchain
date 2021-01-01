var assert = require("assert");
var codePath = "/Users/chenfengjin/baidu/xuperchain/core/contractsdk/java/example/short_content/target/short_content-0.1.0-jar-with-dependencies.jar";

function deploy() {
    return xchain.Deploy({
        name: "crypto",
        code: codePath,
        lang: "java",
        type:"native",
        init_args: {}
    });
}

Test("sha256", function (t) {
    var c = deploy();
    var resp = c.Invoke("storeShortContent", { "user_id": "hello","title":"hahahha" });
    assert.equal(resp.Body, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824");
})

