package d7024e

import (
	fmt "fmt"
	"net"
	"sync"

	pb "../protobuf"
	proto "github.com/golang/protobuf/proto"
)

type Network struct {
	rt     *RoutingTable
	me     *Contact
	mtx    *sync.Mutex
	target *KademliaID
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
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "ping"})
	sendMsg(contact.Address, message)
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "find_node", network.target.String()})
	sendMsg(contact.Address, message)
}

func (network *Network) SendFindDataMessage(hash string, contact *Contact) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "find_val", hash})
	sendMsg(contact.Address, message)
}

func (network *Network) SendStoreMessage(value string, key *KademliaID, contact *Contact) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "store", key.String(), value})
	sendMsg(contact.Address, message)
}

//Message-delarna kan kanske flyttas till egen fil?
func handleMsg(channel chan []byte, me *Contact, network *Network) {
	data := <-channel
	message := &pb.KMessage{}
	err := proto.Unmarshal(data, message)
	if err != nil {
		fmt.Println(err)
	}
	//update RoutingTable here?
	switch message.GetMsgType() {
	case "ping":
		response := buildMsg([]string{me.ID.String(), me.Address, "pong"})
		sendMsg(message.GetSndrAddress(), response)
	case "pong":

	case "find_node":
		//TODO
	case "find_node_response":
		//TODO
	case "find_val":
		//TODO
	case "find_val_response":
		//TODO
	case "store":
		//TODO

	default:
		fmt.Println("Error in handleMsg switch")

	}
}

func buildMsg(input []string) *pb.KMessage {
	if input[2] == "ping" || input[2] == "pong" {
		msg := &pb.KMessage{
			SndrID:      proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			MsgType:     proto.String(input[2]),
		}
		return msg
	}

	if input[2] == "find_node" {
		msg := &pb.KMessage{
			SndrID:      proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			MsgType:     proto.String(input[2]),
			RcvrID:      proto.String(input[3]),
		}
		return msg
	}

	if input[2] == "find_val" {
		msg := &pb.KMessage{
			SndrID:      proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			MsgType:     proto.String(input[2]),
			Key:         proto.String(input[3]),
		}
		return msg
	}

	if input[2] == "find_node_response" || input[2] == "find_val_response" {
		msg := &pb.KMessage{
			SndrID:      proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			MsgType:     proto.String(input[2]),
			Data:        []byte(input[3]),
		}
		return msg
	}

	if input[2] == "store" {
		msg := &pb.KMessage{
			SndrID:      proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			MsgType:     proto.String(input[2]),
			Key:         proto.String(input[3]),
			data:        []byte(input[4]),
		}
		return msg
	} else {
		msg := &pb.KMessage{
			SndrID:      proto.String(input[0]),
			SndrAddress: proto.String(input[1]),
			SndrID:      proto.String("Error, no valid message"),
		}
		return msg
	}
}

func sendMsg(address string, msg *pb.KMessage) {
	data, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("Marshalling error: ", err, " in sendMsg")
	}
	Conn, err := net.Dial("udp", address)
	if err != nil {
		fmt.Println("UDP error: ", err, " in sendMsg")
	}
	defer Conn.Close()
	_, err = Conn.Write(data)
	if err != nil {
		fmt.Println("Writing error: ", err, " in sendMsg")
	}
}
