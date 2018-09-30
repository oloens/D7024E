package d7024e

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestFileStore(t *testing.T) {
	kademlia := Kademlia{}
	data := []byte("testvalue")
	kademlia.Store(data)

	expectedResult := "fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8"
	found := false
	for _, files := range kademlia.GetFiles() {
		if files.Hash == expectedResult {
			found = true
		}
	}
	if !found {
		t.Fatalf("TestFileStore failed: Could not Find File!")
	}
	fmt.Println("PASSED TestFileStore")
}

func TestFileStoreNotFound(t *testing.T) {
	kademlia := Kademlia{}
	data := []byte("testvalue2")
	kademlia.Store(data)

	expectedResult := "fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8"
	found := false
	for _, files := range kademlia.GetFiles() {
		if files.Hash == expectedResult {
			found = true
		}
	}
	if found {
		t.Fatalf("TestFileStoreNotFound failed: Found File!")
	}
	fmt.Println("PASSED TestFileStoreNotFound")
}

func TestFilesRemoveNotPinned(t *testing.T) {
	kademlia := Kademlia{}
	val := []byte("testvalue")
	val2 := []byte("testvalue2")
	kademlia.Store(val)
	kademlia.Store(val2)

	expectedHashes := [2]string{"fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8", "7464aed0c369d740a99c39b170d8a2880af87bd9"}
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] ||
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveNotPinned failed: Files not stored properly!")
	}
	hash2 := Hash(val2)
	kademlia.Pin(hash2)
	go kademlia.Purge()

	time.Sleep(6 * time.Second)
	if kademlia.GetFiles()[0].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveNotPinned failed: kademlia.GetFiles() does not have expected hash string!")
	}
	fmt.Println("PASSED TestFilesRemoveNotPinned")
}

func TestFilesRemoveAll(t *testing.T) {
	kademlia := Kademlia{}
	val := []byte("testvalue")
	val2 := []byte("testvalue2")
	kademlia.Store(val)
	kademlia.Store(val2)

	expectedHashes := [2]string{"fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8", "7464aed0c369d740a99c39b170d8a2880af87bd9"}
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] ||
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveAll failed: Files not stored properly!")
	}
	go kademlia.Purge()

	time.Sleep(6 * time.Second)
	if len(kademlia.GetFiles()) != 0 {
		t.Fatalf("TestFilesRemoveAll failed: kademlia.GetFiles() is not empty!")
	}
	fmt.Println("PASSED TestFilesRemoveAll")
}

func TestFilesRemoveNothing(t *testing.T) {
	kademlia := Kademlia{}
	val := []byte("testvalue")
	val2 := []byte("testvalue2")
	kademlia.Store(val)
	kademlia.Store(val2)

	expectedHashes := [2]string{"fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8", "7464aed0c369d740a99c39b170d8a2880af87bd9"}
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] ||
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveNothing failed: Files not stored properly!")
	}
	hash1 := Hash(val)
	hash2 := Hash(val2)
	kademlia.Pin(hash1)
	kademlia.Pin(hash2)
	go kademlia.Purge()

	time.Sleep(6 * time.Second)
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] &&
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveNothing failed: A pinned file was removed!")
	}
	fmt.Println("PASSED TestFilesRemoveNothing")
}

func TestFilesRemoveUnpinned(t *testing.T) {
	kademlia := Kademlia{}
	val := []byte("testvalue")
	val2 := []byte("testvalue2")
	kademlia.Store(val)
	kademlia.Store(val2)

	expectedHashes := [2]string{"fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8", "7464aed0c369d740a99c39b170d8a2880af87bd9"}
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] ||
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveUnpinned failed: Files not stored properly!")
	}
	hash1 := Hash(val)
	hash2 := Hash(val2)
	kademlia.Pin(hash1)
	kademlia.Pin(hash2)
	go kademlia.Purge()

	kademlia.Unpin(hash2)
	time.Sleep(6 * time.Second)

	if kademlia.GetFiles()[0].Hash != expectedHashes[0] &&
		len(kademlia.GetFiles()) != 1 {
		t.Fatalf("TestFilesRemoveUnpinned: A file was not removed!")
	}
	fmt.Println("PASSED TestFilesRemoveUnpinned")
}

func TestLookupData(t *testing.T) {
	kademlia := Kademlia{}
	data := []byte("testvalue")
	hash := Hash(data)
	kademlia.Store(data)

	expectedResult := []byte{116, 101, 115, 116, 118, 97, 108, 117, 101}
	val := kademlia.LookupData(hash)
	if !bytes.Equal(val, expectedResult) {
		t.Fatalf("TestLookupData failed: Value mismatch!")
	}
	fmt.Println("PASSED TestLookupData")
}

func TestLookupDataNotFound(t *testing.T) {
	kademlia := Kademlia{}
	data := []byte("testvalue2")
	hash := Hash(data)
	kademlia.Store(data)

	expectedResult := []byte{116, 101, 115, 116, 118, 97, 108, 117, 101}
	val := kademlia.LookupData(hash)
	if bytes.Equal(val, expectedResult) {
		t.Fatalf("TestLookupDataNotFound failed: Value match!")
	}
	fmt.Println("PASSED TestLookupDataNotFound")
}

func TestLookupFile(t *testing.T) {
	kademlia := Kademlia{}
	data := []byte("testvalue")
	hash := Hash(data)
	kademlia.Store(data)

	expectedResult := File{hash, []byte{116, 101, 115, 116, 118, 97, 108, 117, 101}, false, 0}
	file := kademlia.LookupFile(hash)
	if !reflect.DeepEqual(file, &expectedResult) {
		t.Fatalf("TestLookupFile: File mismatch!")
	}
	fmt.Println("PASSED TestLookupFile")
}

func TestLookupFileNotFound(t *testing.T) {
	kademlia := Kademlia{}
	data := []byte("testvalue2")
	hash := Hash(data)
	kademlia.Store(data)

	expectedResult := File{hash, []byte{116, 101, 115, 116, 118, 97, 108, 117, 101}, false, 0}
	file := kademlia.LookupFile(hash)
	if reflect.DeepEqual(file, &expectedResult) {
		t.Fatalf("TestLookupFileNotFound: File match!")
	}
	fmt.Println("PASSED TestLookupFileNotFound")
}
