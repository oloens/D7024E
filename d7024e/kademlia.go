package d7024e

import (
	"crypto/sha1"
	"encoding/hex"
	"time"
	"fmt"
	//"strconv"
)

type Kademlia struct {
	files []File
}
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

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) []byte {
	file := kademlia.LookupFile(hash)
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
	//var file *File
	//file = kademlia.LookupFile(hash)
	//fmt.Println("2")
	//fmt.Println(file)
	//if file == nil {
	//	return
	//}
	//file.Pin = true
	i := kademlia.index(hash)
	if i == -1 {
		fmt.Println("error in pin, file not found")
		return
	}
	kademlia.files[i].Pin = true
	fmt.Println("successfully pinned file with hash " + hash)

}
func (kademlia *Kademlia) Unpin(hash string) {
	//file := kademlia.LookupFile(hash)
	//if file == nil {
	//	return
	//}
	//file.Pin = false
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
	for i, _ := range kademlia.files {
		kademlia.files[i].TimeSinceRepublish += 5
	}
	var newfiles []File
	for _, file := range kademlia.files {
		if file.Pin == true || file.TimeSinceRepublish <= 30 {
			newfiles = append(newfiles, file)
		} else {
			fmt.Println("timer expired for file with hash: " + file.Hash + " , removing..")
		}
	}
	kademlia.files = newfiles
	kademlia.Purge()
}

