package d7024e

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func SetupNetworkTest(kademliaPort int, targetPort int) (*Kademlia, *Kademlia) {
	kademlia := &Kademlia{}
	kademlia.Mtx, kademlia.RtMtx = &sync.Mutex{}, &sync.Mutex{}
	kademlia.Me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:"+strconv.Itoa(kademliaPort))
	kademlia.Rt = NewRoutingTable(kademlia.Me)
	kademliaNet := NewNetwork(&kademlia.Me, kademlia.Rt, kademlia, NewMessageChannelManager())
	kademlia.Network = &kademliaNet
	kademlia.K = 20
	kademlia.Alpha = 3

	target := &Kademlia{}
	target.Mtx, target.RtMtx = &sync.Mutex{}, &sync.Mutex{}
	target.Me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:"+strconv.Itoa(targetPort))
	target.Rt = NewRoutingTable(target.Me)
	targetNet := NewNetwork(&target.Me, target.Rt, target, NewMessageChannelManager())
	target.Network = &targetNet
	target.K = 20
	target.Alpha = 3

	go kademlia.Network.Listen(kademlia.Me, kademliaPort)
	go target.Network.Listen(target.Me, targetPort)

	return kademlia, target
}

/*func SetupNetworkTest100() [100]*Kademlia {
	var kademliaArray [100]*Kademlia
	for i := 0; i < 100; i++ {
		kademlia := &Kademlia{}
		if i < 10 {
			kademlia.Me = NewContact(NewKademliaID("FFFFFFFF0000000000000000000000000000000"+strconv.Itoa(i)), "localhost:800"+strconv.Itoa(i))
		} else {
			kademlia.Me = NewContact(NewKademliaID("FFFFFFFF000000000000000000000000000000"+strconv.Itoa(i)), "localhost:80"+strconv.Itoa(i))
		}
		kademlia.Rt = NewRoutingTable(kademlia.Me)
		kademliaNet := NewNetwork(&kademlia.Me, kademlia.Rt, kademlia, NewMessageChannelManager())
		kademlia.Network = &kademliaNet
		kademlia.K = 20
		kademlia.Alpha = 3
		split := strings.Split(kademlia.Me.Address, ":")
		port, _ := strconv.Atoi(split[1])
		go kademlia.Network.Listen(kademlia.Me, port)
		kademliaArray[i] = kademlia
	}
	for i := range kademliaArray {
		fmt.Println(kademliaArray[i].Me)
	}
	return kademliaArray
}
*/

func TestFileStore(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	kademlia.Store([]byte("testvalue"))

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
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	kademlia.Store([]byte("testvalue2"))

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
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	kademlia.Store([]byte("testvalue"))
	kademlia.Store([]byte("testvalue2"))

	expectedHashes := [2]string{"fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8", "7464aed0c369d740a99c39b170d8a2880af87bd9"}
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] ||
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveNotPinned failed: Files not stored properly!")
	}
	kademlia.Pin(Hash([]byte("testvalue2")))
	go kademlia.Purge()

	time.Sleep(6 * time.Second)
	if kademlia.GetFiles()[0].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveNotPinned failed: kademlia.GetFiles() does not have expected hash string!")
	}
	fmt.Println("PASSED TestFilesRemoveNotPinned")
}

func TestFilesRemoveAll(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	kademlia.Store([]byte("testvalue"))
	kademlia.Store([]byte("testvalue2"))

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
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	kademlia.Store([]byte("testvalue"))
	kademlia.Store([]byte("testvalue2"))

	expectedHashes := [2]string{"fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8", "7464aed0c369d740a99c39b170d8a2880af87bd9"}
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] ||
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveNothing failed: Files not stored properly!")
	}
	kademlia.Pin(Hash([]byte("testvalue")))
	kademlia.Pin(Hash([]byte("testvalue2")))
	go kademlia.Purge()

	time.Sleep(6 * time.Second)
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] &&
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveNothing failed: A pinned file was removed!")
	}
	fmt.Println("PASSED TestFilesRemoveNothing")
}

func TestFilesRemoveUnpinned(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	kademlia.Store([]byte("testvalue"))
	kademlia.Store([]byte("testvalue2"))

	expectedHashes := [2]string{"fc3cfe2b4f554c2752444f1e6b7bad1e8d5cccf8", "7464aed0c369d740a99c39b170d8a2880af87bd9"}
	if kademlia.GetFiles()[0].Hash != expectedHashes[0] ||
		kademlia.GetFiles()[1].Hash != expectedHashes[1] {
		t.Fatalf("TestFilesRemoveUnpinned failed: Files not stored properly!")
	}
	kademlia.Pin(Hash([]byte("testvalue")))
	kademlia.Pin(Hash([]byte("testvalue2")))
	go kademlia.Purge()

	kademlia.Unpin(Hash([]byte("testvalue2")))
	time.Sleep(6 * time.Second)

	if kademlia.GetFiles()[0].Hash != expectedHashes[0] &&
		len(kademlia.GetFiles()) != 1 {
		t.Fatalf("TestFilesRemoveUnpinned: A file was not removed!")
	}
	fmt.Println("PASSED TestFilesRemoveUnpinned")
}

