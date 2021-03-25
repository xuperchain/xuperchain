var assert = require("assert");

var codePath = "../build/game_assets.wasm";
var lang = "c"
var type = "wasm"
function deploy(totalSupply) {
    return xchain.Deploy({
        name: "game_assets",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "admin": "xchain" },
        options: { "account": "xchain" }
    });
}


function AddAsset(t) {
    var c = deploy()
    resp = c.Invoke("addAssetType", {
        "type_id": "type_id1",
        "type_desc": "type_desc1"
    }, { "account": "anonymous" })
    assert.equal(resp.Message, "you do not have permission to call this method")
    resp = c.Invoke("addAssetType", { "type_id": "type_id1", "type_desc": "type_desc1" }, { "account": "xchain" })
    assert.equal(resp.Body, "type_id1")
    resp = c.Invoke("addAssetType", {
        "type_id": "type_id1",
        "type_desc": "type_desc1"
    }, { "account": "xchain" })
    assert.equal(resp.Message, "the type_id is already exist, please check again")
    return c
}
function listAssetType(t) {
    var c = deploy(0)
    c.Invoke("addAssetType", { "type_id": "type_id1", "type_desc": "type_desc1" }, { "account": "xchain" })
    c.Invoke("addAssetType", { "type_id": "type_id2", "type_desc": "type_desc2" }, { "account": "xchain" })
    c.Invoke("addAssetType", { "type_id": "type_id3", "type_desc": "type_desc3" }, { "account": "xchain" })
    c.Invoke("addAssetType", { "type_id": "type_id4", "type_desc": "type_desc4" }, { "account": "xchain" })
    resp = c.Invoke("listAssetType", {})
    console.log(resp.Body)
}

function AssetOperations() {
    var c = deploy()
    resp = c.Invoke("addAssetType", { "type_id": "type_id1", "type_desc": "type_desc1" }, { "account": "xchain" })
    c.Invoke("addAssetType", { "type_id": "type_id2", "type_desc": "type_desc2" }, { "account": "xchain" })
    c.Invoke("addAssetType", { "type_id": "type_id3", "type_desc": "type_desc3" }, { "account": "xchain" })
    c.Invoke("addAssetType", { "type_id": "type_id4", "type_desc": "type_desc4" }, { "account": "xchain" })


    resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id1", "asset_id": "asset_id1" })
    assert.equal(resp.Message, "missing initiator")
    {
        resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id1", "asset_id": "asset_id1" }, { "account": "unknown" })
        assert.equal(resp.Message, "you do not have permission to call this method")

        resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id100", "asset_id": "asset_id1" }, { "account": "xchain" })
        assert.equal(resp.Message, "asset type type_id100 not found")
    }

    resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id1", "asset_id": "asset_id1" }, { "account": "xchain" })
    assert.equal(resp.Body, "asset_id1")

    {
        resp = c.Invoke("newAssetToUser", { "user_id": "user_id2", "type_id": "type_id1", "asset_id": "asset_id1" }, { "account": "xchain" })
        console.log(resp.Body)
        assert.equal(resp.Message, "the asset id is already exist, please check again")
    }
    {
        c.Invoke("newAssetToUser", { "user_id": "user_id2", "type_id": "type_id3", "asset_id": "asset_id2" }, { "account": "xchain" })

        c.Invoke("newAssetToUser", { "user_id": "user_id2", "type_id": "type_id1", "asset_id": "asset_id3" }, { "account": "xchain" })
        c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id1", "asset_id": "asset_id4" }, { "account": "xchain" })
    }

    {
        resp = c.Invoke("newAssetToUser", { "user_id": "user_id2", "type_id": "not_exist", "asset_id": "asset_id9" }, { "account": "xchain" })
        assert.deepEqual(resp.Message, "asset type not_exist not found")
    }



    resp = c.Invoke("tradeAsset", { "to": "user_id2", "asset_id": "asset_id1" }, { "account": "user_id1" })
    assert.equal(resp.Status, 200)
}


Test("AssetOperations", AssetOperations)
