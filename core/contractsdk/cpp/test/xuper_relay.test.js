var assert = require("assert");

var blocks = [
  "CAEaIHKOefSs/0BqwsVoSQGnSeUvFvCz/uKDaSQEkhowdMPWIiBpJ0+SeB+ZFIBg0C4Tal7a4zaO\
aBrqFAH7940MMePHuCohUU1CMVhHWUgydkszZ1QxQld1YUpkS1RtY1ZaOVlCdlY4MkcwRQIhAI3e\
lh7zrSi6hEsfHIPtPG9+DAJymtPh8IslSNWNPxGpAiBmAkIP9WOViMiKQERgIunxxPbVBRyzNpjH\
u8MewJdwHDq5AXsiQ3Vydm5hbWUiOiJQLTI1NiIsIlgiOjIyOTA2ODE1OTc2MjI3ODcxNTIxOTc1\
OTIwNjQ2Mzk0MTI0NjkwNjQwMzQ4NjY4NjQ5MTIxNTY2NTYwMDQyNzcyNzQ0MjQxNjQyNTM4LCJZ\
IjoxMDgxNTU0NTAzMzM5NzY5MzU3NzYxNzU4MzY0MTg4MTk2OTExMjg3MjczOTg2ODk5MDExMDY2\
OTEyNjA0MTM2MzIzODcwOTMzMDUzNzN9QiDWiZxOG4Mg7mmgKeigiOJfqpeKJSPtesRngqcRGh5H\
ZEjobVDbiPGWpsnR/hVgBWog7hPsJRDo4rCgStmo+Ikv4xy1CIejNVMc4aExWYU+vktqIFPB2wnt\
fwynmowaUZ0Zb572EPVBwpXuSmCki3paY2iwaiBwhQ/9WLmpHT6sRDPvVx82EpIH0o1MM1dO/MlD\
boFi62og8WRPL/s8yN744pzPABmR2m/+mVejQwH3GU96CELfdv9qIAxCWwDXkXkPKnshEvO/8IA0\
IGJBgewJMEs43U9901eOagBqAGoAaiDd5SfmqUI9K4FIFrwVZmLH6YJTUzSJIpb+tL0aGqCFb2og\
XP5GIIPH97SuP4JMtNQbaVT5N5ppIV3RDjVxZpGtt9pqIO6VuUODnu3ZIk5q80YqoTL78T5IbYPQ\
C3Ft8eijhvvQagBqIKv3aEti2fIzGW8ucE2GICUlmHf3ytROjwQxBzeHI6K4aiCdeA3t/ZvBbYBL\
/YoPG4R796bWV54qodkj7NDFbi2Yvmog1omcThuDIO5poCnooIjiX6qXiiUj7XrEZ4KnERoeR2Rw\
AXogc5w7Rdr0WrgeE7O7yEf1tvz02GX7puMpDOFmcgggK1A=",

  "CAEaIHOcO0Xa9Fq4HhOzu8hH9bb89Nhl+6bjKQzhZnIIICtQIiByjnn0rP9AasLFaEkBp0nlLxbw\
s/7ig2kkBJIaMHTD1iohUU1CMVhHWUgydkszZ1QxQld1YUpkS1RtY1ZaOVlCdlY4MkcwRQIhAIAO\
znXvROccAWxJ9AtBo7wXx8iry/OBMWhvO9PZ+65SAiBBKen+qQdGnBY34CB2J4vOQt4I4b5T3tFY\
xo381EzVMzq5AXsiQ3Vydm5hbWUiOiJQLTI1NiIsIlgiOjIyOTA2ODE1OTc2MjI3ODcxNTIxOTc1\
OTIwNjQ2Mzk0MTI0NjkwNjQwMzQ4NjY4NjQ5MTIxNTY2NTYwMDQyNzcyNzQ0MjQxNjQyNTM4LCJZ\
IjoxMDgxNTU0NTAzMzM5NzY5MzU3NzYxNzU4MzY0MTg4MTk2OTExMjg3MjczOTg2ODk5MDExMDY2\
OTEyNjA0MTM2MzIzODcwOTMzMDUzNzN9QiDonjAlbWtSxL4B438zW3BoGuAREvsJYekq9HmMum58\
90jpbVDfpfyMs8nR/hVgAmogMl+cISmdb555ughPhELOOg0Rtxou0h4fyG7bv+S3O89qIBHqBitQ\
9pKD6k2Ay2fdsHIw8DQTGHm1Gp0it5oKbfMfaiDonjAlbWtSxL4B438zW3BoGuAREvsJYekq9HmM\
um5893ABeiD7KPjgz6O/frrzrY8Yu7thNHOsRysqKkhriwN24QTJoA==",

  "CAEaIPso+ODPo79+uvOtjxi7u2E0c6xHKyoqSGuLA3bhBMmgIiBznDtF2vRauB4Ts7vIR/W2/PTY\
Zfum4ykM4WZyCCArUCohUU1CMVhHWUgydkszZ1QxQld1YUpkS1RtY1ZaOVlCdlY4MkgwRgIhAL9T\
GoY1UOV4b777BYeRB4WNtBWGpIX7mqhxHwHTvvpUAiEAti9l4OA1J/VOOWwwbBMF1gxjH13zOkGp\
kkXAs5ZDQsY6uQF7IkN1cnZuYW1lIjoiUC0yNTYiLCJYIjoyMjkwNjgxNTk3NjIyNzg3MTUyMTk3\
NTkyMDY0NjM5NDEyNDY5MDY0MDM0ODY2ODY0OTEyMTU2NjU2MDA0Mjc3Mjc0NDI0MTY0MjUzOCwi\
WSI6MTA4MTU1NDUwMzMzOTc2OTM1Nzc2MTc1ODM2NDE4ODE5NjkxMTI4NzI3Mzk4Njg5OTAxMTA2\
NjkxMjYwNDEzNjMyMzg3MDkzMzA1MzczfUIg/jGvqbMO1g6mcf2gPuytdcdJ8ejw9T7Oe/30bj+t\
LOJI6m1Q/+iYjb3J0f4VYAFqIP4xr6mzDtYOpnH9oD7srXXHSfHo8PU+znv99G4/rSzicAF6ILyu\
mCyJNk5UpKhcDDWJwYWhqHAo+kT1Sz+7J32YhHcx",

  "CAEaILyumCyJNk5UpKhcDDWJwYWhqHAo+kT1Sz+7J32YhHcxIiD7KPjgz6O/frrzrY8Yu7thNHOs\
RysqKkhriwN24QTJoCohUU1CMVhHWUgydkszZ1QxQld1YUpkS1RtY1ZaOVlCdlY4MkcwRQIgBAub\
MPkkmNXrIJU57TZMITBad1duvil2afQiD42h6hECIQD8j1Z+P+EGDThwRvAHmrR5mWNEISQqY88Q\
/n+lm3s6UTq5AXsiQ3Vydm5hbWUiOiJQLTI1NiIsIlgiOjIyOTA2ODE1OTc2MjI3ODcxNTIxOTc1\
OTIwNjQ2Mzk0MTI0NjkwNjQwMzQ4NjY4NjQ5MTIxNTY2NTYwMDQyNzcyNzQ0MjQxNjQyNTM4LCJZ\
IjoxMDgxNTU0NTAzMzM5NzY5MzU3NzYxNzU4MzY0MTg4MTk2OTExMjg3MjczOTg2ODk5MDExMDY2\
OTEyNjA0MTM2MzIzODcwOTMzMDUzNzN9QiCGSSMIJxT/fmb/v463En6nlFDL1rUCaG5mRCgm9sT5\
j0jrbVCh8r3QyMnR/hVgAWoghkkjCCcU/35m/7+OtxJ+p5RQy9a1AmhuZkQoJvbE+Y9wAXogsIM+\
DLQE3FLTfY3EXQGm8M3auYZTtZxifDo7tlDgG0M=",
];