func TestLookupData(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	hash := Hash([]byte("testvalue"))
	kademlia.Store([]byte("testvalue"))

	expectedResult := []byte{116, 101, 115, 116, 118, 97, 108, 117, 101}
	val := kademlia.LookupData(hash)
	if !bytes.Equal(val, expectedResult) {
		t.Fatalf("TestLookupData failed: Value mismatch!")
	}
	fmt.Println("PASSED TestLookupData")
}

func TestLookupDataNotFound(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	hash := Hash([]byte("testvalue2"))
	kademlia.Store([]byte("testvalue"))

	val := kademlia.LookupData(hash)
	if val != nil {
		t.Fatalf("TestLookupDataNotFound failed: Found a value!")
	}
	fmt.Println("PASSED TestLookupDataNotFound")
}

func TestLookupFile(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	hash := Hash([]byte("testvalue"))
	kademlia.Store([]byte("testvalue"))

	expectedResult := File{hash, []byte{116, 101, 115, 116, 118, 97, 108, 117, 101}, false, 0}
	file := kademlia.LookupFile(hash)
	if !reflect.DeepEqual(file, &expectedResult) {
		t.Fatalf("TestLookupFile: File mismatch!")
	}
	fmt.Println("PASSED TestLookupFile")
}

func TestLookupFileNotFound(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	hash := Hash([]byte("testvalue2"))
	kademlia.Store([]byte("testvalue2"))

	expectedResult := File{hash, []byte{116, 101, 115, 116, 118, 97, 108, 117, 101}, false, 0}
	file := kademlia.LookupFile(hash)
	if reflect.DeepEqual(file, &expectedResult) {
		t.Fatalf("TestLookupFileNotFound: File match!")
	}
	fmt.Println("PASSED TestLookupFileNotFound")
}

func TestFindValueReturnValue(t *testing.T) {
	kademlia, target := SetupNetworkTest(8000, 8001)

	kademlia.Rt.AddContact(target.Me)

	valueID := NewKademliaID(Hash([]byte("testvalue")))
	target.Store([]byte("testvalue"))

	_, resVal := kademlia.FindValue(valueID, &target.Me)
	if resVal == nil {
		t.Fatalf("TestFindValueReturnValue: returned value was nil!")
	}
	fmt.Println("PASSED TestFindValueReturnValue")
}

func TestFindValueReturnContactsSelf(t *testing.T) {
	kademlia, target := SetupNetworkTest(8002, 8003)

	kademlia.Rt.AddContact(target.Me)
	target.Rt.AddContact(kademlia.Me)

	valueID := NewKademliaID(Hash([]byte("testvalue")))

	resConts, _ := kademlia.FindValue(valueID, &target.Me)
	if resConts == nil {
		t.Fatalf("TestFindValueReturnContactsSelf: returned list of contacts was nil!")
	}
	fmt.Println("PASSED TestFindValueReturnContactsSelf")
}

func TestFindValueReturnContactsNil(t *testing.T) {
	kademlia, target := SetupNetworkTest(8004, 8005)

	kademlia.Rt.AddContact(target.Me)

	valueID := NewKademliaID(Hash([]byte("testvalue")))

	resConts, _ := kademlia.FindValue(valueID, &target.Me)
	if resConts != nil {
		t.Fatalf("TestFindValueReturnContactsNil: returned list of contacts was not nil!")
	}
	fmt.Println("PASSED TestFindValueReturnContactsNil")
}

func TestLookupMessageFindContactReturnNil(t *testing.T) {
	kademlia, target := SetupNetworkTest(8006, 8007)

	kademlia.Rt.AddContact(target.Me)

	rpcType := "FIND_CONTACT"
	valueID := NewKademliaID(Hash([]byte("testvalue")))
	resChan := make(chan []string)
	go kademlia.LookupMessage(rpcType, valueID, &target.Me, resChan)
	resType, resConts := <-resChan, <-resChan
	if resType[0] != "FIND_CONTACT" {
		t.Fatalf("TestLookupMessageFindContactReturnNil: response rpcType was not FIND_CONTACT!")
	}
	if resConts != nil {
		t.Fatalf("TestLookupMessageFindContactReturnNil: response list of contacts was not nil!")
	}
	fmt.Println("PASSED TestLookupMessageFindContactReturnNil")
}

