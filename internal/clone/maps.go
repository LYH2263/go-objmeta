package clone

// StringMap 深拷贝 map[string]string；与调用方 map 隔离，nil 保持 nil。
func StringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
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
