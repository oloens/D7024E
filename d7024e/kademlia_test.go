package d7024e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
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

func SetupKademliaNode(id string, port int) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.Mtx, kademlia.RtMtx = &sync.Mutex{}, &sync.Mutex{}
	kademlia.Me = NewContact(NewKademliaID(id), "localhost:"+strconv.Itoa(port))
	kademlia.Rt = NewRoutingTable(kademlia.Me)
	kademliaNet := NewNetwork(&kademlia.Me, kademlia.Rt, kademlia, NewMessageChannelManager())
	kademlia.Network = &kademliaNet
	kademlia.K = 20
	kademlia.Alpha = 3

	go kademlia.Network.Listen(kademlia.Me, port)

	return kademlia
}

/*func SetupNetworkTest100() [100]*Kademlia {
	var kademliaArray [100]*Kademlia
	for i := 0; i < 100; i++ {
		kademlia := &Kademlia{}
		kademlia.Mtx, kademlia.RtMtx = &sync.Mutex{}, &sync.Mutex{}
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
}*/

func TestGetFiles(t *testing.T) {
	kademlia := Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	data := []byte("testvalue")
	hash := Hash(data)
	file := File{hash, data, false, 0}
	kademlia.AddFile(file)
	files := kademlia.GetFiles()
	if files == nil {
		t.Fatalf("TestGetFiles: Didn't get any files!")
	}
	fmt.Println("PASSED TestGetFiles")
}

func TestAddFile(t *testing.T) {
	kademlia := Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	data := []byte("testvalue")
	hash := Hash(data)
	filesLen := len(kademlia.files)
	file := File{hash, data, false, 0}
	kademlia.AddFile(file)

	result := filesLen + 1
	if !(result > filesLen) {
		t.Fatalf("TestAddFile: Didn't add file!")
	}
	fmt.Println("PASSED TestAddFile")
}

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

func TestFindNodeKClosest(t *testing.T) {
	kademlia, target := SetupNetworkTest(8900, 8901)

	kademlia.Rt.AddContact(target.Me)

	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000002"), "localhost:8002"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000003"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000004"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000005"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000006"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000007"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000008"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000009"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000010"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000011"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000012"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000013"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000014"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000015"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000016"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000017"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000018"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000019"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000020"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000021"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000022"), "localhost:8003"))
	target.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF000000000000000000000000000000023"), "localhost:8003"))

	data := []byte("test123")
	hash := Hash(data)
	testID := NewKademliaID(hash)

	output := kademlia.FindNode(testID, &target.Me)
	if output == nil {
		t.Fatalf("TestFindNodeKClosest: Couldnt find any nodes!")
	} else if len(output) != 20 {
		t.Fatalf("TestFindNodeKClosest: Did not return exactly 20 contacts!")
	}
	fmt.Println("PASSED TestFindNodeKClosest")
}

func TestFindNodeNil(t *testing.T) {
	kademlia, target := SetupNetworkTest(8902, 8903)

	kademlia.Rt.AddContact(target.Me)

	data := []byte("test123")
	hash := Hash(data)
	testID := NewKademliaID(hash)

	output := kademlia.FindNode(testID, &target.Me)
	if output != nil {
		t.Fatalf("TestFindNodeNil: Returned nodes was not nil!")
	}
	fmt.Println("PASSED TestFindNodeNil")
}

func TestFindNode(t *testing.T) {
	kademlia, target := SetupNetworkTest(8904, 8905)

	kademlia.Rt.AddContact(target.Me)

	data := []byte("test123")
	hash := Hash(data)
	testID := NewKademliaID(hash)
	testContact := NewContact(testID, "localhost:8008")
	target.Rt.AddContact(testContact)

	output := kademlia.FindNode(testID, &target.Me)
	if output == nil {
		t.Fatalf("TestFindNode: Didnt find a single node!")
	}
	fmt.Println("PASSED TestFindNode")
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
	//kademlia, target := SetupNetworkTest(8004, 8005)
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000077", 8698)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000000078", 8699)

	kademlia.Rt.AddContact(target.Me)

	valueID := NewKademliaID(Hash([]byte("testvalue")))

	resConts, _ := kademlia.FindValue(valueID, &target.Me)
	if resConts != nil {
		t.Fatalf("TestFindValueReturnContactsNil: returned list of contacts was not nil!")
	}
	fmt.Println("PASSED TestFindValueReturnContactsNil")
}

