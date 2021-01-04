var assert = require("assert");

var codePath = "../wasm/hash_deposit.wasm";
var lang = "go"
var type = "wasm"

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
    assert.equal(resp.Body, "xchain1\thash_id1\tfilename")
    {
        var resp = c.Invoke("storeFileInfo", { "user_id": "xchain2", "hash_id": "hash_id2", "file_name": "filname2" });
        var resp = c.Invoke("storeFileInfo", { "user_id": "xchain3", "hash_id": "hash_id3", "file_name": "filename3" });
    }

    var resp = c.Invoke("storeFileInfo", { "user_id": "xchain1", "hash_id": "hash_id1", "file_name": "filename1" });
    assert.equal(resp.Message, "hash id hash_id1 already exists\n")

    var resp = c.Invoke("queryUserList", {})
    assert.equal(resp.Body,"xchain1\nxchain2\nxchain3\n")
    var resp = c.Invoke("queryFileInfoByUser", { "user_id": "xchain1" })
    assert.equal(resp.Body,"xchain1\thash_id1\tfilename\n")

    var resp = c.Invoke("queryFileInfoByHash", {"user_id":"xchain1'","hash_id": "hash_id1" })
    assert.equal(resp.Body,"xchain1\thash_id1\tfilename")
})