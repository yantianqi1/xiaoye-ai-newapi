package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const onePixelPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+Xk9sAAAAASUVORK5CYII="

func TestInitLocalStorageUsesConfiguredDirectory(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("LOCAL_STORAGE_DIR", tempDir)

	oldBucket := bucket
	oldUseLocalStorage := useLocalStorage
	oldLocalStorageDir := localStorageDir
	t.Cleanup(func() {
		bucket = oldBucket
		useLocalStorage = oldUseLocalStorage
		localStorageDir = oldLocalStorageDir
	})

	bucket = nil
	useLocalStorage = false
	localStorageDir = ""

	InitLocalStorage()

	if LocalStorageDir() != tempDir {
		t.Fatalf("expected local storage dir %s, got %s", tempDir, LocalStorageDir())
	}

	url, err := UploadBase64Image(onePixelPNGBase64, "42", "useredit")
	if err != nil {
		t.Fatalf("upload base64 image: %v", err)
	}
	if !strings.HasPrefix(url, "/api/uploads/useredit/") {
		t.Fatalf("unexpected upload URL: %s", url)
	}

	entries, err := os.ReadDir(filepath.Join(tempDir, "useredit"))
	if err != nil {
		t.Fatalf("read uploaded files: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one uploaded file, got %d", len(entries))
	}
}