/*
{
  "version": 1,
  "blockid": "728e79f4acff406ac2c5684901a749e52f16f0b3fee283692404921a3074c3d6",
  "preHash": "69274f92781f99148060d02e136a5edae3368e681aea1401fbf78d0c31e3c7b8",
  "proposer": "QMB1XGYH2vK3gT1BWuaJdKTmcVZ9YBvV8",
  "sign": "30450221008dde961ef3ad28ba844b1f1c83ed3c6f7e0c02729ad3e1f08b2548d58d3f11a902206602420ff5639588c88a40446022e9f1c4f6d5051cb33698c7bbc31ec097701c",
  "pubkey": "{\"Curvname\":\"P-256\",\"X\":22906815976227871521975920646394124690640348668649121566560042772744241642538,\"Y\":108155450333976935776175836418819691128727398689901106691260413632387093305373}",
  "merkleRoot": "d6899c4e1b8320ee69a029e8a088e25faa978a2523ed7ac46782a7111a1e4764",
  "height": 14056,
  "timestamp": 1584499929204409435,
  "transactions": [],
  "txCount": 5,
  "merkleTree": [
    "ee13ec2510e8e2b0a04ad9a8f8892fe31cb50887a335531ce1a13159853ebe4b",
    "53c1db09ed7f0ca79a8c1a519d196f9ef610f541c295ee4a60a48b7a5a6368b0",
    "70850ffd58b9a91d3eac4433ef571f36129207d28d4c33574efcc9436e8162eb",
    "f1644f2ffb3cc8def8e29ccf001991da6ffe9957a34301f7194f7a0842df76ff",
    "0c425b00d791790f2a7b2112f3bff0803420624181ec09304b38dd4f7dd3578e",
    "",
    "",
    "",
    "dde527e6a9423d2b814816bc156662c7e982535334892296feb4bd1a1aa0856f",
    "5cfe462083c7f7b4ae3f824cb4d41b6954f9379a69215dd10e35716691adb7da",
    "ee95b943839eedd9224e6af3462aa132fbf13e486d83d00b716df1e8a386fbd0",
    "",
    "abf7684b62d9f233196f2e704d862025259877f7cad44e8f043107378723a2b8",
    "9d780dedfd9bc16d804bfd8a0f1b847bf7a6d6579e2aa1d923ecd0c56e2d98be",
    "d6899c4e1b8320ee69a029e8a088e25faa978a2523ed7ac46782a7111a1e4764"
  ],
  "inTrunk": true,
  "nextHash": "739c3b45daf45ab81e13b3bbc847f5b6fcf4d865fba6e3290ce1667208202b50",
  "failedTxs": null,
  "curTerm": 0,
  "curBlockNum": 0,
  "justify": {}
}
*/
var block0 = blocks[0];

