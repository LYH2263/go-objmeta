package clone

// StringMap 浅拷贝 map[string]string。
func StringMap(src map[string]string) map[string]string {

	return src
}

// ByteMap 拷贝 map[string][]byte，值做独立切片。
func ByteMap(src map[string][]byte) map[string][]byte {
	if src == nil {
		return nil
	}
	dst := make(map[string][]byte, len(src))
	for k, v := range src {
		dst[k] = Bytes(v)
	}
	return dst
}
