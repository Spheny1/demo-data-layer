package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

//Below is grabbed form env variables set from spinnaker deployment
var serviceDiscovery string;
var postgres string;
var myURI string;
var deploymentName string;
var DB *sql.DB;
func getcount(w http.ResponseWriter, req *http.Request){
	var num int

	err := DB.Ping()
	if err != nil {
		log.Print(err)
		panic(err)
	}
	log.Println("SELECT num FROM count_" + deploymentName)
	rows, err := DB.Query("SELECT num FROM count_" + deploymentName + ";")
	if err != nil {
		log.Fatal(err.Error())
	}
   	defer rows.Close()
	for rows.Next() {
       		rows.Scan(&num)
   	}
	fmt.Fprintf(w,strconv.Itoa(num))	
}

func addcount(w http.ResponseWriter, req *http.Request){
	var num int
	var id int
	rows, err := DB.Query("SELECT * FROM count_" + deploymentName + ";")
	if err != nil {
		log.Fatal(err.Error())
	}
   	defer rows.Close()
	for rows.Next() {
       		rows.Scan(&id ,&num)
   	}
	updateSQL := "UPDATE count_" + deploymentName + " SET num = " + strconv.Itoa(num + 3) + " WHERE idCount = " + strconv.Itoa(id) + ";"
	statement, err := DB.Prepare(updateSQL) 
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() 
	fmt.Fprintf(w,"Incremented! \n")	
}

func main(){
	serviceDiscovery=os.Getenv("SERVICE_DISCOVERY_URI")
	postgres = os.Getenv("POSTGRES")
	deploymentName = os.Getenv("DEPLOYMENT_NAME")
	postgresPort := 5432
	myURI = os.Getenv("MY_URI")
	//Setup service discovery
	log.Println("registering service at " + serviceDiscovery + "/register?ip=" + myURI + "&name=" + deploymentName)
	_, err := http.Get(serviceDiscovery + "/register?ip=" + myURI + "&name=" + deploymentName)
	if err != nil {
		panic(err)
	}
	//Setup db
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable","localhost", postgresPort, "postgres", "mysecretpassword")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
  		panic(err)
	}
	//defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Print(err)
		panic(err)
	}

	createCountCalledSQL := `CREATE TABLE count_` + deploymentName +` (
		idCount serial primary key,		
		num integer
	  );`
	log.Println("Create alias table...")
	statement, err := db.Prepare(createCountCalledSQL) 
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() 
	log.Println("count table created")
	insertCountCalledSQL := `INSERT INTO count_` + deploymentName +` (num) VALUES (0);`
	statement, err = db.Prepare(insertCountCalledSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
	DB = db
	log.Println("count seeded")
	http.HandleFunc("/count", getcount)
	http.HandleFunc("/addcount", addcount)
	log.Print("running server")
	log.Print(http.ListenAndServe("0.0.0.0:8080", nil))
}
