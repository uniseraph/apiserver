package utils

import (
	"net/http"
	"sync"
)

// A WriteFlusher provides synchronized write access to the writer's underlying data stream and ensures that each write is flushed immediately.
type WriteFlusher struct {
	http.ResponseWriter
	flusher http.Flusher
	m       sync.Mutex
}

// Write writes the bytes to a stream and flushes the stream.
func (wf *WriteFlusher) Write(b []byte) (n int, err error) {
	wf.m.Lock()
	defer wf.m.Unlock()
	n, err = wf.ResponseWriter.Write(b)
	wf.flusher.Flush()
	return n, err
}

// Flush flushes the stream immediately.
func (wf *WriteFlusher) Flush() {
	wf.m.Lock()
	defer wf.m.Unlock()
	wf.flusher.Flush()
}

// NewWriteFlusher creates a new WriteFlusher for the writer.
func NewWriteFlusher(w http.ResponseWriter) *WriteFlusher {
	var flusher http.Flusher
	if f, ok := w.(http.Flusher); ok {
		flusher = f
	} else {
		flusher = &NopFlusher{}
	}
	return &WriteFlusher{ResponseWriter: w, flusher: flusher}
}



type NopFlusher struct{}

func (f *NopFlusher) Flush() {}

func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