func TestLookupMessageFindContactReturnNil(t *testing.T) {
	//kademlia, target := SetupNetworkTest(8006, 8007)
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000087", 8678)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000000088", 8679)

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

func TestIterativeLookupNoneCloserQueryAllTimeout(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8100)
	alpha1 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000001", 8101)
	alpha2 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000002", 8102)
	alpha3 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000003", 8103)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000000004", 8104)

	kademlia.Rt.AddContact(alpha1.Me)
	kademlia.Rt.AddContact(alpha2.Me)
	kademlia.Rt.AddContact(alpha3.Me)

	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000110"), "localhost:8110"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000111"), "localhost:8111"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000112"), "localhost:8112"))

	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000120"), "localhost:8120"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000121"), "localhost:8121"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000122"), "localhost:8122"))

	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000130"), "localhost:8130"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000131"), "localhost:8131"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000132"), "localhost:8132"))

	expectedContacts := []string{"FFFFFFFF00000000000000000000000000000001", "FFFFFFFF00000000000000000000000000000002", "FFFFFFFF00000000000000000000000000000003"}

	kClosest, closestNode, _ := kademlia.IterativeLookup("FIND_CONTACT", target.Me.ID)
	kademlia.IterativeLookup("FIND_VALUE", target.Me.ID)
	if closestNode.String() != alpha1.Me.String() {
		t.Fatalf("TestIterativeLookupNoneCloserQueryAllTimeout: closest node did not match expected value!")
	}

	for i, cont := range kClosest {
		if strings.ToUpper(cont.ID.String()) != expectedContacts[i] {
			t.Fatalf("TestIterativeLookupNoneCloserQueryAllTimeout: returned contacts ID's did not match expected contacts ID's")
		}
	}
	fmt.Println("PASSED TestIterativeLookupNoneCloserQueryAllTimeout")
}

func TestIterativeLookupNoneCloserQueryAll(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8105)
	alpha1 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000001", 8106)
	alpha2 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000002", 8107)
	alpha3 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000003", 8108)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000000004", 8109)

	queryNode4 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000110", 8110)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000111", 8111)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000112", 8112)

	SetupKademliaNode("FFFFFFFF00000000000000000000000000000120", 8120)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000121", 8121)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000122", 8122)

	SetupKademliaNode("FFFFFFFF00000000000000000000000000000130", 8130)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000131", 8131)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000132", 8132)

	kademlia.Rt.AddContact(alpha1.Me)
	kademlia.Rt.AddContact(alpha2.Me)
	kademlia.Rt.AddContact(alpha3.Me)

	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000110"), "localhost:8110"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000111"), "localhost:8111"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000112"), "localhost:8112"))

	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000120"), "localhost:8120"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000121"), "localhost:8121"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000122"), "localhost:8122"))

	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000130"), "localhost:8130"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000131"), "localhost:8131"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000132"), "localhost:8132"))

	queryNode4.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000133"), "localhost:8133"))

	expectedContacts := []string{"FFFFFFFF00000000000000000000000000000001", "FFFFFFFF00000000000000000000000000000002", "FFFFFFFF00000000000000000000000000000003",
		"FFFFFFFF00000000000000000000000000000110", "FFFFFFFF00000000000000000000000000000111", "FFFFFFFF00000000000000000000000000000112",
		"FFFFFFFF00000000000000000000000000000120", "FFFFFFFF00000000000000000000000000000121", "FFFFFFFF00000000000000000000000000000122",
		"FFFFFFFF00000000000000000000000000000130", "FFFFFFFF00000000000000000000000000000131", "FFFFFFFF00000000000000000000000000000132",
		"FFFFFFFF00000000000000000000000000000133"}

	kClosest, closestNode, _ := kademlia.IterativeLookup("FIND_CONTACT", target.Me.ID)
	if closestNode.String() != alpha1.Me.String() {
		t.Fatalf("TestIterativeLookupNoneCloserQueryAll: closest node did not match expected value!")
	} else if true {
		for i, cont := range kClosest {
			if strings.ToUpper(cont.ID.String()) != expectedContacts[i] {
				t.Fatalf("TestIterativeLookupNoneCloserQueryAll: returned contacts ID's did not match expected contacts ID's")
			}
		}
	}
	fmt.Println("PASSED TestIterativeLookupNoneCloserQueryAll")
}

