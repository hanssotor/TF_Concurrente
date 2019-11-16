package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type MessageType int32

const (
	NEW_ACADEMIC_RECORD                = 1
	UPDATE_ACADEMIC_RECORD             = 2
	LIST_ACADEMIC_RECORD               = 3
	LIST_HOSTS                         = 4
	PROTOCOL                           = "tcp"
	NEW_HOST               MessageType = 0
	ADD_HOST               MessageType = 1
	ADD_BLOCK              MessageType = 2
	NEW_BLOCK              MessageType = 3
	SET_BLOCKS             MessageType = 4
)

/* [SERVER] END */

var localHost string
var hosts []string

type MessageBody struct {
	MessageType MessageType
	Message     string
}

func GetMessage(conn net.Conn) string {
	reader := bufio.NewReader(conn)
	data, _ := reader.ReadString('\n')
	return strings.TrimSpace(data)
}

func SendMessage(toHost string, message string) {
	conn, _ := net.Dial(PROTOCOL, toHost)
	defer conn.Close()
	fmt.Fprintln(conn, message)
}

func SendMessageWithReply(toHost string, message string) string {
	conn, _ := net.Dial(PROTOCOL, toHost)
	defer conn.Close()
	fmt.Fprintln(conn, message)
	return GetMessage(conn)
}

func RemoveHost(index int, hosts []string) []string {
	n := len(hosts)
	hosts[index] = hosts[n-1]
	hosts[n-1] = ""
	return hosts[:n-1]
}

func RemoveHostByValue(ip string, hosts []string) []string {
	for index, host := range hosts {
		if host == ip {
			return RemoveHost(index, hosts)
		}
	}
	return hosts
}

func Broadcast(newHost string) {
	for _, host := range hosts {
		sendData := append(hosts, newHost, localHost)
		sendData = RemoveHostByValue(host, sendData)
		sendBody := MessageBody{
			MessageType: ADD_HOST,
			Message:     strings.Join(sendData, ","),
		}
		sendJson, _ := json.Marshal(sendBody)
		SendMessage(host, string(sendJson))
	}
}

func BroadcastBlock(newBlock Block) {
	for _, host := range hosts {
		sendData, _ := json.Marshal(newBlock)
		sendBody := MessageBody{
			MessageType: ADD_BLOCK,
			Message:     string(sendData),
		}
		sendJson, _ := json.Marshal(sendBody)
		SendMessage(host, string(sendJson))
	}
}

func Server() {
	ln, _ := net.Listen(PROTOCOL, localHost)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		defer conn.Close()
		recieveBody := MessageBody{}
		data := GetMessage(conn)
		_ = json.Unmarshal([]byte(data), &recieveBody)

		switch recieveBody.MessageType {
		case NEW_HOST:
			/*
				1) Sucede cuando un nuevo cliente se conecta
				2) Recibimos la [IP:PORT] del cliente
				3) Una vez recibido:
				- Enviamos al cliente nuestro directorio de hosts
				- Enviamos a todos los hosts en nuestro directorio la [IP:PORT] del cliente
				- Finalmente, agregamos la [IP:PORT] del cliente a nuestro directorio de hosts
			*/
			clientHost := recieveBody.Message
			totalHost := strings.Join(append(hosts, localHost), ",")
			sendBody := MessageBody{
				MessageType: ADD_HOST,
				Message:     totalHost,
			}
			sendJson, _ := json.Marshal(sendBody)
			SendMessage(clientHost, string(sendJson))
			Broadcast(clientHost)
			hosts = append(hosts, clientHost)
		case ADD_HOST:
			/*
				1) Sucede cuando un nuevo cliente se conecta, después de "NEW_HOST"
				2) Como este nuevo cliente no tiene las direcciones de todos
				3) El host "destHost" es quien se encarga de comunicar a todos
				2) Enviamos la [IP:PORT] del cliente a todos los host del directorio
			*/
			clientHost := recieveBody.Message
			hosts = strings.Split(clientHost, ",")
		case NEW_BLOCK:
			/*
				1) Sucede cuando un nuevo cliente se conecta
				2) El cliente recibe del "destHost" la blockchain
			*/
			clientHost := recieveBody.Message
			sendJson, _ := json.Marshal(localBlockChain.Chain)
			sendBody := MessageBody{
				MessageType: SET_BLOCKS,
				Message:     string(sendJson),
			}
			sendMsg, _ := json.Marshal(sendBody)
			SendMessage(clientHost, string(sendMsg))
		case SET_BLOCKS:
			/*
				1) Sucede luego de "NEW_BLOCK"
				2) Recibimos del cliente la blockchain
			*/
			clientBlockChain := recieveBody.Message
			_ = json.Unmarshal([]byte(clientBlockChain), &localBlockChain.Chain)
		case ADD_BLOCK:
			/*
				1) Sucede luego de "BroadcastBlock"
				2) Recibimos del cliente un nuevo block
			*/
			clientBlock := recieveBody.Message
			block := Block{}
			json.Unmarshal([]byte(clientBlock), &block)
			localBlockChain.Chain = append(localBlockChain.Chain, block)
		}
	}
}

/* [SERVER] END */

/* [BLOCKCHAIN] START */

var localBlockChain BlockChain

type AcademicRecord struct {
	Name         string
	Year         string
	University   string
	Course       string
	Teacher      string
	Calification string
}

type Block struct {
	Index        int
	Timestamp    time.Time
	Data         AcademicRecord
	PreviousHash string
	Hash         string
}

type BlockChain struct {
	Chain []Block
}

