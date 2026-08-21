// Package objmeta 提供本地对象存储的元数据与 multipart 会话层。
//
// 典型用法：
//
//	st, err := objmeta.Open(ctx, objmeta.Options{Root: dir})
//	etag, err := st.Put(ctx, "a/b", bytes.NewReader(body), objmeta.PutOptions{})
//	obj, err := st.Get(ctx, "a/b")
//	up, err := st.CreateMultipart(ctx, "big.bin", objmeta.MultipartOptions{})
//	_, err = st.UploadPart(ctx, up.UploadID, 1, partReader)
//	_, err = st.CompleteMultipart(ctx, up.UploadID, []int{1})
//
// Put/Get 入口会对 body 与返回元数据做拷贝，避免调用方缓冲别名污染。
package objmeta
