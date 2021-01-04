var assert = require("assert");

var codePath = "../game_assets/target/game_assets-0.1.0-jar-with-dependencies.jar"
var lang = "java"
var type = "native"
function deploy(totalSupply) {
    return xchain.Deploy({
        name: "award_manage",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "admin": "xchain" },
        options: { "account": "xchain" }
    });
}


function AddAsset(t) {
    var c = deploy()
    var resp = c.Invoke("addAssetType", { "type_id": "type_id1", "type_desc": "type_desc1" }, { "account": "anonymous" })
    assert.equal(resp.Message, "you do not have permission to call this method")
    var resp = c.Invoke("addAssetType", { "type_id": "type_id1", "type_desc": "type_desc1" }, { "account": "xchain" })
    console.log(resp.Message)
    assert.equal(resp.Body, "type_id1")
    var resp = c.Invoke("addAssetType", { "type_id": "type_id1", "type_desc": "type_desc1" }, { "account": "xchain" })
    assert.equal(resp.Message, "asset type type_id1 already exists")
    return c
}
function ListAssetType(t) {
    var c = deploy(0)
    var resp = c.Invoke("addAssetType", { "type_id": "type_id1", "type_desc": "type_desc1" }, { "account": "xchain" })
    var resp = c.Invoke("addAssetType", { "type_id": "type_id2", "type_desc": "type_desc2" }, { "account": "xchain" })
    var resp = c.Invoke("addAssetType", { "type_id": "type_id3", "type_desc": "type_desc3" }, { "account": "xchain" })
    var resp = c.Invoke("addAssetType", { "type_id": "type_id4", "type_desc": "type_desc4" }, { "account": "xchain" })
    var resp = c.Invoke("listAssetType", {})
    console.log(resp.Body)
}

function AssetOperations() {
    var c = deploy()
    var resp = c.Invoke("addAssetType", { "type_id": "type_id1", "type_desc": "type_desc1" }, { "account": "xchain" })
    // console.log("add type ",resp.Body)
    var resp = c.Invoke("addAssetType", { "type_id": "type_id2", "type_desc": "type_desc2" }, { "account": "xchain" })
    var resp = c.Invoke("addAssetType", { "type_id": "type_id3", "type_desc": "type_desc3" }, { "account": "xchain" })
    var resp = c.Invoke("addAssetType", { "type_id": "type_id4", "type_desc": "type_desc4" }, { "account": "xchain" })


    var resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id1", "asset_id": "asset_id1" })
    assert.equal(resp.Message, "you do not have permission to call this method")
    {
        var resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id1", "asset_id": "asset_id1" }, { "account": "unknown" })
        assert.equal(resp.Message, "you do not have permission to call this method")

        var resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id100", "asset_id": "asset_id1" }, { "account": "xchain" })
        assert.equal(resp.Message, "asset type type_id100 not found")
    }

    var resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id1", "asset_id": "asset_id1" }, { "account": "xchain" })
    assert.equal(resp.Body, "asset_id1")

    {
        var resp = c.Invoke("newAssetToUser", { "user_id": "user_id2", "type_id": "type_id1", "asset_id": "asset_id1" }, { "account": "xchain" })
        console.log(resp.Body)
        assert.equal(resp.Message, "asset asset_id1 exists")
    }
    {
        var resp = c.Invoke("newAssetToUser", { "user_id": "user_id2", "type_id": "type_id3", "asset_id": "asset_id2" }, { "account": "xchain" })

        var resp = c.Invoke("newAssetToUser", { "user_id": "user_id2", "type_id": "type_id1", "asset_id": "asset_id3" }, { "account": "xchain" })
        var resp = c.Invoke("newAssetToUser", { "user_id": "user_id1", "type_id": "type_id1", "asset_id": "asset_id4" }, { "account": "xchain" })
    }

    var resp = c.Invoke("getAssetByUser", { "user_id": "user_id1" }, { "account": "xchain" })
    console.log(resp.Message)
    assert.strictEqual(resp.Body, "assetId=asset_id1,typeId=type_id1,assetDesc=type_desc1\nassetId=asset_id4,typeId=type_id1,assetDesc=type_desc1\n")

    var resp = c.Invoke("getAssetByUser", { "user_id": "user_id2" }, { "account": "xchain" })
    assert.equal(resp.Body, "assetId=asset_id2,typeId=type_id3,assetDesc=type_desc3\nassetId=asset_id3,typeId=type_id1,assetDesc=type_desc1\n")
    {
        var resp = c.Invoke("tradeAsset", { "to": "user_id2", "asset_id": "asset_id2" }, { "account": "user_id1" })
        assert.equal(resp.Message, "asset asset_id2 of user user_id1 not found")
    }

    var resp = c.Invoke("tradeAsset", { "to": "user_id2", "asset_id": "asset_id1" }, { "account": "user_id1" })
    assert.equal(resp.Status, 200)
    console.log(resp.Body)


    var resp = c.Invoke("getAssetByUser", { "user_id": "user_id1" }, { "account": "xchain" })
    console.log("body", resp.Body)
    console.log("message", resp.Message)
    assert.equal(resp.Body, "assetId=asset_id4,typeId=type_id1,assetDesc=type_desc1\n")

    var resp = c.Invoke("getAssetByUser", { "user_id": "user_id2" }, { "account": "xchain" })
    assert.equal(resp.Body, "assetId=asset_id1,typeId=type_id1,assetDesc=type_desc1\nassetId=asset_id2,typeId=type_id3,assetDesc=type_desc3\nassetId=asset_id3,typeId=type_id1,assetDesc=type_desc1\n")
}

Test("AddAsset", AddAsset)
Test("ListAssetType", ListAssetType)
Test("AssetOperations", AssetOperations)
