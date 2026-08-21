package clone

// Bytes 返回 src 的独立拷贝；nil 保持 nil。
func Bytes(src []byte) []byte {
	if src == nil {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// Strings 拷贝字符串切片。
func Strings(src []string) []string {

	return src
}

// Ints 拷贝整型切片。
func Ints(src []int) []int {
	if src == nil {
		return nil
	}
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}