/*
{
  "txid": "ee13ec2510e8e2b0a04ad9a8f8892fe31cb50887a335531ce1a13159853ebe4b",
  "blockid": "728e79f4acff406ac2c5684901a749e52f16f0b3fee283692404921a3074c3d6",
  "txInputs": [
    {
      "refTxid": "fc06bfd7869464e04acd399b00b38038b090264b4d8120ef15e9be490111f43c",
      "refOffset": 0,
      "fromAddr": "QMB1XGYH2vK3gT1BWuaJdKTmcVZ9YBvV8",
      "amount": "428100000000"
    }
  ],
  "txOutputs": [
    {
      "amount": "100",
      "toAddr": "QMB1XGYH2vK3gT1BWuaJdKTmcVZ9YBvV8"
    },
    {
      "amount": "428099999900",
      "toAddr": "QMB1XGYH2vK3gT1BWuaJdKTmcVZ9YBvV8"
    }
  ],
  "desc": "transfer from console",
  "nonce": "158449992798498081",
  "timestamp": 1584499927945161071,
  "version": 1,
  "autogen": false,
  "coinbase": false,
  "txInputsExt": null,
  "txOutputsExt": null,
  "contractRequests": null,
  "initiator": "QMB1XGYH2vK3gT1BWuaJdKTmcVZ9YBvV8",
  "authRequire": [
    "QMB1XGYH2vK3gT1BWuaJdKTmcVZ9YBvV8"
  ],
  "initiatorSigns": [
    {
      "publickey": "{\"Curvname\":\"P-256\",\"X\":22906815976227871521975920646394124690640348668649121566560042772744241642538,\"Y\":10815545$
333976935776175836418819691128727398689901106691260413632387093305373}",
      "sign": "3045022042037faff106c939edeb1d6ce13995c12053d2e546c2421c0d93b74f8fcc0d71022100869bd54d50b3b6156f8b826a522c3475386e416dfb1cee96a$
2e4292a208dc28"
    }
  ],
  "authRequireSigns": [
    {
      "publickey": "{\"Curvname\":\"P-256\",\"X\":22906815976227871521975920646394124690640348668649121566560042772744241642538,\"Y\":108155450
333976935776175836418819691128727398689901106691260413632387093305373}",
      "sign": "3045022039d90659c07357d5378b319ae88073345f63955c028c5f5d3fe4828466e715fb022100ed3caf74eedd5773ea46883336b64063ba5a5b7a74b55cd9a5
bd84a0ee4491a2"
    }
  ],
  "receivedTimestamp": 1584499927962283030,
  "modifyBlock": {
    "marked": false,
    "effectiveHeight": 0,
    "effectiveTxid": ""
  }
}
*/
var tx = "CiDuE+wlEOjisKBK2aj4iS/jHLUIh6M1UxzhoTFZhT6+SxIgco559Kz/QGrCxWhJAadJ5S8W8LP+\
4oNpJASSGjB0w9YaTAog/Aa/14aUZOBKzTmbALOAOLCQJktNgSDvFem+SQER9DwqIVFNQjFYR1lI\
MnZLM2dUMUJXdWFKZEtUbWNWWjlZQnZWODIFY6y/mQAiJgoBZBIhUU1CMVhHWUgydkszZ1QxQld1\
YUpkS1RtY1ZaOVlCdlY4IioKBWOsv5icEiFRTUIxWEdZSDJ2SzNnVDFCV3VhSmRLVG1jVlo5WUJ2\
VjgyFXRyYW5zZmVyIGZyb20gY29uc29sZUISMTU4NDQ5OTkyNzk4NDk4MDgxSO/Str6hydH+FVAB\
0gEhUU1CMVhHWUgydkszZ1QxQld1YUpkS1RtY1ZaOVlCdlY42gEhUU1CMVhHWUgydkszZ1QxQld1\
YUpkS1RtY1ZaOVlCdlY44gGFAgq5AXsiQ3Vydm5hbWUiOiJQLTI1NiIsIlgiOjIyOTA2ODE1OTc2\
MjI3ODcxNTIxOTc1OTIwNjQ2Mzk0MTI0NjkwNjQwMzQ4NjY4NjQ5MTIxNTY2NTYwMDQyNzcyNzQ0\
MjQxNjQyNTM4LCJZIjoxMDgxNTU0NTAzMzM5NzY5MzU3NzYxNzU4MzY0MTg4MTk2OTExMjg3Mjcz\
OTg2ODk5MDExMDY2OTEyNjA0MTM2MzIzODcwOTMzMDUzNzN9EkcwRQIgQgN/r/EGyTnt6x1s4TmV\
wSBT0uVGwkIcDZO3T4/MDXECIQCGm9VNULO2FW+LgmpSLDR1OG5Bbfsc7papLkKSogjcKOoBhQIK\
uQF7IkN1cnZuYW1lIjoiUC0yNTYiLCJYIjoyMjkwNjgxNTk3NjIyNzg3MTUyMTk3NTkyMDY0NjM5\
NDEyNDY5MDY0MDM0ODY2ODY0OTEyMTU2NjU2MDA0Mjc3Mjc0NDI0MTY0MjUzOCwiWSI6MTA4MTU1\
NDUwMzMzOTc2OTM1Nzc2MTc1ODM2NDE4ODE5NjkxMTI4NzI3Mzk4Njg5OTAxMTA2NjkxMjYwNDEz\
NjMyMzg3MDkzMzA1MzczfRJHMEUCIDnZBlnAc1fVN4sxmuiAczRfY5VcAoxfXT/kgoRm5xX7AiEA\
7TyvdO7dV3PqRogzNrZAY7paW3p0tVzZpb2EoO5EkaLwAZbYy8ahydH+FQ==";

