var assert = require("assert");

var codePath = "../wasm/charity.wasm";
var lang = "go"
var type = "wasm"
function deploy(totalSupply) {
    return xchain.Deploy({
        name: "award_manage",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "admin": "xchain" },
    });
}

function Donate(t) {
    c = deploy()
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "1000", "timestamp": "1609590581" }, { "account": "unknown" })
    assert.equal(resp.Message, "you do not have permission to call this method")
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "1000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    assert.equal(resp.Message, "")
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "-100", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    assert.equal(resp.Status, 500)
}

function beforetest() {
    c = deploy()
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    resp = c.Invoke("Donate", { "donor": "donor2", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    return c
}
function Cost(t) {
    c = beforetest()
    resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" })
    assert.equal(resp.Message, "missing initiator")
    resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "bitcoin" })
    assert.equal(resp.Message, "you do not have permission to call this method")
    {
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        resp = c.Invoke("Cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        assert.equal(resp.Status, 200)
        assert.equal(resp.Body, "00000000000000000010")
    }

    resp = c.Invoke("Cost", { "to": "to", "amount": "1000000", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
    assert.equal(resp.Message, "balance not enough")
    resp = c.Invoke("Cost", { "donor": "donor1", "amount": "-100", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    assert.equal(resp.Status, 500)
    return c
}
function Statistics(t) {
    c = Cost(t)
    resp = c.Invoke("Statistics", {})
    assert.deepStrictEqual(JSON.parse(resp.Body), { "total_donate": "210000", "total_cost": "1000", "fund_balance": "209000" })
}


function QueryDonor(t) {
    c = Cost(t)
    resp = c.Invoke("QueryDonor", { "donor": "donor2" })
    console.log(resp.Body)
    console.log(resp.Status, 200)
    console.log(resp.Message)
    assert.deepStrictEqual(JSON.parse(resp.Body), [
        {
            "id": "00000000000000000021",
            "donor": "donor2",
            "amount": "10000",
            "timestamp": "1609590581",
            "comments": "comments1"
        }
    ])
}

function QueryDonates(t) {
    c = Cost(t)
    resp = c.Invoke("QueryDonates", { "start": "00000000000000000005", "limit": "3" })
    console.log(resp.Body)
    assert.deepStrictEqual(JSON.parse(resp.Body), [{
        "id": "00000000000000000005", "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1"
    }, {
        "id": "00000000000000000006", "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1"
    }, {
        "id": "00000000000000000007", "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1"
    }])
}

function QueryCosts(t) {
    c = Cost(t)
    resp = c.Invoke("QueryCosts", { "start": "00000000000000000001", "limit": "3" })
    console.log(resp.Body)
    assert.deepStrictEqual(JSON.parse(resp.Body), [{
        "id": "00000000000000000001", "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments"
    }, {
        "id": "00000000000000000002", "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments"
    },
    {
        "id": "00000000000000000003", "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments"
    }])
    resp = c.Invoke("QueryCosts", { "start": "00000000000000000001", "limit": "10000" })
    assert.equal(resp.Message, "limit exceeded")
}

Test("Donate", Donate)
Test("Cost", Cost)
Test("Statistics", Statistics)
Test("QueryDonor", QueryDonor)
Test("QueryDonates", QueryDonates)
Test("QueryCosts", QueryCosts)