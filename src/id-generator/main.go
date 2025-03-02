package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

)

func main(){

	datacenterID, err := GetDatacenterID()
	if err != nil{
		return  
	}
	machineID, err := GetMachineID()
	if err != nil{
		return  
	}


	snowflake, err := NewSnowflake(datacenterID, machineID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	
	l := log.New(os.Stdout, "id-generator-api", log.LstdFlags)

	gh := NewGeneratorHandler(l, snowflake)

	http.HandleFunc("/getID", gh.GenerateId)


	log.Printf("starting server at :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))

}