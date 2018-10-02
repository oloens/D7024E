package d7024e

import (
	"crypto/sha1"
	"encoding/hex"
	"time"
	"fmt"
	"sync"
	//"strconv"
)

type Kademlia struct {
	files []File
	Me	Contact
	Network *Network
	Rt 	*RoutingTable
	Alpha 	int
	K	int
	Mtx 	*sync.Mutex
	RtMtx 	*sync.Mutex
}
const tRepublish = 30
const tExpire = 35
const ttl = 5 * time.Second

type File struct {
	Hash	string
	Value	[]byte
	Pin	bool
	TimeSinceRepublish int
}
func (kademlia Kademlia) GetFiles() []File {
	return kademlia.files
}
func (kademlia *Kademlia) AddFile(file File) {
	kademlia.files = append(kademlia.files, file)
}
func (kademlia *Kademlia) IterativeLookup(iterateType string, target *KademliaID) (contactList []Contact, closestNode *Contact, value []byte) {
	shortList := &ContactCandidates{}
	kademlia.RtMtx.Lock()
	shortList.Append(kademlia.Rt.FindClosestContacts(target, kademlia.Alpha))
	kademlia.RtMtx.Unlock()
	//removeFromShortlist := &ContactCandidates{}
	var channels []chan []string
	queriedNodes := make(map[string]bool)
	alpha := kademlia.Alpha
	var closest *Contact
	var initialized bool
	if shortList.Len() != 0 {
		closest = &shortList.GetContacts(1)[0]
		initialized = true
	}
	probed := 0
	closestThisRound := closest
	for {
		queryCandidates := shortList.GetContacts(shortList.Len())
		current := 0
		for i := 0; current < alpha && i <shortList.Len(); i++ {
			if queriedNodes[queryCandidates[i].ID.String()] {
				continue
			}
			queriedNodes[queryCandidates[i].ID.String()] = true
			ch := make(chan []string)
			channels = append(channels, ch)
			switch iterateType {
			case "FIND_CONTACT":
				go kademlia.LookupMessage("FIND_CONTACT", target, &queryCandidates[i], ch)
			case "FIND_VALUE": 
				go kademlia.LookupMessage("FIND_VALUE", target, &queryCandidates[i], ch)

			}
			current++
			fmt.Println("Started query for " + queryCandidates[i].ID.String())
		}
		if current == 0 {
			shortList.Sort()
			fmt.Println("Total nodes probed for this lookup: ", probed, "escaped because current==0")
			return shortList.GetContacts(shortList.Len()), closest, nil
		}
		for current > 0 {
			//i := current - 1
			//rpcType, data := <- channels[i-1], <- channels[i-1]
			rpcType, data := <- channels[0], <- channels[0]
			switch rpcType[0] {
			case "FIND_CONTACT":
				for _, cont := range data {
					if cont != kademlia.Me.ID.String() {
						cont_restored := RestoreContact(cont)
						cont_restored.CalcDistance(target)
						if !shortList.Exists(&cont_restored) {
							shortList.Append([]Contact{cont_restored})
						}
						if initialized {
							if cont_restored.Less(closest) {
								closest = &cont_restored
							}
						} else {
							closest = &cont_restored
						}
					}
				}
			case "FIND_VALUE":
				if rpcType[1] == "DATA" {
					return nil, nil, []byte(data[0])
				}
				for _, cont := range data {
                                        if cont != kademlia.Me.ID.String() {
                                                cont_restored := RestoreContact(cont)
                                                cont_restored.CalcDistance(target)
						if !shortList.Exists(&cont_restored) { 
							shortList.Append([]Contact{cont_restored}) 
						}
                                                if initialized {
                                                        if cont_restored.Less(closest) {
                                                                closest = &cont_restored
                                                        }
                                                } else {
                                                        closest = &cont_restored
                                                }
                                        }

				}
			case "TIMEOUT":
				fmt.Println("Node " + data[0] + " inactive, removing from shortlist")
				shortList.Remove(data[0])
			}
			close(channels[0])
			//channels = append(channels[:i], channels[i+1]...)
			channels = channels[1:]
			current--
			probed++

	}
	if probed >= kademlia.K {
		shortList.Sort() // TODO sort by distance
		fmt.Println("Total nodes probed for this lookup: ", probed, "escaped because already probed 20")
		return shortList.GetContacts(kademlia.K), closest, nil

	}
	if closestThisRound == closest {
		fmt.Println("Did not find any closer this round, sending out to k closest unqueried nodes")
		kademlia.RtMtx.Lock()
		cts := kademlia.Rt.FindClosestContacts(target, kademlia.K)
		kademlia.RtMtx.Unlock()
		for i := 0; i<len(cts); i++ {
			if queriedNodes[cts[i].ID.String()] {
				continue
			}
			queriedNodes[cts[i].ID.String()] = true
			ch := make(chan []string)
			channels = append(channels, ch)
			switch iterateType {
			case "FIND_CONTACT":
				go kademlia.LookupMessage("FIND_CONTACT", target, &cts[i], ch)
			case "FIND_VALUE": 
				go kademlia.LookupMessage("FIND_VALUE", target, &cts[i], ch)
			}
			fmt.Println("Started query for " + cts[i].ID.String())
			probed++
		}
		for _, ch := range channels {
			rpcType, data := <- ch, <- ch
			switch rpcType[0] {
			case "FIND_CONTACT":
				for _, cont := range data {
                                        if cont != kademlia.Me.ID.String() {
                                                cont_restored := RestoreContact(cont)
                                                cont_restored.CalcDistance(target)
                                                if !shortList.Exists(&cont_restored) {
                                                        shortList.Append([]Contact{cont_restored})
                                                }
                                        }

				}
			case "FIND_VALUE": 
				if rpcType[1] == "DATA" {
					return nil, nil, []byte(data[0])
				}
				for _, cont := range data {
                                        if cont != kademlia.Me.ID.String() {
                                                cont_restored := RestoreContact(cont)
                                                cont_restored.CalcDistance(target)
						if !shortList.Exists(&cont_restored) { 
							shortList.Append([]Contact{cont_restored}) 
						}
                                                if initialized {
                                                        if cont_restored.Less(closest) {
                                                                closest = &cont_restored
                                                        }
                                                } else {
                                                        closest = &cont_restored
                                                }
                                        }

				}
			case "TIMEOUT":
				fmt.Println("Node " + data[0] + " inactive, removing from shortlist..")
				shortList.Remove(data[0])
			}
		close(ch)
	}
	
		fmt.Println("Total nodes probed for this lookup: ", probed)
		shortList.Sort()
		if (shortList.Len()<kademlia.K) {
			return shortList.GetContacts(shortList.Len()), closest, nil
		} else {
			return shortList.GetContacts(kademlia.K), closest, nil
		}

	}


    }
    fmt.Println("lookup seems to have failed to find a single contact")
    return nil, nil, nil
}