var relayCodePath = "../build/xuper_relayer.wasm";
var crossCodePath = "../build/cross_chain.wasm";
function deployRelay() {
  return xchain.Deploy({
    name: "relayer",
    code: relayCodePath,
    lang: "c",
    init_args: {}
  });
}

function deployCross() {
  return xchain.Deploy({
    name: "cross_chain",
    code: crossCodePath,
    lang: "c",
    init_args: {}
  });
}

Test("ProofTx", function (t) {
  var c = deployRelay();
  t.Run("put_blocks", function (tt) {
    resp = c.Invoke("initAnchorBlockHeader", {
      "blockHeader": atob(block0),
    })
    assert.equal(resp.Status, 200, resp.Message)
    for (i = 1; i < blocks.length; i++) {
      resp = c.Invoke("putBlockHeader", {
        "blockHeader": atob(blocks[i]),
      })
      assert.equal(resp.Status, 200, resp.Message)
    }
  })

  t.Run("tx_exists", function (tt) {
    var paths = [
      "53c1db09ed7f0ca79a8c1a519d196f9ef610f541c295ee4a60a48b7a5a6368b0",
      "5cfe462083c7f7b4ae3f824cb4d41b6954f9379a69215dd10e35716691adb7da",
      "9d780dedfd9bc16d804bfd8a0f1b847bf7a6d6579e2aa1d923ecd0c56e2d98be"
    ];
    resp = c.Invoke("verifyTx", {
      // blockid of block0
      "blockid": "728e79f4acff406ac2c5684901a749e52f16f0b3fee283692404921a3074c3d6",
      "txid": "ee13ec2510e8e2b0a04ad9a8f8892fe31cb50887a335531ce1a13159853ebe4b",
      "txIndex": "0",
      "proofPath": paths.join(","),
    })
    assert.equal(resp.Status, 200, resp.Message)
  })
});

