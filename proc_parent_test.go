package overseer

import (
	"os"
	"sync"
	"testing"
	"time"
)

// TestOverwriteBinary tests the overwriteBinary function and the withFileLock functionality.
func TestOverwriteBinaryPositivePath(t *testing.T) {
	// Create temporary files to simulate binPath and tmpBinPath.
	binFile, err := os.CreateTemp("", "overseer_bin")
	if err != nil {
		t.Fatalf("Failed to create temp bin file: %v", err)
	}
	defer os.Remove(binFile.Name())

	tmpBinFile, err := os.CreateTemp("", "overseer_tmp_bin")
	if err != nil {
		t.Fatalf("Failed to create temp tmpBin file: %v", err)
	}
	defer os.Remove(tmpBinFile.Name())

	// Write initial content to binPath.
	initialContent := []byte("initial binary content")
	if _, err := binFile.Write(initialContent); err != nil {
		t.Fatalf("Failed to write to bin file: %v", err)
	}
	binFile.Close()

	// Write new content to tmpBinPath.
	newContent := []byte("new binary content")
	if _, err := tmpBinFile.Write(newContent); err != nil {
		t.Fatalf("Failed to write to tmpBin file: %v", err)
	}
	tmpBinFile.Close()

	// Create a parent struct with binPath set to the fake bin file.
	mp := &parent{
		binPath: binFile.Name(),
	}

	// Test that overwriteBinary correctly overwrites binPath.
	err = mp.overwriteBinary(tmpBinFile.Name())
	if err != nil {
		t.Fatalf("overwriteBinary failed: %v", err)
	}

	// Read the content of binPath and verify it matches newContent.
	updatedContent, err := os.ReadFile(mp.binPath)
	if err != nil {
		t.Fatalf("Failed to read updated bin file: %v", err)
	}

	if string(updatedContent) != string(newContent) {
		t.Errorf("binPath content does not match new content. Got: %s, Want: %s", string(updatedContent), string(newContent))
	}
}

func TestOverwriteBinarySafety(t *testing.T) {
	// Create temporary files to simulate binPath and tmpBinPath.
	binFile, err := os.CreateTemp("", "overseer_bin")
	if err != nil {
		t.Fatalf("Failed to create temp bin file: %v", err)
	}
	defer os.Remove(binFile.Name())

	// Create a parent struct with binPath set to the fake bin file.
	mp := &parent{
		binPath: binFile.Name(),
	}

	// Create two parent instances to simulate two processes.
	mp1 := &parent{
		binPath: binFile.Name(),
	}
	mp2 := &parent{
		binPath: binFile.Name(),
	}

	// Create new temp files to act as tmpBinPath for the concurrent overwrites.
	tmpBinFile1, err := os.CreateTemp("", "overseer_tmp_bin1")
	if err != nil {
		t.Fatalf("Failed to create temp tmpBin file 1: %v", err)
	}
	defer os.Remove(tmpBinFile1.Name())

	tmpBinFile2, err := os.CreateTemp("", "overseer_tmp_bin2")
	if err != nil {
		t.Fatalf("Failed to create temp tmpBin file 2: %v", err)
	}
	defer os.Remove(tmpBinFile2.Name())

	// Write different content to tmpBinPath files.
	newContent1 := []byte("new binary content 1")
	if _, err := tmpBinFile1.Write(newContent1); err != nil {
		t.Fatalf("Failed to write to tmpBin file 1: %v", err)
	}
	tmpBinFile1.Close()

	newContent2 := []byte("new binary content 2")
	if _, err := tmpBinFile2.Write(newContent2); err != nil {
		t.Fatalf("Failed to write to tmpBin file 2: %v", err)
	}
	tmpBinFile2.Close()

	// Use a WaitGroup to synchronize the goroutines.
	var wg sync.WaitGroup
	wg.Add(2)

	// Variables to capture errors from goroutines.
	var err1, err2 error

	// Start the first overwriteBinary call in a goroutine.
	go func() {
		defer wg.Done()
		err1 = mp1.overwriteBinary(tmpBinFile1.Name())
	}()

	// Small delay to increase the chance of overlap.
	time.Sleep(10 * time.Millisecond)

	// Start the second overwriteBinary call in another goroutine.
	go func() {
		defer wg.Done()
		err2 = mp2.overwriteBinary(tmpBinFile2.Name())
	}()

	// Wait for both goroutines to finish.
	wg.Wait()

	// Both calls should succeed without errors.
	if err1 != nil {
		t.Errorf("First overwriteBinary call failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second overwriteBinary call failed: %v", err2)
	}

	// Verify that the binPath content is one of the new contents.
	finalContent, err := os.ReadFile(mp.binPath)
	if err != nil {
		t.Fatalf("Failed to read final bin file: %v", err)
	}

	finalContentStr := string(finalContent)
	if finalContentStr != string(newContent1) && finalContentStr != string(newContent2) {
		t.Errorf("Final binPath content does not match any of the new contents. Got: %s", finalContentStr)
	}
}