func (kademlia *Kademlia) LookupMessage(rpctype string, target *KademliaID, contact *Contact, ch chan []string) {
	switch rpctype {
	case "FIND_CONTACT":
		result := kademlia.FindNode(target, contact)
		if result[0] == "TIMEOUT" {
			ch <- result
			ch <- []string{contact.ID.String()}
			return
		}
		ch <- []string{rpctype}
		ch <- result
		return
	case "FIND_VALUE":
		cts, data := kademlia.FindValue(target, contact)
		if data != nil {
			ch <- []string{rpctype, "DATA"}
			ch <- []string{string(data[:])}
			return
		} else {
		        if cts[0] == "TIMEOUT" {
                                ch <- cts
                                ch <- []string{contact.ID.String()}
                                return
                        }

			ch <- []string{rpctype, "CONTACTS"}
			ch <- cts
			return
		}
	}
	fmt.Println("LookupMessage failed")
	return
}

func (kademlia *Kademlia) FindNode(target *KademliaID, contact *Contact) []string {
	id := NewRandomKademliaID()
	msgchan := NewMessageChannel(id)
	kademlia.Network.Mgr.AddMessageChannel(msgchan)
	kademlia.Network.SendFindContactMessage(target, contact, id)
	select {
	case response := <-msgchan.Channel:
		close(msgchan.Channel)
		kademlia.Network.Mgr.RemoveMessageChannel(id)
		kademlia.RtMtx.Lock()
		kademlia.Rt.AddContact(*contact)
		kademlia.RtMtx.Unlock()
		return response.GetContacts()
	case <-time.After(ttl):
		fmt.Println("Request sent to " + contact.ID.String() + " timed out")
		close(msgchan.Channel)
		kademlia.Network.Mgr.RemoveMessageChannel(id)
		return []string{"TIMEOUT"}
	}
}
func (kademlia *Kademlia) FindValue(value *KademliaID, contact *Contact) ([]string, []byte) {
        id := NewRandomKademliaID()
        msgchan := NewMessageChannel(id)
        kademlia.Network.Mgr.AddMessageChannel(msgchan)
        kademlia.Network.SendFindDataMessage(value.String(), contact, id)
        select {
	case response := <-msgchan.Channel:
		if response.GetData() != nil {
			return nil, response.GetData()
		}
		close(msgchan.Channel)
		kademlia.Network.Mgr.RemoveMessageChannel(id)
		kademlia.RtMtx.Lock()
		kademlia.Rt.AddContact(*contact)
		kademlia.RtMtx.Unlock()
        	return response.GetContacts(), nil
	case <-time.After(ttl):
		fmt.Println("Request sent to " + contact.ID.String() + " timed out")
		close(msgchan.Channel)
		kademlia.Network.Mgr.RemoveMessageChannel(id)
		return []string{"TIMEOUT"}, nil
	}
}
func (kademlia *Kademlia) SendStore(hash string, value []byte) {
	kclosest, _, _ := kademlia.IterativeLookup("FIND_CONTACT", NewKademliaID(hash))
	for _, contact := range kclosest {
		kademlia.Network.SendStoreMessage(string(value[:]), NewKademliaID(hash), &contact)
		if contact.ID.String() != kademlia.Me.ID.String() {
			//kademlia.Rt.AddContact(contact) //when do we update routing table
		}
	}
	go kademlia.Republish(hash, value)
}
func (kademlia *Kademlia) Republish(hash string, value []byte) {
	time.Sleep(tRepublish * time.Second)
	fmt.Println("Republishing file with hash: " + hash)
	kademlia.SendStore(hash, value)
}
func (kademlia *Kademlia) Ping(contact *Contact) {
	id := NewRandomKademliaID()
        msgchan := NewMessageChannel(id)
        kademlia.Network.Mgr.AddMessageChannel(msgchan)
        kademlia.Network.SendPingMessage(contact, id)
        select {
        case response := <-msgchan.Channel:
                close(msgchan.Channel)
                kademlia.Network.Mgr.RemoveMessageChannel(id)
		kademlia.RtMtx.Lock()
                kademlia.Rt.AddContact(*contact)
		kademlia.RtMtx.Unlock()
		fmt.Println("ping response from " + response.GetSndrID())
                //return response.GetContacts()

        case <-time.After(ttl):
                fmt.Println("Request sent to " + contact.ID.String() + " timed out")
		close(msgchan.Channel)
		kademlia.Network.Mgr.RemoveMessageChannel(id)
                //return []string{"TIMEOUT"}
        }

}
func (kademlia *Kademlia) Bootstrap() bool {
	kclosest, _, _ := kademlia.IterativeLookup("FIND_CONTACT", kademlia.Me.ID)
	if len(kclosest) == 0 {
		return false
	}

	fmt.Println("Bootstrap complete! k closest:")
	for _, contact := range kclosest {
		if contact.ID.String() != kademlia.Me.ID.String() {
			//kademlia.Rt.AddContact(contact)
		}
		fmt.Println(contact.String())
	}
	return true

}
func (kademlia *Kademlia) SendFindValue(hash string) (*Contact, []byte){
	kclosest, closest, val := kademlia.IterativeLookup("FIND_VALUE", NewKademliaID(hash))
	if val != nil {
		return nil, val
	}

	for _, contact := range kclosest {
		if contact.ID.String() != kademlia.Me.ID.String() {
			//kademlia.Rt.AddContact(contact)
		}
	}
	return closest, val
}
func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}



