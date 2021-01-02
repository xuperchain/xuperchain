var codePath = "../wasm/luck_draw.wasm";

function deploy(totalSupply) {
    return xchain.Deploy({
        name: "luck_draw",
        code: codePath,
        lang: "go",
        init_args: { "admin": "xchain" },
    });
}