Test("VerifyTx", function (t) {
  var relay = deployRelay();
  resp = relay.Invoke("initAnchorBlockHeader", {
    "blockHeader": atob(block0),
  })
  assert.equal(resp.Status, 200, resp.Message)
  for (i = 1; i < blocks.length; i++) {
    resp = relay.Invoke("putBlockHeader", {
      "blockHeader": atob(blocks[i]),
    })
    assert.equal(resp.Status, 200, resp.Message)
  }

  var cross = deployCross();
  var paths = [
    "53c1db09ed7f0ca79a8c1a519d196f9ef610f541c295ee4a60a48b7a5a6368b0",
    "5cfe462083c7f7b4ae3f824cb4d41b6954f9379a69215dd10e35716691adb7da",
    "9d780dedfd9bc16d804bfd8a0f1b847bf7a6d6579e2aa1d923ecd0c56e2d98be"
  ];
  resp = cross.Invoke("verifyTx", {
    // relay contract name
    "relay": relay.Name,

    // 下面的是对tx是否存在的验证
    "blockid": "728e79f4acff406ac2c5684901a749e52f16f0b3fee283692404921a3074c3d6",
    "txIndex": "0",
    "proofPath": paths.join(","),
    "tx": atob(tx),

    // 是否包含转账给自己的记录
    "amount": "100",
  }, {
    "account": "QMB1XGYH2vK3gT1BWuaJdKTmcVZ9YBvV8",
  })
  assert.equal(resp.Status, 200, resp.Message)
})