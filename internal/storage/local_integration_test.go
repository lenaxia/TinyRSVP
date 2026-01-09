package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLocalProvider_Integration_ConcurrentWrites(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			path := filepath.Join("concurrent", "file_"+string(rune('0'+id))+".txt")
			data := []byte("data from goroutine " + string(rune('0'+id)))
			err := provider.PutObject(ctx, path, bytes.NewReader(data), "text/plain")
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent write failed: %v", err)
	}

	objects, err := provider.ListObjects(ctx, "concurrent/")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(objects) != numGoroutines {
		t.Errorf("Expected %d objects, got %d", numGoroutines, len(objects))
	}
}

func TestLocalProvider_Integration_ConcurrentReads(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	path := "test/file.txt"
	testData := []byte("test data for concurrent reads")
	err := provider.PutObject(ctx, path, bytes.NewReader(testData), "text/plain")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	const numGoroutines = 20
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reader, err := provider.GetObject(ctx, path)
			if err != nil {
				errors <- err
				return
			}
			defer reader.Close()

			data, err := io.ReadAll(reader)
			if err != nil {
				errors <- err
				return
			}

			if !bytes.Equal(data, testData) {
				errors <- io.ErrUnexpectedEOF
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent read failed: %v", err)
	}
}

func TestLocalProvider_Integration_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	largeData := make([]byte, 5*1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	path := "large/file.bin"
	err := provider.PutObject(ctx, path, bytes.NewReader(largeData), "application/octet-stream")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	reader, err := provider.GetObject(ctx, path)
	if err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}
	defer reader.Close()

	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if !bytes.Equal(retrieved, largeData) {
		t.Error("Large file data mismatch")
	}

	objects, err := provider.ListObjects(ctx, "large/")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(objects) != 1 {
		t.Fatalf("Expected 1 object, got %d", len(objects))
	}

	if objects[0].Size != int64(len(largeData)) {
		t.Errorf("Object size = %d, want %d", objects[0].Size, len(largeData))
	}
}

func TestLocalProvider_Integration_ManySmallFiles(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	const numFiles = 100
	for i := 0; i < numFiles; i++ {
		path := filepath.Join("many", "file_"+string(rune('0'+i%10))+string(rune('0'+i/10))+".txt")
		data := []byte("data " + string(rune('0'+i%10)) + string(rune('0'+i/10)))
		err := provider.PutObject(ctx, path, bytes.NewReader(data), "text/plain")
		if err != nil {
			t.Fatalf("PutObject() failed for file %d: %v", i, err)
		}
	}

	objects, err := provider.ListObjects(ctx, "many/")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(objects) != numFiles {
		t.Errorf("Expected %d objects, got %d", numFiles, len(objects))
	}
}

func TestLocalProvider_Integration_DirectoryPermissions(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	path := "deep/nested/directory/structure/file.txt"
	err := provider.PutObject(ctx, path, bytes.NewReader([]byte("data")), "text/plain")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	dirs := []string{
		"deep",
		"deep/nested",
		"deep/nested/directory",
		"deep/nested/directory/structure",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(tempDir, dir)
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Fatalf("Failed to stat directory %s: %v", dir, err)
		}

		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}

		if info.Mode().Perm() != 0755 {
			t.Errorf("Directory %s permissions = %o, want 0755", dir, info.Mode().Perm())
		}
	}
}

func TestLocalProvider_Integration_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	paths := []string{
		"file1.txt",
		"dir/file2.txt",
		"deep/nested/file3.txt",
	}

	for _, path := range paths {
		err := provider.PutObject(ctx, path, bytes.NewReader([]byte("data")), "text/plain")
		if err != nil {
			t.Fatalf("PutObject() failed for %s: %v", path, err)
		}

		fullPath := filepath.Join(tempDir, path)
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Fatalf("Failed to stat file %s: %v", path, err)
		}

		if info.Mode().Perm() != 0644 {
			t.Errorf("File %s permissions = %o, want 0644", path, info.Mode().Perm())
		}
	}
}

func TestLocalProvider_Integration_DeleteAndRecreate(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	path := "test/file.txt"
	originalData := []byte("original data")

	err := provider.PutObject(ctx, path, bytes.NewReader(originalData), "text/plain")
	if err != nil {
		t.Fatalf("Initial PutObject() error = %v", err)
	}

	err = provider.DeleteObject(ctx, path)
	if err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}

	_, err = provider.GetObject(ctx, path)
	if err == nil {
		t.Error("Expected error getting deleted file")
	}

	newData := []byte("new data")
	err = provider.PutObject(ctx, path, bytes.NewReader(newData), "text/plain")
	if err != nil {
		t.Fatalf("Recreate PutObject() error = %v", err)
	}

	reader, err := provider.GetObject(ctx, path)
	if err != nil {
		t.Fatalf("GetObject() after recreate error = %v", err)
	}
	defer reader.Close()

	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if !bytes.Equal(retrieved, newData) {
		t.Error("Recreated file has wrong content")
	}
}

func TestLocalProvider_Integration_ComplexWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewLocalProvider(tempDir, "http://localhost:8080")
	ctx := context.Background()

	eventID := "123"
	imagePaths := []string{
		"images/" + eventID + "/logo.png",
		"images/" + eventID + "/banner.jpg",
		"images/" + eventID + "/photo1.png",
		"images/" + eventID + "/photo2.png",
	}

	for _, path := range imagePaths {
		err := provider.PutObject(ctx, path, bytes.NewReader([]byte("image data")), "image/png")
		if err != nil {
			t.Fatalf("PutObject() failed for %s: %v", path, err)
		}
	}

	objects, err := provider.ListObjects(ctx, "images/"+eventID+"/")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}

	if len(objects) != len(imagePaths) {
		t.Errorf("Expected %d objects, got %d", len(imagePaths), len(objects))
	}

	for _, obj := range objects {
		url, err := provider.GetPublicURL(ctx, obj.Path)
		if err != nil {
			t.Errorf("GetPublicURL() failed for %s: %v", obj.Path, err)
		}

		expectedURL := "http://localhost:8080/assets/" + obj.Path
		if url != expectedURL {
			t.Errorf("GetPublicURL() = %s, want %s", url, expectedURL)
		}
	}

	err = provider.DeleteObject(ctx, imagePaths[0])
	if err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}

	objects, err = provider.ListObjects(ctx, "images/"+eventID+"/")
	if err != nil {
		t.Fatalf("ListObjects() after delete error = %v", err)
	}

	if len(objects) != len(imagePaths)-1 {
		t.Errorf("Expected %d objects after delete, got %d", len(imagePaths)-1, len(objects))
	}
}
