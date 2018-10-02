package d7024e

import (
	fmt "fmt"
	"net"
	"strconv"
	"sync"
	"time"

	pb "../protobuf"

	proto "../proto"
)

type Network struct {
	rt       *RoutingTable
	me       *Contact
	mtx      *sync.Mutex
	target   *KademliaID
	kademlia *Kademlia
	Mgr      *MessageChannelManager
}

func NewNetwork(me *Contact, rt *RoutingTable, kademlia *Kademlia, mgr *MessageChannelManager) Network {
	network := Network{}
	network.me = me
	network.rt = rt
	network.mtx = &sync.Mutex{}
	network.kademlia = kademlia
	network.Mgr = mgr
	return network
}

func (network *Network) Listen(me Contact, port int) {
	address, err1 := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port)) //me.Address)
	Conn, err2 := net.ListenUDP("udp", address)
	if (err1 != nil) || (err2 != nil) {
		fmt.Println("Error listener: ", err1, " .... and : ", err2)
	}
	defer Conn.Close()
	channel := make(chan []byte)
	bufr := make([]byte, 4096)
	//fmt.Println("Listening...")
	for {
		time.Sleep(5 * time.Millisecond)
		n, _, err := Conn.ReadFromUDP(bufr)
		go handleMsg(channel, &me, network)
		channel <- bufr[:n]
		if err != nil {
			fmt.Println("Read error @Listen: ", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func (network *Network) SendPingMessage(contact *Contact, rpc_id *KademliaID) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "ping", rpc_id.String()})
	sendMsg(contact.Address, message)
}

func (network *Network) SendFindContactMessage(id *KademliaID, contact *Contact, rpc_id *KademliaID) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "find_node", contact.ID.String(), id.String(), rpc_id.String()})
	sendMsg(contact.Address, message)
}

func (network *Network) SendFindDataMessage(hash string, contact *Contact, rpc_id *KademliaID) {
	message := buildMsg([]string{network.me.ID.String(), network.me.Address, "find_val", hash, rpc_id.String()})
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
		response := buildMsg([]string{me.ID.String(), me.Address, "pong", message.GetRpcID()})
		sender := NewContact(NewKademliaID(message.GetSndrID()), message.GetSndrAddress())
		network.kademlia.RtMtx.Lock()
		network.rt.AddContact(sender)
		network.kademlia.RtMtx.Unlock()
		sendMsg(message.GetSndrAddress(), response)
	case "pong":
		rpc_id := NewKademliaID(message.GetRpcID())
		msgchan := network.Mgr.GetMessageChannel(rpc_id)
		msgchan.Channel <- message
	case "find_node":
		targetKey := message.GetKey()
		target := NewKademliaID(targetKey)
		network.kademlia.RtMtx.Lock()
		contacts := network.rt.FindClosestContacts(target, 20)
		sender := NewContact(NewKademliaID(message.GetSndrID()), message.GetSndrAddress())
		sender.CalcDistance(me.ID)
		network.rt.AddContact(sender)
		network.kademlia.RtMtx.Unlock()

		//me_with_dist := *me
		//me_with_dist.CalcDistance(target)
		//if len(contacts)<20 {
		//	contacts = append(contacts, me_with_dist)
		//}
		var contacts_string []string
		for _, ct := range contacts {
			contacts_string = append(contacts_string, ct.String())
		}
		response := buildMsgWithArray([]string{me.ID.String(), me.Address, "find_node_response", message.GetRpcID()}, contacts_string)
		sendMsg(message.GetSndrAddress(), response)
	case "find_node_response":
		rpc_id := NewKademliaID(message.GetRpcID())
		msgchan := network.Mgr.GetMessageChannel(rpc_id)
		msgchan.Channel <- message
	case "find_val":
		targetKey := message.GetKey()
		result := network.kademlia.LookupData(targetKey)
		if result != nil {
			response := buildMsg([]string{me.ID.String(), me.Address, "find_val_response", string(result[:]), message.GetRpcID()})
			sendMsg(message.GetSndrAddress(), response)
			sender := NewContact(NewKademliaID(message.GetSndrID()), message.GetSndrAddress())
			network.kademlia.RtMtx.Lock()
			network.rt.AddContact(sender)
			network.kademlia.RtMtx.Unlock()

		} else {
			target := NewKademliaID(targetKey)
			network.kademlia.RtMtx.Lock()
			contacts := network.rt.FindClosestContacts(target, 20)
			sender := NewContact(NewKademliaID(message.GetSndrID()), message.GetSndrAddress())
			sender.CalcDistance(me.ID)
			network.rt.AddContact(sender)
			network.kademlia.RtMtx.Unlock()

			//me_with_dist := *me
			//me_with_dist.CalcDistance(target)
			//if len(contacts)<20 {
			//	contacts = append(contacts, me_with_dist)
			//}
			var contacts_string []string
			for _, ct := range contacts {
				contacts_string = append(contacts_string, ct.String())
			}
			response := buildMsgWithArray([]string{me.ID.String(), me.Address, "find_val_response", message.GetRpcID()}, contacts_string)
			sendMsg(message.GetSndrAddress(), response)

		}

	case "find_val_response":
		msgchan := network.Mgr.GetMessageChannel(NewKademliaID(message.GetRpcID()))
		msgchan.Channel <- message
	case "store":
		data := message.GetData()
		key := message.GetKey()
		network.kademlia.Store(data)
		sender := NewContact(NewKademliaID(message.GetSndrID()), message.GetSndrAddress())
		network.kademlia.RtMtx.Lock()
		network.rt.AddContact(sender)
		network.kademlia.RtMtx.Unlock()

		fmt.Println("Store RPC received, storing file with hash: " + key)

	default:
		fmt.Println("Error in handleMsg switch")

	}
}

