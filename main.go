package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*-------- Models --------*/

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
	Csd    Department = "CSD"
	Ec   Department = "EC"
	Ele   Department = "ELE"
	Ic   Department = "IC"
	It   Department = "IT"
	Mech   Department = "MEC"
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
type Batch string
const (
	FY    Batch = "FY"
	SY    Batch = "SY"
	TY    Batch = "TY"
	LY    Batch = "LY"
)
type Participant struct{
	gorm.Model
	
	Name string	`json:"name" validate:"required"`
	Email string `json:"email" gorm:"unique_index"`
	Phone int `json:"phone" validate:"required"`
	// Gender string `gorm:"type:gender" json:"gender"`
	Department `json:"department" validate:"required"`
	Batch `json:"batch" validate:"required"`
	ShirtSize `json:"shirt_size"`
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
	CIVIL Wing = "CIVIL"
)
type Location struct {
	gorm.Model
	Name string `json:"name" validate:"required"`
	Wing `json:"wing" validate:"required"`
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

	// r := mux.NewRouter()
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Route("/teams", func(r chi.Router) {
		r.Get("/", GetTeams)
		r.Get("/{id}", GetTeam)

		r.Post("/", CreateTeam)
		r.Delete("/{id}", DeleteTeam)
	})

	r.Route("/participants", func(r chi.Router) {
		r.Get("/", GetParticipants)
		r.Get("/{id}", GetParticipant)

		r.Post("/", CreateParticipant)
		r.Delete("/{id}", DeleteParticipant)
	})

	r.Route("/locations", func(r chi.Router) {
		r.Get("/", GetLocations)
		r.Get("/{id}", GetLocation)

		r.Post("/", CreateLocation)
		r.Delete("/{id}", DeleteLocation)
	})
	
	r.Route("/attendance", func(r chi.Router) {
		r.Get("/", GetAttendance)
		// r.Get("/count/{action_id}", GetAttendance)
		
		r.Post("/", CreateAttendance)
		r.Delete("/{id}", DeleteAttendance)
	})
	
	r.Route("/actions", func(r chi.Router) {
		r.Get("/", GetActions)
		r.Get("/with_data", GetActionsWithData)
		
		r.Post("/", CreateAction)
		r.Delete("/{id}", DeleteAction)
	})

	http.ListenAndServe(":8080", r)
}

/*-------- API Controllers --------*/

/*----- Team ------*/
func GetTeam(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r,"id")
	var team Team
	var participants []Participant

	db.First(&team, id)
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
	// fmt.Printf("%+v", team)
	createdTeam := db.Create(&team)
	err = createdTeam.Error
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&createdTeam)
}
func DeleteTeam(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r,"id")

	var team Team

	db.First(&team, id)
	db.Delete(&team)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&team)
}

/*------- Participant ------*/
func GetParticipant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r,"id")
	var participant Participant

	db.Preload("Team").Preload("Team.Location").First(&participant, id)

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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
	}

// 	var team Team
// 	db.Where("ID = ?", participant.TeamID).First(&team)

//   // Increment Counter
//   	if(participant.Gender == MALE){
// 		db.Model(&team).Update("MaleCount", team.MaleCount+1)
// 	} else{
// 		db.Model(&team).Update("FemaleCount", team.FemaleCount+1)
// 	}
	

	// db.Model(&Team{}).Where("ID = ?", participant.TeamID).Update("name", "hello")

	w.Header().Set("Content-Type", "application/json")
	
	json.NewEncoder(w).Encode(&createdParticipant)
}
func DeleteParticipant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r,"id")

	var participant Participant

	db.First(&participant, id)
	db.Delete(&participant)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&participant)
}

/*------- Location ------*/
func GetLocation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r,"id")
	var location Location

	db.First(&location, id)

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
	j, _ := json.MarshalIndent(createdLocation, "", "🐱");
	fmt.Println(string(j))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&createdLocation)
}
func DeleteLocation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r,"id")

	var location Location

	db.First(&location, id)
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
// func GetAttendanceCount(w http.ResponseWriter, r *http.Request) {
// 	action_id := chi.URLParam(r,"action_id")
// 	var attendance []Attendance

// 	db.Table("Attendance").Select("COUNT()").Find(&attendance)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(&attendance)
// }
func CreateAttendance(w http.ResponseWriter, r *http.Request) {
	var attendance Attendance
	json.NewDecoder(r.Body).Decode(&attendance)

	allow_duplicates:= chi.URLParam(r,"allowDuplicates")
	fmt.Println(allow_duplicates)
	if(allow_duplicates=="true"){
		fmt.Println("Allowed duplicates")
		createdAttendance := db.Create(&attendance)
		err = createdAttendance.Error
		if err != nil {
			fmt.Println(err)
			}	
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&createdAttendance)
		}
	if(allow_duplicates=="false" || allow_duplicates==""){
		fmt.Println("Disallowed duplicates")
			createdAttendance := db.FirstOrCreate(&attendance)
		fmt.Println("2",allow_duplicates)
		err = createdAttendance.Error
		if err != nil {
			fmt.Println(err)
		}	
		if(createdAttendance.RowsAffected ==0){
			w.WriteHeader(304)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&createdAttendance)
	}

}
func DeleteAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r,"id")

	var attendance Attendance

	db.First(&attendance, id)
	db.Delete(&attendance)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&attendance)
}

/*------- Action ------*/
func GetActions(w http.ResponseWriter, r *http.Request) {
	// var actions struct {
	// 	ID uint
	// 	title string }
// db.Table("actions").Select("ID","title").Where("valid is true").Scan(&actions)
	var actions []Action

	db.Preload("Attendance").Preload("Attendance.Action").Where(&Action{Valid: true}).Find(&actions)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&actions)
}
func GetActionsWithData(w http.ResponseWriter, r *http.Request) {
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
	id := chi.URLParam(r,"id")

	var action Action

	db.First(&action, id)
	db.Delete(&action)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&action)
}