func (kademlia *Kademlia) LookupData(hash string) []byte {
	kademlia.Mtx.Lock()
	file := kademlia.LookupFile(hash)
	kademlia.Mtx.Unlock()
	if file != nil {
		return file.Value
	}
	return nil
}
func (kademlia *Kademlia) LookupFile(hash string) *File {
	for _, file := range kademlia.files {
                if hash == file.Hash {
                        return &file
                }
        }
        return nil
}

func Hash(data []byte) string {
        hashbytes := sha1.Sum(data)
        hash := hex.EncodeToString(hashbytes[0:IDLength])
	return hash
}
func (kademlia *Kademlia) Store(data []byte) {
	hash := Hash(data)
	kademlia.Mtx.Lock()
	defer kademlia.Mtx.Unlock()
	for i, file := range kademlia.files {
		if file.Hash == hash {
			kademlia.files[i].TimeSinceRepublish = 0
			return
		}
	}

	file := File{hash, data, false, 0}
	kademlia.AddFile(file)
}
func (kademlia *Kademlia) index(hash string) int { //using the outcommended stuff in pin/unpin did not work, TODO fix race conditions here lol
	for i, file := range kademlia.files {
		if file.Hash == hash {
			return i
		}
	}
	return -1
}
func (kademlia *Kademlia) Pin(hash string) {
	kademlia.Mtx.Lock()
	defer kademlia.Mtx.Unlock()
	i := kademlia.index(hash)
	if i == -1 {
		fmt.Println("error in pin, file not found")
		return
	}
	kademlia.files[i].Pin = true
	fmt.Println("successfully pinned file with hash " + hash)

}
func (kademlia *Kademlia) Unpin(hash string) {
	kademlia.Mtx.Lock()
	defer kademlia.Mtx.Unlock()
	i := kademlia.index(hash)
        if i == -1 {
                fmt.Println("error in unpin, file not found")
                return
        }
	kademlia.files[i].Pin = false
	fmt.Println("successfully unpinned file with hash " + hash)
}

func (kademlia *Kademlia) Purge() {
	timer := time.NewTimer(5 * time.Second)
	<-timer.C
	kademlia.Mtx.Lock()
	for i, _ := range kademlia.files {
		kademlia.files[i].TimeSinceRepublish += 5
	}
	var newfiles []File
	for _, file := range kademlia.files {
		if file.Pin == true || file.TimeSinceRepublish <= tExpire {
			newfiles = append(newfiles, file)
		} else {
			fmt.Println("timer expired for file with hash: " + file.Hash + " , removing..")
		}
	}
	kademlia.files = newfiles
	kademlia.Mtx.Unlock()
	kademlia.Purge()
}