func TestLookupMessageFindContactReturnContacts(t *testing.T) {
	kademlia, target := SetupNetworkTest(8008, 8009)

	kademlia.Rt.AddContact(target.Me)
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000002"), "localhost:8002"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000003"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000004"), "localhost:8004"))

	rpcType := "FIND_CONTACT"
	valueID := NewKademliaID(Hash([]byte("testvalue")))
	resChan := make(chan []string)
	go kademlia.LookupMessage(rpcType, valueID, &target.Me, resChan)
	resType, resConts := <-resChan, <-resChan
	if resType[0] != "FIND_CONTACT" {
		t.Fatalf("TestLookupMessageFindContactReturnContacts: response rpcType was not FIND_CONTACT!")
	}
	if resConts == nil {
		t.Fatalf("TestLookupMessageFindContactReturnContacts: response list of contacts was nil!")
	}
	fmt.Println("PASSED TestLookupMessageFindContactReturnContacts")
}

func TestLookupMessageFindValueReturnValue(t *testing.T) {
	kademlia, target := SetupNetworkTest(8010, 8011)

	kademlia.Rt.AddContact(target.Me)

	target.Store([]byte("testvalue"))
	expectedResult := "testvalue"

	rpcType := "FIND_VALUE"
	valueID := NewKademliaID(Hash([]byte("testvalue")))
	resChan := make(chan []string)
	go kademlia.LookupMessage(rpcType, valueID, &target.Me, resChan)
	resType, resVal := <-resChan, <-resChan
	if resType[0] != "FIND_VALUE" {
		t.Fatalf("TestLookupMessageFindValueReturnValue: response rpcType was not FIND_VALUE!")
	} else if resType[1] != "DATA" {
		t.Fatalf("TestLookupMessageFindValueReturnValue: response data type was not DATA!")
	} else if resVal[0] != expectedResult {
		t.Fatalf("TestLookupMessageFindValueReturnValue: returned value did not match the expected result!")
	}
	fmt.Println("PASSED TestLookupMessageFindValueReturnValue")
}

func TestLookupMessageFindValueReturnContacts(t *testing.T) {
	kademlia, target := SetupNetworkTest(8012, 8013)

	kademlia.Rt.AddContact(target.Me)

	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000002"), "localhost:8002"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000003"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000004"), "localhost:8004"))

	expectedContacts := []Contact{NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000002"), "localhost:8002"),
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000003"), "localhost:8003"),
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000004"), "localhost:8004")}

	expectedResult := []string{expectedContacts[0].String(),
		expectedContacts[1].String(),
		expectedContacts[2].String()}

	rpcType := "FIND_VALUE"
	valueID := NewKademliaID(Hash([]byte("testvalue")))
	resChan := make(chan []string)
	go kademlia.LookupMessage(rpcType, valueID, &target.Me, resChan)
	resType, resVal := <-resChan, <-resChan
	if resType[0] != "FIND_VALUE" {
		t.Fatalf("TestLookupMessageFindValueReturnContacts: response rpcType was not FIND_VALUE!")
	} else if resType[1] != "CONTACTS" {
		t.Fatalf("TestLookupMessageFindValueReturnContacts: response data type was not CONTACTS!")
	} else if !reflect.DeepEqual(resVal, expectedResult) {
		t.Fatalf("TestLookupMessageFindValueReturnContacts: returned contacts did not match the expected result!")
	}
	fmt.Println("PASSED TestLookupMessageFindValueReturnContacts")
}

func TestLookupMessageFindValueReturnNilContacts(t *testing.T) {
	kademlia, target := SetupNetworkTest(8014, 8015)

	kademlia.Rt.AddContact(target.Me)

	rpcType := "FIND_VALUE"
	valueID := NewKademliaID(Hash([]byte("testvalue")))
	resChan := make(chan []string)
	go kademlia.LookupMessage(rpcType, valueID, &target.Me, resChan)
	resType, resVal := <-resChan, <-resChan
	if resType[0] != "FIND_VALUE" {
		t.Fatalf("TestLookupMessageFindValueReturnNilContacts: response rpcType was not FIND_VALUE!")
	} else if resType[1] != "CONTACTS" {
		t.Fatalf("TestLookupMessageFindValueReturnNilContacts: response data type was not CONTACTS!")
	} else if resVal != nil {
		t.Fatalf("TestLookupMessageFindValueReturnNilContacts: returned contacts was not nil!")
	}
	fmt.Println("PASSED TestLookupMessageFindValueReturnNilContacts")
}

/*func Test100Nodes(t *testing.T) {
	SetupNetworkTest100()
}*/