func (block *Block) CalculateHash() string {
	src := fmt.Sprintf("%d-%s-%s", block.Index, block.Timestamp.String(), block.Data)
	return base64.StdEncoding.EncodeToString([]byte(src))
}

func (blockChain *BlockChain) CreateGenesisBlock() Block {
	block := Block{
		Index:        0,
		Timestamp:    time.Now(),
		Data:         AcademicRecord{},
		PreviousHash: "0",
	}
	block.Hash = block.CalculateHash()
	return block
}

func (blockChain *BlockChain) GetLatesBlock() Block {
	n := len(blockChain.Chain)
	return blockChain.Chain[n-1]
}

func (blockChain *BlockChain) AddBlock(block Block) {
	block.Timestamp = time.Now()
	block.Index = blockChain.GetLatesBlock().Index + 1
	block.PreviousHash = blockChain.GetLatesBlock().Hash
	block.Hash = block.CalculateHash()
	blockChain.Chain = append(blockChain.Chain, block)
}

func (blockChain *BlockChain) IsChainValid() bool {
	n := len(blockChain.Chain)
	for i := 1; i < n; i++ {
		currentBlock := blockChain.Chain[i]
		previousBlock := blockChain.Chain[i-1]
		if currentBlock.Hash != currentBlock.CalculateHash() {
			return false
		}
		if currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}
	}
	return true
}

func CreateBlockChain() BlockChain {
	bc := BlockChain{}
	genesisBlock := bc.CreateGenesisBlock()
	bc.Chain = append(bc.Chain, genesisBlock)
	return bc
}

/* [BLOCKCHAIN] END */

func NewRecord() {
	fmt.Println("==== New academic record ====")
	in := bufio.NewReader(os.Stdin)
	fmt.Println("*Please enter the next information.")
	fmt.Print("Name: ")
	name, _ := in.ReadString('\n')
	name = strings.TrimSpace(name)
	fmt.Print("Year: ")
	year, _ := in.ReadString('\n')
	year = strings.TrimSpace(year)
	fmt.Print("University: ")
	university, _ := in.ReadString('\n')
	university = strings.TrimSpace(university)
	fmt.Print("Course: ")
	course, _ := in.ReadString('\n')
	course = strings.TrimSpace(course)
	fmt.Print("Calification: ")
	calification, _ := in.ReadString('\n')
	calification = strings.TrimSpace(calification)
	record := AcademicRecord{
		Name:         name,
		Year:         year,
		University:   university,
		Course:       course,
		Calification: calification,
	}
	newBlock := Block{
		Data: record,
	}
	localBlockChain.AddBlock(newBlock)
	BroadcastBlock(newBlock)
	fmt.Println("You have registered successfully.")
	time.Sleep(2 * time.Second)
}

func UpdateRecord() {
	fmt.Println("Update record!")
}

func ListRecord() {
	blocks := localBlockChain.Chain[1:]
	fmt.Printf("==== List academic records [Total: %d] ====\n", len(blocks))
	for index, block := range blocks {
		record := block.Data
		fmt.Printf("Record [#%d]\n", index+1)
		fmt.Printf(" Name: %s\n", record.Name)
		fmt.Printf(" Year: %s\n", record.Year)
		fmt.Printf(" University: %s\n", record.University)
		fmt.Printf(" Course: %s\n", record.Course)
		fmt.Printf(" Calification: %s\n", record.Calification)
	}
}

func ListHost() {
	var nroHost = 1
	fmt.Printf("==== List hosts [Total: %d] ====\n", len(hosts)+1)
	fmt.Printf("%d. %s (Your host)\n", nroHost, localHost)
	for _, host := range hosts {
		nroHost++
		fmt.Printf("%d. %s\n", nroHost, host)
	}
}

func main() {
	var destHost string
	var option int
	in := bufio.NewReader(os.Stdin)

	fmt.Print("==== Academic Block APP ====\n")
	fmt.Print("==== Conection to Blockchain ====\n")
	fmt.Print("Enter your local host [IP:PORT]: ")
	localHost, _ = in.ReadString('\n')
	localHost = strings.TrimSpace(localHost)
	fmt.Print("Enter destination host [IP:PORT] (Empty to be the first host): ")
	destHost, _ = in.ReadString('\n')
	destHost = strings.TrimSpace(destHost)

	localBlockChain = CreateBlockChain()
	go Server()

	fmt.Println("==== Welcome to Academic Block APP ====")

	if destHost != "" {
		sendBody := MessageBody{
			MessageType: NEW_HOST,
			Message:     localHost,
		}
		sendJson, _ := json.Marshal(sendBody)
		SendMessage(destHost, string(sendJson))
		sendBody = MessageBody{
			MessageType: NEW_BLOCK,
			Message:     localHost,
		}
		sendJson, _ = json.Marshal(sendBody)
		SendMessage(destHost, string(sendJson))
	}

	for {
		fmt.Println("==== Menu option ====")
		fmt.Print("1. New academic record\n2. Update academic record\n3. List academic records\n4. List hosts\n")
		fmt.Print("Enter option [1|2|3|4]: ")
		fmt.Scanf("%d\n", &option)

		switch option {
		case NEW_ACADEMIC_RECORD:
			NewRecord()
		case UPDATE_ACADEMIC_RECORD:
			UpdateRecord()
		case LIST_ACADEMIC_RECORD:
			ListRecord()
		case LIST_HOSTS:
			ListHost()
		default:
			fmt.Println("Enter option is no valid, please try again.")
		}
	}
}
