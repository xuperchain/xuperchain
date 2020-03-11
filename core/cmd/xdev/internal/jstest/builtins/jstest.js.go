package builtins

var testSourceCode = `
function makeTestContext(t) {
	var o = Object.create(t);
	// overwrite Run method in testing.T
	o.Run = function(name, body) {
		return t.Run(name, makeTestBody(body));
	}
	return o;
}

function makeTestBody(body) {
	return function(t) {
		newt = makeTestContext(t);
		try {
			body(newt);
		} catch(err) {
			if (err.stack != undefined) {
				t.Error(err.name + ": " + err.message);
			} else {
				t.Error(err);
			}
		}
	}
}

function Test(name, body) {
	_test(name, makeTestBody(body));
}

module.exports = {
	"Test" : Test,
}
`

func init() {
	RegisterModule("jstest", testSourceCode)
}
