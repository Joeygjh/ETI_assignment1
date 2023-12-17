package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"io/ioutil"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type User struct {
	FirstName      string `json:"FirstName"`
	LastName       string `json:"LastName"`
	MobileNo       int    `json:"MobileNo"`
	EmailAddress   string `json:"EmailAddress"`
	IsCarOwner     bool   `json:"IsCarOwner"`
	DriverLicense  string `json:"DriverLicense"`
	CarPlateNumber string `json:"CarPlateNumber"`
}

type Trip struct {
	PickUpLoc     string `json:"PickUpLoc"`
	AltPickUpLoc   string `json:"AltPickupLoc"`
	StartTravellingTime time.Time `json:"StartTravellingTime"`
	Destination   string `json:"Destination"`
	NoOfPassenger  int   `json:"NoOfPassenger"`
}

type Users struct {
	Users map[string]User `json:"Users"`
}

type Trips struct {
	Trips map[string]Trip `json:"Trips"`
}

var (
	db  *sql.DB
	err error
)

func main() {
	db, err = sql.Open("mysql", "Joey:101204@tcp(127.0.0.1:3306)/db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	for {
		fmt.Println(strings.Repeat("=", 10))
		fmt.Println("Carpooling management console\n",
			"1. Create new user account\n",
			"2. Update user\n",
			"3. Delete Account\n",
			"4. Publish Carpool trip\n",
			"5. Select trip\n",
			"6. Start trip\n",
			"7. Cancel trip\n",
			"9. Quit")
		fmt.Print("Enter an option: ")

		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			user()
		case 2:
			updateUser()
		case 3:
			delUser()
		case 4:
			insertTrip()
		case 5:
			GetTrips()
		case 7:
			delTrip()
		case 9:
			return
		}
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/Users/{userid}", user).Methods("GET", "DELETE", "POST", "PATCH", "PUT", "OPTIONS")
	fmt.Println("Listening at port 5000")
	log.Fatal(http.ListenAndServe(":5000", router))
}
func user(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if r.Method == "POST" {
		if body, err := ioutil.ReadAll(r.Body); err == nil {
			var data User
			fmt.Println(string(body))
			if err := json.Unmarshal(body, &data); err == nil {
				if _, ok := isExist(params["userid"]); !ok {
					fmt.Println(data)
					//users[params["userid"]] = data
					insertUser(params["userid"], data)

					w.WriteHeader(http.StatusAccepted)
				} else {
					w.WriteHeader(http.StatusConflict)
					fmt.Fprintf(w, "User ID exist")
				}
			} else {
				fmt.Println(err)
			}
		}
	} else if r.Method == "PUT" {
		if body, err := ioutil.ReadAll(r.Body); err == nil {
			var data User

			if err := json.Unmarshal(body, &data); err == nil {
				if _, ok := isExist(params["userid"]); ok {
					fmt.Println(data)
					//users[params["userid"]] = data
					updateUser(params["userid"], data)
					w.WriteHeader(http.StatusAccepted)
				} else {
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprintf(w, "User ID does not exist")
				}
			} else {
				fmt.Println(err)
			}
		}
	} else if r.Method == "PATCH" {
		if body, err := ioutil.ReadAll(r.Body); err == nil {
			var data map[string]interface{}

			if err := json.Unmarshal(body, &data); err == nil {
				if orig, ok := isExist(params["userid"]); ok {
					fmt.Println(data)

					for k, v := range data {
						switch k {
						case "First Name":
							orig.FirstName = v.(string)
						case "LastName":
							orig.LastName = v.(string)
						case "Mobile Number":
							orig.MobileNo = v.(int)
						case "Email Address":
							orig.EmailAddress = v.(string)
						case "Is Car Owner":
							orig.IsCarOwner = v.(bool)
						case "Driver License":
							orig.DriverLicense = v.(string)
						case "Car Plate Number":
							orig.CarPlateNumber = v.(string)
						}
					}
					//users[params["userid"]] = orig
					updateUser(params["userid"], orig)
					w.WriteHeader(http.StatusAccepted)
				} else {
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprintf(w, "User ID does not exist")
				}
			} else {
				fmt.Println(err)
			}
		}
	} else if val, ok := isExist(params["userid"]); ok {
		if r.Method == "DELETE" {
			fmt.Fprintf(w, params["userid"]+" Deleted")
			//delete(courses, params["userid"])
			delCourse(params["userid"])
		} else {
			json.NewEncoder(w).Encode(val)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Invalid User ID")
	}
}

func GetTrips() map[string]Trip {
	results, err := db.Query("select * from Trip")
	if err != nil {
		panic(err.Error())
	}

	var trips map[string]Trip = map[string]Trip{}

	for results.Next() {
		var t Trip
		var id string
		err = results.Scan(&id, &t.PickUpLoc, &t.AltPickUpLoc, &t.StartTravellingTime, &t.Destination, &t.NoOfPassenger)
		if err != nil {
			panic(err.Error())
		}

		trips[id] = t
	}

	return trips
}

func isExist(id string) (Trip, bool) {
	var t Trip

	result := db.QueryRow("select * from trip where id=?", id)
	err := result.Scan(&id, &t.PickUpLoc, &t.AltPickUpLoc, &t.StartTravellingTime, &t.Destination, &t.NoOfPassenger)
	if err == sql.ErrNoRows {
		return t, false
	}

	return t, true
}

func delTrip(id string) (int64, error) {
	result, err := db.Exec("delete from trip where id=?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func delUser(id string) (int64, error) {
	result, err := db.Exec("delete from user where id=?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func insertTrip(id string, t Trip) {
	_, err := db.Exec("insert into trip values(?,?,?,?,?,?)", id, t.PickUpLoc, t.AltPickUpLoc, t.StartTravellingTime, t.Destination, t.NoOfPassenger)
	if err != nil {
		panic(err.Error())
	}
}

func updateUser(id int, User) {
	_, err := db.Exec("update user set firstname=?, lastname=?, mobileno=?, emailaddress=?, iscarowner=?, driverlicense=?, carplatenumber=? where id=?", FirstName, LastName, MobileNo, EmailAddress, IsCarOwner, DriverLicense, CarPlateNumber, id)
	if err != nil {
		panic(err.Error())
	}
}

func Trips(id int) (map[string]Trip, bool) {
	results, err := db.Query("select * from Trip where id == ?", id)
	if err != nil {
		panic(err.Error())
	}

	var trips map[string]Trip = map[string]Trip{}

	for results.Next() {
		var t Trip
		var id string
		err = results.Scan(&id, &t.PickUpLoc, &t.AltPickUpLoc, &t.StartTravellingTime, &t.Destination, &t.NoOfPassenger)
		if err != nil {
			panic(err.Error())
		}

		trips[id] = t
	}

	if len(trips) == 0 {
		return trips, false
	}
	return trips, true
}
