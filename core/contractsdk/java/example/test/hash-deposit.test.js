var assert = require("assert");

var codePath = "../hash-deposit/target/hash-despoit-0.1.0-jar-with-dependencies.jar"
var lang = "java"
var type = "native"

function deploy() {
    return xchain.Deploy({
        name: "hash_deposit",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "admin": "xchain" },
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

    {
        var resp = c.Invoke("storeFileInfo", { "user_id": "xchain1", "hash_id": "hash_id1", "file_name": "filename1" });
        assert.equal(resp.Message, "hashid hash_id1 already exists")
    }
    var resp = c.Invoke("queryUserList", {})
    assert.equal(resp.Body, "xchain1\txchain2\txchain3\t")

    var resp = c.Invoke("queryFileInfoByUser", { "user_id": "xchain1" })
    assert.equal(resp.Body, "xchain1\thash_id1\tfilename\n")

    var resp = c.Invoke("queryFileInfoByHash", { "hash_id": "hash_id1" })
    assert.equal(resp.Body, "xchain1\thash_id1\tfilename")
})