package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Participants
type Gender string
const (
	MALE    Gender = "MALE"
	FEMALE    Gender = "FEMALE"
)
type Department string
const (
	Aero    Department = "AERO"
	Ce    Department = "CE"
	Civil    Department = "CIVIL"
	Ec   Department = "EC"
	Ele   Department = "ELE"
	Ic   Department = "IC"
	It   Department = "IT"
	Mca   Department = "MCA"
)
type ShirtSize string
const (
	XS    ShirtSize = "XS"
	S    ShirtSize = "S"
	M    ShirtSize = "M"
	L    ShirtSize = "L"
	XL    ShirtSize = "XL"
	XXL    ShirtSize = "XXL"
	XXXL    ShirtSize = "XXXL"
)
type Participant struct{
	gorm.Model
	
	Name string	`json:"name" validate:"required"`
	Email string `json:"email" gorm:"unique_index" validate:"required"`
	Phone int `json:"phone" validate:"required"`
	Gender `gorm:"type:gender" json:"gender" validate:"required"`
	Department `json:"department" validate:"required"`
	ShirtSize `json:"shirt_size" validate:"required"`
	TeamID uint `json:"team_id" validate:"required"`
	Team Team `json:"team"`
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

  Name string `json:"name" validate:"required"`
  MaleCount int `json:"male_count" gorm:"default:0"`
  FemaleCount int `json:"female_count" gorm:"default:0"`
  ProjectType `gorm:"type:project_type" json:"project_type" validate:"required"` 
  LocationID uint `json:"location_id" validate:"required"`
  Location Location `json:"location" validate:"required"`
  Members []Participant `json:"members" validate:"required"`
}

// Locations
type Wing string
const (
	CEF Wing = "CEF"
	CES Wing = "CES"
	IT Wing = "IT"
	EC Wing = "EC"
	MCA Wing ="MCA"
	ARCH Wing = "ARCH"
)
type Location struct {
	gorm.Model
	Name string `json:"name" validate:"required"`
	Wing string `json:"wing" validate:"required"`
	Capacity int `json:"capacity" validate:"required"` // in terms of teams
	Teams []Team `json:"teams" validate:"required"`
}

// Attendance
type Attendance struct {
	gorm.Model
	ActionID uint `json:"action_id" validate:"required"` 
	Action Action `json:"action" validate:"required"`
	ParticipantID uint `json:"participant_id" validate:"required"` 
	Participant Participant `json:"participant" validate:"required"`
}

// Actions
type Action struct {
	gorm.Model
	Title string `json:"title" validate:"required"`
	Valid bool `json:"valid" gorm:"default:true"`
	Attendance []Attendance
}

var db *gorm.DB
var err error

