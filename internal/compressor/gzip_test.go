package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipHandle(t *testing.T) {
	handler := GzipHandle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))

	t.Run("no gzip support", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Header().Get("Content-Encoding") == "gzip" {
			t.Errorf("expected no gzip encoding")
		}
		if resp.Body.String() != "Hello, World!" {
			t.Errorf("expected Hello, World!, got %s", resp.Body.String())
		}
	})

	t.Run("with gzip support", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("expected gzip encoding")
		}

		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer gz.Close()

		body, err := io.ReadAll(gz)
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != "Hello, World!" {
			t.Errorf("expected Hello, World!, got %s", string(body))
		}
	})

	t.Run("compressed request body", func(t *testing.T) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write([]byte("Compressed Request"))
		gz.Close()

		req := httptest.NewRequest("POST", "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		handlerWithEcho := GzipHandle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			w.Write(body)
		}))

		resp := httptest.NewRecorder()
		handlerWithEcho.ServeHTTP(resp, req)

		gzResp, err := gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer gzResp.Close()

		body, _ := io.ReadAll(gzResp)
		if string(body) != "Compressed Request" {
			t.Errorf("expected Compressed Request, got %s", string(body))
		}
	})
}
