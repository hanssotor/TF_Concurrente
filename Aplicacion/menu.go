package main

import (
	"fmt"
)

const (
	NEW_ACADEMIC_RECORD    = 1
	UPDATE_ACADEMIC_RECORD = 2
	LIST_ACADEMIC_RECORD   = 3
	LIST_HOSTS             = 4
)

var LOCALHOST string

func NewRecord() {
	fmt.Println("New record!")
}

func UpdateRecord() {
	fmt.Println("Update record!")
}

func ListRecord() {
	fmt.Println("List record!")
}

func ListHost() {
	fmt.Println("List host!")
}

func main() {
	var destHost string
	var option int

	fmt.Print("==== Academic Block APP ====\n")
	fmt.Print("==== Conection to Blockchain ====\n")
	fmt.Print("\tEnter your local host [IP:PORT]: ")
	fmt.Scanf("%s\n", &LOCALHOST)
	fmt.Print("\tEnter destination host [IP:PORT] (Empty to be the first host): ")
	fmt.Scanf("%s\n", &destHost)

	fmt.Println("==== Welcome to Academic Block APP ====")

	for {
		fmt.Println("==== Menu option ====")
		fmt.Print("1. New academic record\n2. Update academic record\n3. List academic records\n4. List hosts\n")
		fmt.Print("\tEnter option [1|2|3|4]: ")
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
			fmt.Println("\tEnter option is no valid, please try again.")
		}
	}
}
