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
)


//just testing deployment of go in a docker container, and subsequently
// deploying a cluster using docker swarm
func main() {
	


	rawip := getIP()
	ip := rawip+":8000"
	seed := rawip[len(rawip)-1:]
	id := d7024e.NewRandomKademliaID()
	seed_int, _ := strconv.Atoi(seed)
	for i := 0; i < seed_int; i++ {
	    id = d7024e.NewRandomKademliaID()
	}
	
	me := d7024e.NewContact(id, ip)
//	strContact := me.String()


	rt := d7024e.NewRoutingTable(me)
	//rt.AddContact(me)

	//testing store
	kademlia := &d7024e.Kademlia{}
	val := []byte("testvalueasdasdtjenatjena")
	val2 := []byte("testvalue2asdasdtjenatjena")
	kademlia.Store(val)
	kademlia.Store(val2)
	//hash := d7024e.Hash(val)
	hash2 := d7024e.Hash(val2)
	kademlia.Pin(hash2)
	go kademlia.Purge()
	//result := kademlia.LookupData(hash)
	//fmt.Println(string(result[:]))
	//fmt.Println(string(hash[:]))
	//fmt.Println(me.ID.String())

	
	net := d7024e.NewNetwork(&me, rt, kademlia)
	go net.Listen(me, 8000)
	tarip := "172.17.0.2:8000"
	//tarip := "10.0.0.2:8000"
	tar := d7024e.NewContact(d7024e.NewRandomKademliaID(), tarip)
	if ip != tarip {
		//tar := d7024e.NewContact(d7024e.NewRandomKademliaID(), tarip)
		time.Sleep(1000 * time.Millisecond)
		net.SendFindContactMessage(&tar)
		//net.SendFindDataMessage(hash, &tar)
		//fmt.Println("sent ping msg, sleeping...")
	
	}
	
	reader := bufio.NewReader(os.Stdin)
	for {
		   command, _ := reader.ReadString('\n')
		   command = strings.Replace(command, "\n", "", -1)
		   split := strings.Split(command, " ")
		   switch split[0] {
		   case "store":
			   fmt.Println("executing store command on string: " + split[1])
			   value := []byte(split[1])
			   hash := d7024e.Hash(value)
			   val_string := string(value[:])
			   key := d7024e.NewKademliaID(hash)

			   net.SendStoreMessage(val_string, key, &tar)
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
