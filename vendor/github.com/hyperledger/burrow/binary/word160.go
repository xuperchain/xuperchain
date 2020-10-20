package binary

const Word160Length = 20
const Word256Word160Delta = 12

var Zero160 = Word160{}

type Word160 [Word160Length]byte

// Pad a Word160 on the left and embed it in a Word256 (as it is for account addresses in EVM)
func (w Word160) Word256() (word256 Word256) {
	copy(word256[Word256Word160Delta:], w[:])
	return
}

func (w Word160) Bytes() []byte {
	return w[:]
}
