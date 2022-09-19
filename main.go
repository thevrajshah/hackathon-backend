package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

// Participants
type Gender string
const (
	MALE    Gender = "MALE"
	FEMALE    Gender = "FEMALE"
)
type Participant struct{
	gorm.Model
	
	Name string	
	Email string `gorm:"unique_index"`
	Phone int
	TeamID uint
	Gender `gorm:"type:project_type"`
} 
	
// Teams
type ProjectType string
const (
	HARDWARE    ProjectType = "HARDWARE"
	IOT    ProjectType = "IOT"
	SOFTWARE    ProjectType = "SOFTWARE"
)
type Team struct {
  gorm.Model

  Name string
  MaleCount int
  FemaleCount int
  ProjectType `gorm:"type:project_type"`
  LocationID int
  Location Location
  Members []Participant
}

// Locations
type Wing string
const (
	CE_F Wing = "CE_F"
	CE_S Wing = "CE_S"
	IT Wing = "IT"
	EC Wing = "EC"
	MCA Wing ="MCA"
	ARCH Wing = "ARCH"
)
type Location struct {
	gorm.Model
	Name string
	Wing  `gorm:"type:wing"`
	TeamCapacity int // in terms of teams
}

var db *gorm.DB
var err error

func main() {

	// Loading enviroment variables
	// dialect := os.Getenv("DIALECT")
	host := os.Getenv("DB_HOST")
	dbport := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	dbpassword := os.Getenv("DB_PASSWORD")

	// Database connection string
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s port=%s", host, user, dbname, dbpassword, dbport)
	
	// Openning connection to database
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Database connected successfully 🤝")
	}

	// Migrate the schema
	db.AutoMigrate(&Location{})
	db.AutoMigrate(&Team{})
	db.AutoMigrate(&Participant{})

	// Close the databse connection when the main function closes
	//// defer db.Close()

	/*----------- API routes ------------*/
	router := mux.NewRouter()

	router.HandleFunc("/teams", GetTeams).Methods("GET")
	router.HandleFunc("/team/{id}", GetTeam).Methods("GET")
	router.HandleFunc("/participants", GetParticipants).Methods("GET")
	router.HandleFunc("/participant/{id}", GetParticipant).Methods("GET")
	router.HandleFunc("/locations", GetParticipants).Methods("GET")
	router.HandleFunc("/location/{id}", GetParticipant).Methods("GET")

	router.HandleFunc("/create/team", CreateTeam).Methods("POST")
	router.HandleFunc("/create/participant", CreateParticipant).Methods("POST")

	router.HandleFunc("/delete/team/{id}", DeleteTeam).Methods("DELETE")
	router.HandleFunc("/delete/participant/{id}", DeleteParticipant).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))


//   // Create
//   db.Create(&Product{Code: "D42", Price: 100})

//   // Read
//   var product Product
//   db.First(&product, 1) // find product with integer primary key
//   db.First(&product, "code = ?", "D42") // find product with code D42

//   // Update - update product's price to 200
//   db.Model(&product).Update("Price", 200)
//   // Update - update multiple fields
//   db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
//   db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

//   // Delete - delete product
//   db.Delete(&product, 1)
}


/*-------- API Controllers --------*/

/*----- Team ------*/
func GetTeam(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var team Team
	var participants []Participant

	db.First(&team, params["id"])
	db.Model(&team)

	team.Members = participants

	json.NewEncoder(w).Encode(&team)
}

func GetTeams(w http.ResponseWriter, r *http.Request) {
	var teams []Team

	db.Find(&teams)

	json.NewEncoder(w).Encode(&teams)
}

func CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team Team
	json.NewDecoder(r.Body).Decode(&team)

	createdTeam := db.Create(&team)
	err = createdTeam.Error
	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(&createdTeam)
}

func DeleteTeam(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var team Team

	db.First(&team, params["id"])
	db.Delete(&team)

	json.NewEncoder(w).Encode(&team)
}

/*------- Participant ------*/
func GetParticipant(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var participant Participant

	db.First(&participant, params["id"])

	json.NewEncoder(w).Encode(&participant)
}

func GetParticipants(w http.ResponseWriter, r *http.Request) {
	var participants []Participant

	db.Find(&participants)

	json.NewEncoder(w).Encode(&participants)
}

func CreateParticipant(w http.ResponseWriter, r *http.Request) {
	var participant Participant
	json.NewDecoder(r.Body).Decode(&participant)

	createdParticipant := db.Create(&participant)
	err = createdParticipant.Error
	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(&createdParticipant)
}

func DeleteParticipant(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var participant Participant

	db.First(&participant, params["id"])
	db.Delete(&participant)

	json.NewEncoder(w).Encode(&participant)
}

/*------- Location ------*/
func GetLocation(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var location Location

	db.First(&location, params["id"])

	json.NewEncoder(w).Encode(&location)
}

func GetLocations(w http.ResponseWriter, r *http.Request) {
	var locations []Location

	db.Find(&locations)

	json.NewEncoder(w).Encode(&locations)
}

func CreateLocation(w http.ResponseWriter, r *http.Request) {
	var location Location
	json.NewDecoder(r.Body).Decode(&location)

	createdLocation := db.Create(&location)
	err = createdLocation.Error
	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(&createdLocation)
}

func DeleteLocation(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var location Location

	db.First(&location, params["id"])
	db.Delete(&location)

	json.NewEncoder(w).Encode(&location)
}