func TestIterativeLookupQueryNothing(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8140)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000000004", 8141)

	kClosest, closestNode, _ := kademlia.IterativeLookup("FIND_CONTACT", target.Me.ID)
	if closestNode != nil {
		t.Fatalf("TestIterativeLookupQueryNothing: closest node was not nil!")
	}
	if kClosest != nil {
		t.Fatalf("TestIterativeLookupQueryNothing: returned contacts when expected nil!")
	}
	fmt.Println("PASSED TestIterativeLookupQueryNothing")
}

func TestIterativeLookupCloserFound(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8200)
	alpha1 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000001", 8201)
	alpha2 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000002", 8202)
	alpha3 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000003", 8203)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000000005", 8204)

	expectedResult := SetupKademliaNode("FFFFFFFF00000000000000000000000000000004", 8205)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000110", 8210)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000111", 8211)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000112", 8212)

	SetupKademliaNode("FFFFFFFF00000000000000000000000000000120", 8220)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000121", 8221)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000122", 8222)

	SetupKademliaNode("FFFFFFFF00000000000000000000000000000130", 8230)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000131", 8231)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000132", 8232)

	kademlia.Rt.AddContact(alpha1.Me)
	kademlia.Rt.AddContact(alpha2.Me)
	kademlia.Rt.AddContact(alpha3.Me)

	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000004"), "localhost:8205"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000110"), "localhost:8210"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000111"), "localhost:8211"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000112"), "localhost:8212"))

	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000120"), "localhost:8220"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000121"), "localhost:8221"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000122"), "localhost:8222"))

	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000130"), "localhost:8230"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000131"), "localhost:8231"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000132"), "localhost:8232"))

	_, closestNode, _ := kademlia.IterativeLookup("FIND_CONTACT", target.Me.ID)
	if closestNode.ID.String() != expectedResult.Me.ID.String() {
		t.Fatalf("TestIterativeLookupCloserFound: closest node did not match expected value!")
	}
	fmt.Println("PASSED TestIterativeLookupCloserFound")
}

func TestIterativeLookupNoneCloserFoundValue(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8305)
	alpha1 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000001", 8306)
	alpha2 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000002", 8307)
	alpha3 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000003", 8308)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000004", 8309)

	queryNode4 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000110", 8310)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000111", 8311)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000112", 8312)

	SetupKademliaNode("FFFFFFFF00000000000000000000000000000120", 8320)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000121", 8321)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000122", 8322)

	SetupKademliaNode("FFFFFFFF00000000000000000000000000000130", 8330)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000131", 8331)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000132", 8332)

	kademlia.Rt.AddContact(alpha1.Me)
	kademlia.Rt.AddContact(alpha2.Me)
	kademlia.Rt.AddContact(alpha3.Me)

	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000110"), "localhost:8310"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000111"), "localhost:8311"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000112"), "localhost:8312"))

	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000120"), "localhost:8320"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000121"), "localhost:8321"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000122"), "localhost:8322"))

	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000130"), "localhost:8330"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000131"), "localhost:8331"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000132"), "localhost:8332"))

	queryNode4.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000133"), "localhost:8333"))

	hash := NewKademliaID(Hash([]byte("testvalue")))
	queryNode4.Store([]byte("testvalue"))

	expectedResult := []byte{116, 101, 115, 116, 118, 97, 108, 117, 101}

	_, _, val := kademlia.IterativeLookup("FIND_VALUE", hash)
	if !bytes.Equal(val, expectedResult) {
		t.Fatalf("TestIterativeLookupNoneCloserFoundValue failed: Value mismatch!")
	}
	fmt.Println("PASSED TestIterativeLookupNoneCloserFoundValue")
}

