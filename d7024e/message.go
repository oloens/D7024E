package d7024e

import (
	fmt "fmt"
	"sync"

	pb "../protobuf"
	proto "github.com/golang/protobuf/proto"
)

type Message struct {
	network *Network //network
	mutex   *sync.Mutex
}

func (this *Message) handleMsg(channel chan []byte, me *Contact, network *Network) {
	data := <-channel
	message := &pb.KMessage{}
	err := proto.Unmarshal(data, message)
	if err != nil {
		fmt.Println(err)
	}
	sndr := NewContact(NewKademliaID(message.GetSndrID()), message.GetSndrAddress())
	this.network //UPDATE routingtable for sender, need some kind of UpdateRoutingTable(sender) func

	switch message.GetMsgType() {
	//Cases
	}
}

func buildMsg(input []string) *pb.KMessage {
	return
}
