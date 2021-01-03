var assert = require("assert");

var codePath = "../hash_deposit/target/hash_despoit-0.1.0-jar-with-dependencies.jar"
var lang = "java"
var type = "native"

function deploy() {
    return xchain.Deploy({
        name: "hash_deposit",
        code: codePath,
        lang: lang,
        type: type,
        init_args: {},
        options: { "account": "xchain" }
    });
}

Test("HashDeposit", function (t) {
    var c = deploy();
    var resp = c.Invoke("storeFileInfo", { "user_id": "xchain1", "hash_id": "hash_id1", "file_name": "filename" });

    console.log(resp.Body)
    var resp = c.Invoke("storeFileInfo", { "user_id": "xchain2", "hash_id": "hash_id2", "file_name": "filname2" });
    console.log(resp.Body)
    var resp = c.Invoke("storeFileInfo", { "user_id": "xchain3", "hash_id": "hash_id3", "file_name": "filename3" });
    console.log(resp.Body)

    var resp = c.Invoke("storeFileInfo", { "user_id": "xchain1", "hash_id": "hash_id1", "file_name": "filename1" });
    console.log(resp.Message)

    var resp = c.Invoke("queryUserList", {})
    console.log(resp.Body)

    var resp = c.Invoke("queryFileInfoByUser", { "user_id": "xchain1" })
    console.log(resp.Body)

    var resp = c.Invoke("queryFileInfoByHash", { "hash_id": "hash_id1'" })
    console.log(resp.Body)
})