package main
import (
	"fmt"
	"bufio"
	"strings"
	"os"
	"io"
	"net/http"
	"mime/multipart"
	"bytes"
	"net/url"
	"encoding/json"
)
type ResponseData struct {
        Value        []byte   `json:"value,omitempty"`
        Key          string   `json:"key,omitempty"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	url := "http://localhost:8001/"
	for {
		command, _ := reader.ReadString('\n')
		command = strings.Replace(command, "\n", "", -1)
		split := strings.Split(command, " ")
		if split[0] != "dfs" {
			fmt.Println("Syntax: dfs store <filename> or dfs cat/pin/unpin <hash>")
		}
		switch split[1] {
		case "store":
			store(split[2], url)
		case "cat":
			cat(split[2], url)
		case "pin":
			pin(split[2], url)
		case "unpin":
			unpin(split[2], url)
		default:
			fmt.Println("Syntax: dfs store <filename> or dfs cat/pin/unpin <hash>")
		}
	}
}

func store(name, tar string) error {
	buf := &bytes.Buffer{}
        writer := multipart.NewWriter(buf)
        fileWriter, err := writer.CreateFormFile("file", name)
        if err != nil {
		//fmt.Println("error writing to buffer")
                return err
        }
	fh, err := os.Open(name)
        if err != nil {
               // fmt.Println("error opening file")
                return err
        }
        defer fh.Close()
        io.Copy(fileWriter, fh)
        contentType := writer.FormDataContentType()
        writer.Close()
        resp, err2 := http.Post(tar+"store", contentType, buf)
        if err2 != nil {
                fmt.Println("error in http.Post")
                fmt.Println(err2)
	}
	defer resp.Body.Close()
	var data ResponseData
	json.NewDecoder(resp.Body).Decode(&data)
	fmt.Println(data.Key)
	return nil

}
func cat(hash, tar string) error {
	resp, err := http.PostForm(tar+"cat", url.Values{"key": {hash}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var data ResponseData
	json.NewDecoder(resp.Body).Decode(&data)
	fmt.Println(string(data.Value[:]))
	return nil

}
func pin(hash, tar string) error {
        resp, err := http.PostForm(tar+"pin", url.Values{"key": {hash}})
        if err != nil {
                return err
        }
        defer resp.Body.Close()
        var data ResponseData
        json.NewDecoder(resp.Body).Decode(&data)
        fmt.Println("Pinned ", data.Key, " on max 20 nodes")
        return nil

}
func unpin(hash, tar string) error {
        resp, err := http.PostForm(tar+"unpin", url.Values{"key": {hash}})
        if err != nil {
                return err
        }
        defer resp.Body.Close()
        var data ResponseData
        json.NewDecoder(resp.Body).Decode(&data)
        fmt.Println("Unpinned ", data.Key, " on max 20 nodes")
        return nil

}
