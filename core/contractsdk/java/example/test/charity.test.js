var assert = require("assert");

var codePath = "../charity/target/charity-0.1.0-jar-with-dependencies.jar"
var lang = "java"
var type = "native"
function deploy(totalSupply) {
    return xchain.Deploy({
        name: "charity",
        code: codePath,
        lang: lang,
        type: type,
        init_args: { "admin": "xchain" },
    });
}

function Donate(t) {
    c = deploy()
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "1000", "timestamp": "1609590581" }, { "account": "unknown" })
    assert.equal(resp.Message, "you do not have permission to call this method")
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "1000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    assert.equal(resp.Message, "")
}

function beforetest() {
    c = deploy()
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    var resp = c.Invoke("donate", { "donor": "donor1", "amount": "10000", "timestamp": "1609590581", "comments": "comments1" }, { "account": "xchain" })
    return c
}
function Cost(t) {
    c = beforetest()
    var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" })
    assert.equal(resp.Message, "missing caller")
    var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "bitcoin" })
    assert.equal(resp.Message, "you do not have permission to call this method")
    {
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        var resp = c.Invoke("cost", { "to": "to", "amount": "100", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
        assert.equal(resp.Status, 200)
        assert.equal(resp.Body, "00000000000000000010")
    }

    var resp = c.Invoke("cost", { "to": "to", "amount": "1000000", "timestamp": "1609590581", "comments": "comments" }, { "account": "xchain" })
    assert.equal(resp.Message, "balance not enough")
    return c
}
function Statistics(t) {
    c = Cost(t)
    var resp = c.Invoke("statistics", {})
    assert.equal(resp.Body, "totalDonates=10000,totalCost=400,fundBalance=9600")
}


function QueryDonor(t) {
    c = Cost(t)
    var resp = c.Invoke("queryDonor", { "donor": "donor1" })
    console.log(resp.Message)
    assert.equal(resp.Body, "total donate count:1\nid=000000000000000000001,donor=donor1,amount=10000,timestamp=1609590581,commnets=comments1\n")
}

function QueryDonates(t) {
    c = Cost(t)
    var resp = c.Invoke("queryDonates", { "start": "00000000000000000005", "limit": "10" })
    console.log(resp.Body)
    console.log(resp.Message)
    assert.equal(resp.Body, "id=00000000000000000005,donor=donor1,amount=10000,timestamp=1609590581,commnets=comments1") // TODO 
}

function QueryCosts(t) {
    c = Cost(t)
    var resp = c.Invoke("queryCosts", { "start": "00000000000000000001", "limit": "10" })
    console.log(resp.Message)
    assert.equal(resp.Body, "id=00000000000000000001to=to,amount=100,timestamp=1609590581,comments=comments\n")
    var resp = c.Invoke("queryCosts", { "start": "00000000000000000001", "limit": "10000" })
    assert.equal(resp.Message, "limit exceeded")

}

Test("Donate", Donate)
Test("Cost", Cost)
Test("Statistics", Statistics)
Test("QueryDonor", QueryDonor)
Test("QueryDonates", QueryDonates)
Test("QueryCosts", QueryCosts)