func main() {
	// Loading enviroment variables
	host := os.Getenv("DB_HOST")
	dbport := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	dbpassword := os.Getenv("DB_PASSWORD")

	// Database connection string
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s port=%s", host, user, dbname, dbpassword, dbport)
	
	// Openning connection to database
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		FullSaveAssociations: true,
	})

	if err != nil {
		panic(err)
	} else {
		log.Printf("🤝 Database connected successfully") 
	}

	// Migrate the schema
	db.AutoMigrate(&Location{}, &Team{}, &Participant{}, &Action{}, &Attendance{})
	db.AutoMigrate(&Team{})
	db.AutoMigrate(&Participant{})

	// Close the databse connection when the main function closes
	//// defer db.Close()

	/*----------- API routes ------------*/
	port := "8080"

  	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
    	port = fromEnv
  	}
	log.Printf("🚀 Starting up on http://localhost:%s", port)

	r := mux.NewRouter()
	
    http.Handle("*",handlers.LoggingHandler(os.Stdout, r))

	r.HandleFunc("/teams", GetTeams).Methods("GET")
	r.HandleFunc("/teams/{id}", GetTeam).Methods("GET")
	r.HandleFunc("/participants", GetParticipants).Methods("GET")
	r.HandleFunc("/participants/{id}", GetParticipant).Methods("GET")
	r.HandleFunc("/locations", GetLocations).Methods("GET")
	r.HandleFunc("/attendance", GetAttendance).Methods("GET")
	r.HandleFunc("/actions", GetActions).Methods("GET")

	r.HandleFunc("/teams", CreateTeam).Methods("POST")
	r.HandleFunc("/participants", CreateParticipant).Methods("POST")
	r.HandleFunc("/attendance", CreateAttendance).Methods("POST")
	r.HandleFunc("/actions", CreateAction).Methods("POST")

	r.HandleFunc("/teams/{id}", DeleteTeam).Methods("DELETE")
	r.HandleFunc("/participants/{id}", DeleteParticipant).Methods("DELETE")
	r.HandleFunc("/attendance/{id}", DeleteAttendance).Methods("DELETE")
	r.HandleFunc("/actions/{id}", DeleteAction).Methods("DELETE")

	http.ListenAndServe(":8080", r)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&team)
}
func GetTeams(w http.ResponseWriter, r *http.Request) {
	var teams []Team

	db.Preload("Location").Preload("Members").Find(&teams)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&teams)
}
func CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team Team
	// ctx := context.WithValue(r.Context(), "user", "123")
	json.NewDecoder(r.Body).Decode(&team)
	fmt.Printf("%+v", team)
	createdTeam := db.Create(&team)
	err = createdTeam.Error
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&createdTeam)
}
func DeleteTeam(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var team Team

	db.First(&team, params["id"])
	db.Delete(&team)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&team)
}

/*------- Participant ------*/
func GetParticipant(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var participant Participant

	db.First(&participant, params["id"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&participant)
}
func GetParticipants(w http.ResponseWriter, r *http.Request) {
	var participants []Participant

	db.Preload("Team").Find(&participants)

	w.Header().Set("Content-Type", "application/json")
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

	var team Team
	db.Where("ID = ?", participant.TeamID).First(&team)

  // Increment Counter
  	if(participant.Gender == MALE){
		db.Model(&team).Update("MaleCount", team.MaleCount+1)
	} else{
		db.Model(&team).Update("FemaleCount", team.FemaleCount+1)
	}
	

	// db.Model(&Team{}).Where("ID = ?", participant.TeamID).Update("name", "hello")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&createdParticipant)
}
func DeleteParticipant(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var participant Participant

	db.First(&participant, params["id"])
	db.Delete(&participant)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&participant)
}

/*------- Location ------*/
func GetLocation(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var location Location

	db.First(&location, params["id"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&location)
}
func GetLocations(w http.ResponseWriter, r *http.Request) {
	var locations []Location

	db.Preload("Teams").Find(&locations)

	w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&createdLocation)
}
func DeleteLocation(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var location Location

	db.First(&location, params["id"])
	db.Delete(&location)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&location)
}

/*------- Attendance ------*/
func GetAttendance(w http.ResponseWriter, r *http.Request) {
	var attendance []Attendance

	db.Preload("Participant").Preload("Action").Find(&attendance)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&attendance)
}
func CreateAttendance(w http.ResponseWriter, r *http.Request) {
	var attendance Attendance
	json.NewDecoder(r.Body).Decode(&attendance)

	createdAttendance := db.Create(&attendance)
	err = createdAttendance.Error
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&createdAttendance)
}
func DeleteAttendance(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var attendance Attendance

	db.First(&attendance, params["id"])
	db.Delete(&attendance)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&attendance)
}

/*------- Action ------*/
func GetActions(w http.ResponseWriter, r *http.Request) {
	var actions []Action

	db.Preload("Attendance").Preload("Attendance.Action").Preload("Attendance.Participant").Preload("Attendance.Participant.Team").Preload("Attendance.Participant.Team.Location").Find(&actions)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&actions)
}
func CreateAction(w http.ResponseWriter, r *http.Request) {
	var action Action
	json.NewDecoder(r.Body).Decode(&action)

	createdAction := db.Create(&action)
	err = createdAction.Error
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&createdAction)
}
func DeleteAction(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var action Action

	db.First(&action, params["id"])
	db.Delete(&action)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&action)
}