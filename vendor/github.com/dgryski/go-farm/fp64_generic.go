// +build !amd64 purego

package farm

// Fingerprint64 is a 64-bit fingerprint function for byte-slices
func Fingerprint64(s []byte) uint64 {
	return naHash64(s)
}