func buildMsgWithArray(input []string, contacts []string) *pb.KMessage {
	msg := &pb.KMessage{
		SndrID:      input[0],
		SndrAddress: input[1],
		MsgType:     input[2],
		RpcID:       input[3],
		Contacts:    contacts,
	}
	return msg
}
func buildMsg(input []string) *pb.KMessage {
	if input[2] == "ping" || input[2] == "pong" {
		msg := &pb.KMessage{
			SndrID:      input[0], //proto.String(input[0]),
			SndrAddress: input[1], //proto.String(input[1]),
			MsgType:     input[2], //proto.String(input[2]),
			RpcID:       input[3],
		}
		return msg
	}

	if input[2] == "find_node" {
		msg := &pb.KMessage{
			SndrID:      input[0], //proto.String(input[0]),
			SndrAddress: input[1], //proto.String(input[1]),
			MsgType:     input[2], //proto.String(input[2]),
			RcvrID:      input[3], //proto.String(input[3]),
			Key:         input[4],
			RpcID:       input[5],
		}
		return msg
	}

	if input[2] == "find_val" {
		msg := &pb.KMessage{
			SndrID:      input[0], //proto.String(input[0]),
			SndrAddress: input[1], //proto.String(input[1]),
			MsgType:     input[2], //proto.String(input[2]),
			Key:         input[3], //proto.String(input[3]),
			RpcID:       input[4],
		}
		return msg
	}

	if input[2] == "find_node_response" || input[2] == "find_val_response" {
		msg := &pb.KMessage{
			SndrID:      input[0], //proto.String(input[0]),
			SndrAddress: input[1], //proto.String(input[1]),
			MsgType:     input[2], //proto.String(input[2]),
			Data:        []byte(input[3]),
			RpcID:       input[4],
		}
		return msg
	}

	if input[2] == "store" {
		msg := &pb.KMessage{
			SndrID:      input[0], //proto.String(input[0]),
			SndrAddress: input[1], //proto.String(input[1]),
			MsgType:     input[2], //proto.String(input[2]),
			Key:         input[3], //proto.String(input[3]),
			Data:        []byte(input[4]),
		}
		return msg
	} else {
		msg := &pb.KMessage{
			SndrID:      input[0],                  //proto.String(input[0]),
			SndrAddress: input[1],                  //proto.String(input[1]),
			MsgType:     "Error, no valid message", //proto.String("Error, no valid message"),
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
