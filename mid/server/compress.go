// Copyright 2013 The Gorilla Authors.
// Copyright 2020 Joseph Boudou.
// All rights reserved. Use of this source code is governed by
// a BSD-style license that can be found in the LICENSE file.

package server

import (
	"compress/flate"
	"compress/gzip"
	"crypto/rand"
	"io"
	"net/http"
	"strings"

	"github.com/JBoudou/Itero/pkg/b64buff"
	"github.com/felixge/httpsnoop"
)

const acceptEncoding string = "Accept-Encoding"

type compressResponseWriter struct {
	compressor io.Writer
	w          http.ResponseWriter
}

func (cw *compressResponseWriter) WriteHeader(c int) {
	cw.w.Header().Del("Content-Length")
	cw.w.WriteHeader(c)
}

func (cw *compressResponseWriter) Write(b []byte) (int, error) {
	h := cw.w.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}
	h.Del("Content-Length")

	return cw.compressor.Write(b)
}

func (cw *compressResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	return io.Copy(cw.compressor, r)
}

type flusher interface {
	Flush() error
}

func (w *compressResponseWriter) Flush() {
	// Flush compressed data if compressor supports it.
	if f, ok := w.compressor.(flusher); ok {
		f.Flush()
	}
	// Flush HTTP response.
	if f, ok := w.w.(http.Flusher); ok {
		f.Flush()
	}
}

// Compress gzip compresses HTTP responses for clients that support it
// via the 'Accept-Encoding' header.
//
// Compressing TLS traffic may leak the page contents to an attacker if the
// page contains user input: http://security.stackexchange.com/a/102015/12208
func Compress(h http.Handler) http.Handler {
	return CompressHandlerLevel(h, gzip.DefaultCompression)
}

// CompressHandlerLevel gzip compresses HTTP responses with specified compression level
// for clients that support it via the 'Accept-Encoding' header.
//
// The compression level should be gzip.DefaultCompression, gzip.NoCompression,
// or any integer value between gzip.BestSpeed and gzip.BestCompression inclusive.
// gzip.DefaultCompression is used in case of invalid compression level.
func CompressHandlerLevel(h http.Handler, level int) http.Handler {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	const (
		gzipEncoding  = "gzip"
		flateEncoding = "deflate"
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// To prevent content leak, compress is discarded on non-GET requests.
		if r.Method != "GET" {
			h.ServeHTTP(w, r)
			return
		}

		// detect what encoding to use
		var encoding string
		for _, curEnc := range strings.Split(r.Header.Get(acceptEncoding), ",") {
			curEnc = strings.TrimSpace(curEnc)
			if curEnc == gzipEncoding || curEnc == flateEncoding {
				encoding = curEnc
				break
			}
		}

		// always add Accept-Encoding to Vary to prevent intermediate caches corruption
		w.Header().Add("Vary", acceptEncoding)

		// if we weren't able to identify an encoding we're familiar with, pass on the
		// request to the handler and return
		if encoding == "" {
			h.ServeHTTP(w, r)
			return
		}

		if r.Header.Get("Upgrade") != "" {
			h.ServeHTTP(w, r)
			return
		}

		// Mitigate BREACH with a variable size header
		if breach, err := breachString(); err == nil {
			w.Header().Add("X-Breach", breach)
		}

		// wrap the ResponseWriter with the writer for the chosen encoding
		var encWriter io.WriteCloser
		if encoding == gzipEncoding {
			encWriter, _ = gzip.NewWriterLevel(w, level)
		} else if encoding == flateEncoding {
			encWriter, _ = flate.NewWriter(w, level)
		}
		defer encWriter.Close()

		w.Header().Set("Content-Encoding", encoding)
		r.Header.Del(acceptEncoding)

		cw := &compressResponseWriter{
			w:          w,
			compressor: encWriter,
		}

		w = httpsnoop.Wrap(w, httpsnoop.Hooks{
			Write: func(httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return cw.Write
			},
			WriteHeader: func(httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return cw.WriteHeader
			},
			Flush: func(httpsnoop.FlushFunc) httpsnoop.FlushFunc {
				return cw.Flush
			},
			ReadFrom: func(rff httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return cw.ReadFrom
			},
		})

		h.ServeHTTP(w, r)
	})
}

func breachString() (ret string, err error) {
	buff := make([]byte, 1)
	if _, err := rand.Reader.Read(buff); err != nil {
		return "", err
	}
	size := uint32(buff[0]) & 0x1F
	return b64buff.RandomString(size)
}
