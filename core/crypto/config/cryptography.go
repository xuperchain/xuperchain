package config

// 定义创建账户时产生的助记词中的标记符的值，及其所对应的椭圆曲线密码学算法的类型
const (
	// 不同语言标准不一样，也许这里用const直接定义值还是好一些
	_ = iota
	// NIST
	Nist // = 1
	// 国密
	Gm // = 2
	// P-256 + schnorr
	NistSN
)

// 定义创建账户时产生的助记词中的标记符的值，及其所对应的预留标记位的类型
const (
	// 不同语言标准不一样，也许这里用const直接定义值还是好一些
	_ = iota
	// 预留标记位的类型1
	ReservedType1
	// 预留标记位的类型2
	ReservedType2
)

// 定义公私钥中所包含的标记符的值，及其所对应的椭圆曲线密码学算法的类型
const (
	// 美国Federal Information Processing Standards的椭圆曲线
	CurveNist = "P-256"
	// 国密椭圆曲线
	CurveGm = "SM2-P-256"
	// Nist P256 + schnorr
	CurveNistSN = "P-256-SN"
)

// IsValidCryptoType 判断是否支持的加密类型
func IsValidCryptoType(ctype byte) bool {
	valid := true
	switch ctype {
	case Nist:
	case Gm:
	case NistSN:
	default:
		valid = false
	}
	return valid
}
