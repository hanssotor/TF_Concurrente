package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	NEW_ACADEMIC_RECORD    = 1
	UPDATE_ACADEMIC_RECORD = 2
	LIST_ACADEMIC_RECORD   = 3
	LIST_HOSTS             = 4
)

var localHost string
var hosts []string

/* [BLOCKCHAIN] START*/

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
	//BroadcastBlock(newBlock)
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

	fmt.Println("==== Welcome to Academic Block APP ====")
	localBlockChain = CreateBlockChain()

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
