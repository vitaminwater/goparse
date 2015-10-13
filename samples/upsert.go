package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vitaminwater/goparse"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s API_ID REST_API_KEY\n")
		return
	}

	parse.InitAPI(os.Args[1], os.Args[2])

	address := parse.NewClassObject("Address")
	address.Set("name", "Test name")
	address.Set("price", 350000)
	address.Set("type", "appt")
	address.Set("description", "Lorem ipsum pouet pen erj ejrl ekjrel")
	address.Set("active", true)

	address.Set("address", "53 rue des petits champs")
	address.Set("zip", "75002")
	address.Set("city", "Paris")

	address.Set("rooms", 3)
	address.Set("surface", 42)

	if err := address.Save(); err != nil {
		panic(err)
	}

	address.Set("name", "toto test 2")
	address.Set("address", "42 rue des sablons")
	if err := address.Save(); err != nil {
		panic(err)
	}
	json, _ := json.Marshal(address)
	fmt.Println("saved: ", string(json))
}
