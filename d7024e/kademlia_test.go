package d7024e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

func TestGetFiles(t *testing.T) {
	kademlia := Kademlia{}
	data := []byte("bajs")
	hash := Hash(data)
	file := File{hash, data, false, 0}
	kademlia.AddFile(file)
	files := kademlia.GetFiles()
	fmt.Println("GetFiles")
	if files == nil {
		t.Fatalf("Didn't get any files")
	}
}

func TestAddFile(t *testing.T) {
	kademlia := Kademlia{}
	data := []byte("test123")
	hash := Hash(data)
	filesLen := len(kademlia.files)
	file := File{hash, data, false, 0}
	kademlia.AddFile(file)

	result := filesLen + 1
	fmt.Println("AddFile")
	if !(result > filesLen) {
		t.Fatalf("Error, didn't add file")
	}
}

func TestFindNodeKClosest(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	kademlia.Rt = NewRoutingTable(kademlia.Me)
	kademliaNet := NewNetwork(&kademlia.Me, kademlia.Rt, kademlia, NewMessageChannelManager())
	kademlia.Network = &kademliaNet
	kademlia.K = 20
	kademlia.Alpha = 3

	target := &Kademlia{}
	target.Me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:8001")
	target.Rt = NewRoutingTable(target.Me)
	targetNet := NewNetwork(&target.Me, target.Rt, target, NewMessageChannelManager())
	target.Network = &targetNet
	target.K = 20
	target.Alpha = 3

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

	go kademliaNet.Listen(kademlia.Me, 8000)
	go targetNet.Listen(target.Me, 8001)

	data := []byte("test123")
	hash := Hash(data)
	testID := NewKademliaID(hash)

	output := kademlia.FindNode(testID, &target.Me)
	fmt.Println(output)
	if output == nil {
		t.Errorf("TestFindNode failed..")
	}

}

func TestFindNodeNil(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8002")
	kademlia.Rt = NewRoutingTable(kademlia.Me)
	kademliaNet := NewNetwork(&kademlia.Me, kademlia.Rt, kademlia, NewMessageChannelManager())
	kademlia.Network = &kademliaNet
	kademlia.K = 20
	kademlia.Alpha = 3

	target := &Kademlia{}
	target.Me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:8003")
	target.Rt = NewRoutingTable(target.Me)
	targetNet := NewNetwork(&target.Me, target.Rt, target, NewMessageChannelManager())
	target.Network = &targetNet
	target.K = 20
	target.Alpha = 3

	kademlia.Rt.AddContact(target.Me)

	go kademliaNet.Listen(kademlia.Me, 8002)
	go targetNet.Listen(target.Me, 8003)

	data := []byte("test123")
	hash := Hash(data)
	testID := NewKademliaID(hash)

	output := kademlia.FindNode(testID, &target.Me)
	fmt.Println(output)
	if !(output == nil) {
		t.Errorf("TestFindNode failed..")
	}
}

func TestFindNode(t *testing.T) {
	kademlia := &Kademlia{}
	kademlia.Me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8004")
	kademlia.Rt = NewRoutingTable(kademlia.Me)
	kademliaNet := NewNetwork(&kademlia.Me, kademlia.Rt, kademlia, NewMessageChannelManager())
	kademlia.Network = &kademliaNet
	kademlia.K = 20
	kademlia.Alpha = 3

	target := &Kademlia{}
	target.Me = NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:8005")
	target.Rt = NewRoutingTable(target.Me)
	targetNet := NewNetwork(&target.Me, target.Rt, target, NewMessageChannelManager())
	target.Network = &targetNet
	target.K = 20
	target.Alpha = 3

	kademlia.Rt.AddContact(target.Me)

	go kademliaNet.Listen(kademlia.Me, 8004)
	go targetNet.Listen(target.Me, 8005)

	data := []byte("test123")
	hash := Hash(data)
	testID := NewKademliaID(hash)
	testContact := NewContact(testID, "localhost:8008")
	target.Rt.AddContact(testContact)

	output := kademlia.FindNode(testID, &target.Me)
	fmt.Println(output)
	if output == nil {
		t.Errorf("TestFindNode failed..")
	}
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

func TestPing(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8100)
	target := SetupKademliaNode("FFFFFFFF00000000000000000000000000000001", 8200)
	kademlia.Rt.AddContact(target.Me)

	expected := "ping response from ffffffff00000000000000000000000000000001"

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
		t.Errorf("Ping failure")
	}
}

func TestPingNoResponse(t *testing.T) {
	kademlia := SetupKademliaNode("FFFFFFFF00000000000000000000000000000000", 8111)
	deadNode := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000099"), "localhost:8112")

	expected := "Request sent to " + deadNode.ID.String() + " timed out"

	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	kademlia.Ping(&deadNode)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	output := BytesToString(out)
	output = strings.TrimSuffix(output, "\n")
	//fmt.Println(output)
	if !(output == expected) {
		t.Errorf("Ping Node Dead failure")
	}
}

func BytesToString(data []byte) string {
	return string(data[:])
}
