package jstest

// Adapter is the interface of underlying logic manager
type Adapter interface {
	// Before all test begin
	OnSetup(r *Runner)
	// When all test in a testfile is done
	OnTeardown(r *Runner)
	// When test case is register
	OnTestCase(r *Runner, test TestCase) TestCase
}

type defaultAdapter struct{}

func (d defaultAdapter) OnSetup(r *Runner)                            {}
func (d defaultAdapter) OnTeardown(r *Runner)                         {}
func (d defaultAdapter) OnTestCase(r *Runner, test TestCase) TestCase { return test }
