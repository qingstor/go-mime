package mime

import (
	"mime"
	"testing"
)

func TestDetectFileExt(t *testing.T) {
	x := DetectFileExt("pdf")
	if x != "application/pdf" {
		t.Errorf("expect %s instead of %s", "application/pdf", x)
	}
}

func TestDetectFilePath(t *testing.T) {
	x := DetectFilePath("/root/a.pdf")
	if x != "application/pdf" {
		t.Errorf("expect %s instead of %s", "application/pdf", x)
	}
}

func BenchmarkDetectFilePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DetectFilePath("/root/a.pdf")
	}
}

func BenchmarkDetectFileExt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DetectFileExt("pdf")
	}
}

func BenchmarkDetectFileExtWithMissingExt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DetectFileExt("asdasdasdaddasda")
	}
}

func BenchmarkGoMime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mime.TypeByExtension("pdf")
	}
}

func BenchmarkGoMimeWithMissingExt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mime.TypeByExtension("asdasdasdaddasda")
	}
}
