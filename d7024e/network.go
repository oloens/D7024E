package d7024e

import (
	fmt "fmt"
	"net"
	"sync"

	pb "../protobuf"
	proto "github.com/golang/protobuf/proto"
)

type Network struct {
	rt  *RoutingTable
	me  *Contact
	mtx *sync.Mutex
}

func NewNetwork(me *Contact, rt *RoutingTable) Network {
	network := Network{}
	network.me = me
	network.rt = rt
	network.mtx = &sync.Mutex{}
	return network
}

func Listen(ip string, port int) {
	// TODO
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}

//Message-delarna kan kanske flyttas till egen fil?
func handleMsg(channel chan []byte, addr *net.UDPAddr) {
	data := <-channel
	message := &pb.KMessage{}
	err := proto.Unmarshal(data, message)
	if err != nil {
		fmt.Println(err)
	}
	switch message.GetMsgType() {
	case "ping":
		//TODO
	case "find_node":
		//TODO
	case "find_val":
		//TODO
	case "store":
		//TODO

	default:
		fmt.Println("Error in handleMsg switch")

	}
}

func buildMsg(input []string) *pb.KMessage {
	if input[0] == "ping" || input[0] == "pong" {
		msg := &pb.KMessage{
			MsgType:     proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			SndrID:      proto.String(input[2]),
		}
		return msg
	}

	if input[0] == "find_node" {
		msg := &pb.KMessage{
			MsgType:     proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			SndrID:      proto.String(input[2]),
			RcvrID:      proto.String(input[3]),
		}
		return msg
	}

	if input[0] == "find_val" {
		msg := &pb.KMessage{
			MsgType:     proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			SndrID:      proto.String(input[2]),
			Key:         proto.String(input[3]),
		}
		return msg
	}

	if input[0] == "find_node_response" || input[0] == "find_val_response" {
		msg := &pb.KMessage{
			MsgType:     proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			SndrID:      proto.String(input[2]),
			Data:        []byte(input[3]),
		}
		return msg
	}

	if input[0] == "store" {
		msg := &pb.KMessage{
			MsgType:     proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			SndrID:      proto.String(input[2]),
			Key:         proto.String(input[3]),
			Val:         proto.String(input[4]),
		}
		return msg
	} else {
		msg := &pb.KMessage{
			MsgType:     proto.String("Error, no valid message"),
			SndrAddress: proto.String(input[1]),
			SndrID:      proto.String(input[2]),
		}
		return msg
	}
}