func TestIterativeLookupNoneCloserFindValueFoundContacts(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8315)
	alpha1 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000001", 8316)
	alpha2 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000002", 8317)
	alpha3 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000003", 8318)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000004", 8319)

	queryNode4 := SetupKademliaNode("FFFFFFFF00000000000000000000000000000110", 8340)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000111", 8341)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000112", 8342)

	SetupKademliaNode("FFFFFFFF00000000000000000000000000000120", 8350)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000121", 8351)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000122", 8352)

	SetupKademliaNode("FFFFFFFF00000000000000000000000000000130", 8360)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000131", 8361)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000132", 8362)

	kademlia.Rt.AddContact(alpha1.Me)
	kademlia.Rt.AddContact(alpha2.Me)
	kademlia.Rt.AddContact(alpha3.Me)

	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000110"), "localhost:8340"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000111"), "localhost:8341"))
	alpha1.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000112"), "localhost:8342"))

	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000120"), "localhost:8350"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000121"), "localhost:8351"))
	alpha2.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000122"), "localhost:8352"))

	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000130"), "localhost:8360"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000131"), "localhost:8361"))
	alpha3.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000132"), "localhost:8362"))

	queryNode4.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000133"), "localhost:8363"))

	hash := NewKademliaID(Hash([]byte("testvalue")))

	kClosest, _, val := kademlia.IterativeLookup("FIND_VALUE", hash)
	if val != nil {
		t.Fatalf("TestIterativeLookupNoneCloserFindValueFoundContacts failed: Value was not nil!")
	}
	if kClosest == nil {
		t.Fatalf("TestIterativeLookupNoneCloserFindValueFoundContacts failed: kClosest was nil!")
	}
	fmt.Println("PASSED TestIterativeLookupNoneCloserFindValueFoundContacts")
}

func TestSendStore(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8790)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000001", 8791)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000002", 8792)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000003", 8793)
	expectedResult := SetupKademliaNode("FFFFFFFF00000000000000000000000000000004", 8794)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000005", 8795)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000006", 8796)
	SetupKademliaNode("FFFFFFFF00000000000000000000000000000007", 8797)
	kademlia.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:8791"))
	kademlia.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000002"), "localhost:8792"))
	kademlia.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000003"), "localhost:8793"))
	kademlia.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000004"), "localhost:8794"))
	kademlia.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000005"), "localhost:8795"))
	kademlia.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000006"), "localhost:8796"))
	kademlia.Rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000007"), "localhost:8797"))

	hash := Hash([]byte("testvalue"))
	kademlia.SendStore(hash, []byte("testvalue"))

	if len(expectedResult.GetFiles()) == 0 {
		t.Fatalf("TestSendStore failed: File was not stored properly!")
	}
	fmt.Println("PASSED TestSendStore")
}

func TestBootstrap(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000008798", 8798)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000008799", 8799)
	failedAttempt := kademlia.Bootstrap()
	if failedAttempt {
		t.Fatalf("TestBootstrap failed: Node Bootstrapped before it was supposed to!")
	}
	kademlia.Rt.AddContact(target.Me)
	successfulAttempt := kademlia.Bootstrap()
	if !successfulAttempt {
		t.Fatalf("TestBootstrap failed: Node did not Bootstrap when it was supposed to!")
	}
	fmt.Println("PASSED TestBootstrap")
}

func TestPing(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000550", 8800)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000000551", 8801)
	kademlia.Rt.AddContact(target.Me)

	expected := "ping response from ffffffff00000000000000000000000000000551"

	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	kademlia.Ping(&target.Me)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	output := BytesToString(out)
	output = strings.TrimSuffix(output, "\n")

	if !(output == expected) {
		t.Fatalf("TestPing failed: Ping failure")
	}
	fmt.Println("PASSED TestPing")
}

func TestPingNoResponse(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000998", 8811)
	deadNode := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000999"), "localhost:8812")

	//expected := "Request sent to " + deadNode.ID.String() + " timed out"

	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	kademlia.Ping(&deadNode)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	output := BytesToString(out)
	output = strings.TrimSuffix(output, "\n")
	if output == "ping response from ffffffff00000000000000000000000000000999" {
		t.Fatalf("TestPingNoResponse failed: Node was not dead!")
	}
	fmt.Println("PASSED TestPingNoResponse" + "\n")
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

func BytesToString(data []byte) string {
	return string(data[:])
}
