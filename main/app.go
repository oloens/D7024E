package main
import (
	"time"
	"fmt"
	"d7024e"
	"net"
	pb "protobuf"
	proto "proto"
	"strconv"
	"bufio"
	"os"
	"strings"
	"math/rand"
	"sync"
)


//just testing deployment of go in a docker container, and subsequently
// deploying a cluster using docker swarm
func main() {
	


	rawip := getIP()
	ip := rawip+":8000"
	split_ip := strings.Split(rawip, ".")
	seed, _ := strconv.Atoi(split_ip[3])
        //tarip := "172.17.0.2:8000" // for normal docker run
	tarip := "10.0.0.4:8000" // for docker swarm
	var me d7024e.Contact
	if tarip == ip { // base bootstrap node
                me = d7024e.NewContact(d7024e.NewKademliaID("8d92ca43f193dee47f591549f597a811c8fa67ab"), ip)
        } else {
		for i := 0; i < seed; i++ {
	    		rand.Seed(time.Now().UTC().UnixNano())
		}
		id := d7024e.NewRandomKademliaID()
		me = d7024e.NewContact(id, ip)
	}
	fmt.Println("my id is :" + me.ID.String())

	rt := d7024e.NewRoutingTable(me)

	//some initial stuff stored
	kademlia := &d7024e.Kademlia{}
	kademlia.Mtx = &sync.Mutex{}
	val := []byte("testvalueasdasdtjenatjena")
	val2 := []byte("testvalue2asdasdtjenatjena")
	kademlia.Store(val)
	kademlia.Store(val2)
	hash2 := d7024e.Hash(val2)
	kademlia.Pin(hash2)
	go kademlia.Purge()

	
	net := d7024e.NewNetwork(&me, rt, kademlia, d7024e.NewMessageChannelManager())
	kademlia.Network = &net
	kademlia.Rt = rt
	kademlia.K = 20
	kademlia.Alpha = 3
	kademlia.Me = me
	go net.Listen(me, 8000)
	if ip != tarip {
		kademlia.Rt.AddContact(d7024e.NewContact(d7024e.NewKademliaID("8d92ca43f193dee47f591549f597a811c8fa67ab"), tarip))
		success := kademlia.Bootstrap()
		for !success {
			fmt.Println("bootstrap node was inactive, trying again in 5 seconds...")
			time.Sleep(5 * time.Second)
			success = kademlia.Bootstrap()
		}
	
	}
	
	reader := bufio.NewReader(os.Stdin)
	for {
		   command, _ := reader.ReadString('\n')
		   command = strings.Replace(command, "\n", "", -1)
		   split := strings.Split(command, " ")
		   switch split[0] {
		   case "find":
			   fmt.Println("trying to find file with data: " + split[1])
			   closest, data := kademlia.SendFindValue(d7024e.Hash([]byte(split[1])))
			   if data != nil {
				   fmt.Println("data found! data is : " + string(data[:]))

			   } else {
				   fmt.Println("data not found, closest node to hash is: " + closest.ID.String())
			   }
		   case "info": 
		   	   fmt.Println("Node with ID: ", kademlia.Me.ID.String())
			   fmt.Println("The 20 closest contacts in my routing table are: ")
			   for _, ct := range kademlia.Rt.FindClosestContacts(kademlia.Me.ID, 20) {
				   fmt.Println(ct.ID.String())
			   }
		   case "store":
			   fmt.Println("executing store command on string: " + split[1])
			   value := []byte(split[1])
			   hash := d7024e.Hash(value)
			   //val_string := string(value[:])
			   //key := d7024e.NewKademliaID(hash)

			   kademlia.SendStore(hash, value)
			   //kademlia.Store(value)
			   //if split[len(split)-1] == "--pin" {
			//	kademlia.Pin(d7024e.Hash(value))
			//}
		   case "files":
		   	   fmt.Println("showing all file values")
		   	   for i, file := range kademlia.GetFiles() {
				   str := strconv.Itoa(i) + ": " + string(file.Value[:]) + " *TIME_SINCE_REPUBLISH: " + strconv.Itoa(file.TimeSinceRepublish) + "s*"
				   if file.Pin {
					   str = str + "   *PINNED* "
				   }
				   fmt.Println(str)


			   }
		   default:
			   fmt.Println("invalid command or not yet implemented")
		   }

}
	
}
func getIP() string {

	iface, _ := net.InterfaceByName("eth0")
        addrs, _ := iface.Addrs()
        for _, addr := range addrs {
                var ip net.IP
                switch v := addr.(type) {
                        case *net.IPNet:
                                ip = v.IP
                        case *net.IPAddr:
                                ip = v.IP
                        }
                return ip.String()
        }
	return "error"
    }
func buildMessage() *pb.KMessage {
	t1 := "testaddress"
	msg := &pb.KMessage{
		SndrAddress: t1,//proto.String(t1),
	}
	return msg
}
func tempListen(c chan *pb.KMessage) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":8000")
	if err != nil {
		fmt.Println("error on resolveudpaddr")
	}
	udpconn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error on listenUDP")
	}
	fmt.Println("listening")
	for {
		//time.Sleep(500 * time.Millisecond)
		 receiveUDP(udpconn, c)
	}


}
func receiveUDP(conn net.Conn, c chan *pb.KMessage) {
	//fmt.Println("here")
	//defer conn.Close() //close connection upon function termination
	data := make([]byte, 2048)
	n, err := conn.Read(data)
	//fmt.Println("here2")
	if err != nil {
		fmt.Println("Error reading data")
	}
	pbdata := new(pb.KMessage)
	//fmt.Println("here3")
	proto.Unmarshal(data[0:n], pbdata)
	//fmt.Println("msg received:", pbdata.GetSndrAddress())
	//temporary print
	c <- pbdata
	fmt.Println("msg received:", pbdata.GetSndrAddress())
}
func sendUDP(msg *pb.KMessage, target *d7024e.Contact) {
	ip := target.Address
	//fmt.Println(ip)
	conn, err := net.Dial("udp", ip)
	if err != nil {
		fmt.Println("error dialing for send")
		return
	}
	//TODO marshal msg into data
	data, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("error marshaling msg")
		return
	}
	conn.Write(data)
	///conn.Close()

}
