package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"example.com/objmeta"
)

func main() {
	addr := flag.String("addr", ":8103", "listen address")
	root := flag.String("root", "./data", "object root")
	web := flag.String("web", "web", "static web dir")
	flag.Parse()
	if err := os.MkdirAll(*root, 0o755); err != nil {
		log.Fatal(err)
	}
	st, err := objmeta.Open(context.Background(), objmeta.Options{Root: *root})
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*web)))
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "time": time.Now().UTC()})
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, st.Stats())
	})
	mux.HandleFunc("/api/list", func(w http.ResponseWriter, r *http.Request) {
		prefix := r.URL.Query().Get("prefix")
		res, err := st.List(r.Context(), objmeta.ListOptions{Prefix: prefix, Limit: 200})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, res)
	})
	mux.HandleFunc("/api/put", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		key := r.URL.Query().Get("key")
		body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		info, err := st.Put(r.Context(), key, strings.NewReader(string(body)), objmeta.PutOptions{ContentType: r.Header.Get("Content-Type")})
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		writeJSON(w, info)
	})
	mux.HandleFunc("/api/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		obj, err := st.Get(r.Context(), key)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		defer obj.Body.Close()
		w.Header().Set("ETag", obj.Info.ETag)
		w.Header().Set("Content-Type", obj.Info.ContentType)
		_, _ = io.Copy(w, obj.Body)
	})
	mux.HandleFunc("/api/multipart/create", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		up, err := st.CreateMultipart(r.Context(), key, objmeta.MultipartOptions{})
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		writeJSON(w, up)
	})
	mux.HandleFunc("/api/multipart/abort", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("uploadId")
		if err := st.AbortMultipart(r.Context(), id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		writeJSON(w, map[string]string{"status": "aborted"})
	})
	mux.HandleFunc("/api/head", func(w http.ResponseWriter, r *http.Request) {
		info, err := st.Head(r.Context(), r.URL.Query().Get("key"))
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		writeJSON(w, info)
	})
	_ = filepath.Walk
	_ = strconv.Itoa
	fmt.Println("objd listening on", *addr, "root", *root)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
