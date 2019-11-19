package main

import (
	"bufio"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sajari/regression"
)

type MessageType int32

const (
	NEW_ACADEMIC_RECORD              = 1
	LIST_ACADEMIC_RECORD             = 2
	LIST_HOSTS                       = 3
	PREDICT                          = 4
	PROTOCOL                         = "tcp"
	NEW_HOST             MessageType = 0
	ADD_HOST             MessageType = 1
	ADD_BLOCK            MessageType = 2
	NEW_BLOCK            MessageType = 3
	SET_BLOCKS           MessageType = 4
)

/* [SERVER] END */

var localHost string
var hosts []string

type MessageBody struct {
	MessageType MessageType
	Message     string
}

func errHandle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseRecord(record []string) AcademicRecord {
	var rec AcademicRecord
	rec.Nombre = record[0]
	rec.Carrera = record[1]
	rec.Ciclo = record[2]
	rec.Nivel = record[3]
	rec.Universidad = record[4]
	rec.Promedio = record[5]
	rec.NumCursos = record[6]
	rec.Creditos = record[7]

	return rec
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

func Server(updatedBlocks chan<- int) {
	ln, _ := net.Listen(PROTOCOL, localHost)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go Handle(conn, updatedBlocks)
	}
}

func Handle(conn net.Conn, updatedBlocks chan<- int) {
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
		updatedBlocks <- 0
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

/* [SERVER] END */

/* [BLOCKCHAIN] START */

var localBlockChain BlockChain

type AcademicRecord struct {
	Nombre            string
	Carrera           string
	Ciclo             string
	Nivel             string
	Universidad       string
	Promedio          string
	NumCursos         string
	Creditos          string
	Promedio_esperado float64
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
	src := fmt.Sprintf("%d-%s-%s-s%", block.Index, block.Timestamp.String(), block.Data, block.PreviousHash)
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
	fmt.Print("Nombre: ")
	nombre, _ := in.ReadString('\n')
	nombre = strings.TrimSpace(nombre)
	fmt.Print("Carrera: ")
	carrera, _ := in.ReadString('\n')
	carrera = strings.TrimSpace(carrera)
	fmt.Print("Ciclo: ")
	ciclo, _ := in.ReadString('\n')
	ciclo = strings.TrimSpace(ciclo)
	fmt.Print("Nivel: ")
	nivel, _ := in.ReadString('\n')
	nivel = strings.TrimSpace(nivel)
	fmt.Print("Universidad: ")
	universidad, _ := in.ReadString('\n')
	universidad = strings.TrimSpace(universidad)
	fmt.Print("Promedio: ")
	promedio, _ := in.ReadString('\n')
	promedio = strings.TrimSpace(promedio)
	fmt.Print("Numero de cursos: ")
	numcursos, _ := in.ReadString('\n')
	numcursos = strings.TrimSpace(numcursos)
	fmt.Print("Numero de creditos: ")
	creditos, _ := in.ReadString('\n')
	creditos = strings.TrimSpace(creditos)

	record := AcademicRecord{
		Nombre:      nombre,
		Carrera:     carrera,
		Ciclo:       ciclo,
		Nivel:       nivel,
		Universidad: universidad,
		Promedio:    promedio,
		NumCursos:   numcursos,
		Creditos:    creditos,
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
		fmt.Printf(" Nombre: %s\n", record.Nombre)
		fmt.Printf(" Carrera: %s\n", record.Carrera)
		fmt.Printf(" Ciclo: %s\n", record.Ciclo)
		fmt.Printf(" Universidad: %s\n", record.Universidad)
		fmt.Printf(" Promedio: %s\n", record.Promedio)
		fmt.Printf(" Numero de cursos: %s\n", record.NumCursos)
		fmt.Printf(" Numero de creditos: %s\n", record.Creditos)

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

func Predict() {
	dataSetArc, err := os.Open("Records.csv")
	errHandle(err)
	defer dataSetArc.Close()
	reader := csv.NewReader(dataSetArc)
	reader.Comma = ','
	var dataSet []AcademicRecord
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		errHandle(err)
		dataSet = append(dataSet, parseRecord(record))
	}
	r := new(regression.Regression)
	r.SetObserved("Promedio esperado")
	r.SetVar(0, "Numero de cursos")
	r.SetVar(1, "Numero de creditos")
	r.SetVar(2, "Nivel")
	for i := 0; i < len(dataSet); i++ {
		prom, _ := strconv.ParseFloat(dataSet[i].Promedio, 64)
		cur, _ := strconv.ParseFloat(dataSet[i].NumCursos, 64)
		cred, _ := strconv.ParseFloat(dataSet[i].Creditos, 64)
		niv, _ := strconv.ParseFloat(dataSet[i].Nivel, 64)
		r.Train(regression.DataPoint(prom, []float64{cur, cred, niv}))
	}
	r.Run()
	var curs float64
	var creds float64
	var nivs float64

	fmt.Print("==== Prediction ====\n")
	fmt.Print("Ingrese los datos del estudiante:\n")
	fmt.Print("Numero de cursos que lleva: ")
	fmt.Scanf("%f\n", &curs)
	fmt.Print("Numero de creditos que lleva: ")
	fmt.Scanf("%f\n", &creds)
	fmt.Print("Nivel (ciclo) en que se encuentra: ")
	fmt.Scanf("%f\n", &nivs)
	prediction, _ := r.Predict([]float64{curs, creds, nivs})
	fmt.Printf("Regression formula:\n%v\n", r.Formula)
	fmt.Printf("El promedio esperado es de:\n%v\n", prediction)

}

func main() {
	updatedBlocks := make(chan int)
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
	go Server(updatedBlocks)

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
		<-updatedBlocks
	}

	for {
		fmt.Println("==== Menu de opciones ====")
		fmt.Print("1. Nuevo registro academico\n2. Listar registros academicos\n3. Listar hosts\n4. Predecir\n")
		fmt.Print("Enter option [1|2|3|4]: ")
		fmt.Scanf("%d\n", &option)

		switch option {
		case NEW_ACADEMIC_RECORD:
			NewRecord()
		case LIST_ACADEMIC_RECORD:
			ListRecord()
		case LIST_HOSTS:
			ListHost()
		case PREDICT:
			Predict()
		default:
			fmt.Println("Enter option is no valid, please try again.")
		}
	}
}
