package main

import (
	"encoding/json"
	"fmt"
	//"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"mongo-tools/vendor/src/gopkg.in/mgo.v2"
	//"mongo-tools/vendor/src/gopkg.in/mgo.v2/bson"
	//"mgo.v2"
	//"mgo.v2/bson"
	"bytes"
	"encoding/binary"
	"httprouter"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const UBER_ACCESS_TOKEN string = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsInJlcXVlc3QiLCJoaXN0b3J5Il0sInN1YiI6ImQ1NTcwMDYyLWY3YmEtNDEzZi1iNjFmLTA5M2ZlNWU2M2M1YSIsImlzcyI6InViZXItdXMxIiwianRpIjoiMDM5MTZmMzgtMGI2MS00NTk2LTkyOTktODU3NGFiODlkYzQ0IiwiZXhwIjoxNDUwNTY2NzA2LCJpYXQiOjE0NDc5NzQ3MDUsInVhY3QiOiI5WUNBbjRDMjF2Ulg5MGN4SUJQZ3drcFVpb2R2azUiLCJuYmYiOjE0NDc5NzQ2MTUsImF1ZCI6InNCQlRibUNqWS1lZ1ozZTJOY1d0NGt1R0h4aEJCLXY2In0.H4R01tAYuVbx2OKVcWxz39sNkchK2FRvUciBIImxeNu6VaDhlnxzFB3IlnluVA9PjR7mpewnveWznLJd4AGo4Q9ACghl3vnxDmL8Y5K9Z8ykaKWaolR9QnH-swVIxEBHV0-aA019LmPqxw9Tol2ajg3EcfftBjYHgEePaQjJcMcAdHROHK-GZQs9EiFJRuDQsLfm5dDtcDLIGYBHIe_iMYGrh2a122RRS5Gl9QuMp9gNO0tRtRMdGT1JL2BoKV-VdInNsf35QAxaTTTU4SajVltp-HvxnVDu6dtkJQHK3VelzbixjkrFrbieqZ-f7gbsoPZZXYNGCA4V5kqbpr92Ng"
const UBER_SERVER_TOKEN string = "Token ogLziO28JWvnYhmXSD7yAQJNxBK-JtF18r0gE2Li"
const UBER_URL string = "https://sandbox-api.uber.com/v1"

type TripRequest struct {
	LocationIds            []string `json:"location_ids"`
	StartingFromLocationID string   `json:"starting_from_location_id"`
}

type TripDetails struct {
	ID                        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Status                    string        `json:"status"`
	StartingFromLocationID    string        `json:"starting_from_location_id,omitempty"`
	NextDestinationLocationID string        `json:"next_destination_location_id,omitempty"`
	BestRouteLocationIds      []string      `json:"best_route_location_ids"`
	TotalUberCosts            int           `json:"total_uber_costs"`
	TotalUberDuration         int           `json:"total_uber_duration"`
	TotalDistance             float64       `json:"total_distance"`
	UberWaitTime              string        `json:"uber_wait_time_eta,omitempty"`
}

type LocationReq struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type LocationRes struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Address    string        `json:"address"`
	City       string        `json:"city"`
	State      string        `json:"state"`
	Zip        string        `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

type RideRequest struct {
	ProductID      string `json:"product_id"`
	StartLatitude  string `json:"start_latitude"`
	StartLongitude string `json:"start_longitude"`
	EndLatitude    string `json:"end_latitude"`
	EndLongitude   string `json:"end_longitude"`
}

type RideResponse struct {
	Driver          interface{} `json:"driver"`
	Eta             int         `json:"eta"`
	Location        interface{} `json:"location"`
	RequestID       string      `json:"request_id"`
	Status          string      `json:"status"`
	SurgeMultiplier int         `json:"surge_multiplier"`
	Vehicle         interface{} `json:"vehicle"`
}

type UberProducts struct {
	Products []struct {
		Capacity     int    `json:"capacity"`
		Description  string `json:"description"`
		DisplayName  string `json:"display_name"`
		Image        string `json:"image"`
		PriceDetails struct {
			Base            float64 `json:"base"`
			CancellationFee int     `json:"cancellation_fee"`
			CostPerDistance float64 `json:"cost_per_distance"`
			CostPerMinute   float64 `json:"cost_per_minute"`
			CurrencyCode    string  `json:"currency_code"`
			DistanceUnit    string  `json:"distance_unit"`
			Minimum         float64 `json:"minimum"`
			ServiceFees     []struct {
				Fee  float64 `json:"fee"`
				Name string  `json:"name"`
			} `json:"service_fees"`
		} `json:"price_details"`
		ProductID string `json:"product_id"`
	} `json:"products"`
}

type UberPriceEstimates struct {
	Prices []struct {
		CurrencyCode    string  `json:"currency_code"`
		DisplayName     string  `json:"display_name"`
		Distance        float64 `json:"distance"`
		Duration        int     `json:"duration"`
		Estimate        string  `json:"estimate"`
		HighEstimate    int     `json:"high_estimate"`
		LowEstimate     int     `json:"low_estimate"`
		ProductID       string  `json:"product_id"`
		SurgeMultiplier int     `json:"surge_multiplier"`
	} `json:"prices"`
}

type GeoLoc struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type GoogleLocationRes struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

var collection *mgo.Collection
var locRes LocationRes

const (
	timeout = time.Duration(time.Second * 100)
)

func connectMongo() {
	uri := "mongodb://arpit9691:Arpit#9691@ds043694.mongolab.com:43694/location_db"
	//uri := "mongodb://nipun:nipun@ds045464.mongolab.com:45464/db2"
	ses, err := mgo.Dial(uri)

	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
	} else {
		ses.SetSafe(&mgo.Safe{})
		collection = ses.DB("location_db").C("test")
		//collection = ses.DB("db2").C("qwerty")
	}
}

func getGoogleLoc(address string) (geoLocation GeoLoc) {

	client := http.Client{Timeout: timeout}
	url := fmt.Sprintf("http://maps.google.com/maps/api/geocode/json?address=%s", address)
	res, err := client.Get(url)
	if err != nil {
		fmt.Errorf("Can't read Google API: %v", err)
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&geoLocation)
	if err != nil {
		fmt.Errorf("Error in decoding the Google: %v", err)
	}
	return geoLocation
}

func getLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id := bson.ObjectIdHex(p.ByName("locationID"))
	err := collection.FindId(id).One(&locRes)
	if err != nil {
		fmt.Printf("error finding a doc %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(200)
	json.NewEncoder(rw).Encode(locRes)
}

func addLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var tempLocReq LocationReq
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&tempLocReq)
	if err != nil {
		fmt.Errorf("Error in decoding the Input: %v", err)
	}
	address := tempLocReq.Address + " " + tempLocReq.City + " " + tempLocReq.State + " " + tempLocReq.Zip
	address = strings.Replace(address, " ", "%20", -1)

	locationDetails := getGoogleLoc(address)

	locRes.ID = bson.NewObjectId()
	locRes.Address = tempLocReq.Address
	locRes.City = tempLocReq.City
	locRes.Name = tempLocReq.Name
	locRes.State = tempLocReq.State
	locRes.Zip = tempLocReq.Zip
	locRes.Coordinate.Lat = locationDetails.Results[0].Geometry.Location.Lat
	locRes.Coordinate.Lng = locationDetails.Results[0].Geometry.Location.Lng

	err = collection.Insert(locRes)
	if err != nil {
		fmt.Printf("Can't insert document: %v\n", err)
	}

	err = collection.FindId(locRes.ID).One(&locRes)
	if err != nil {
		fmt.Printf("error finding a doc %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(201)
	json.NewEncoder(rw).Encode(locRes)
}

func updateLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var tempLocRes LocationRes
	var locRes LocationRes
	id := bson.ObjectIdHex(p.ByName("locationID"))
	err := collection.FindId(id).One(&locRes)
	if err != nil {
		fmt.Printf("error While Searching document %v\n")
	}
	tempLocRes.Name = locRes.Name
	tempLocRes.Address = locRes.Address
	tempLocRes.City = locRes.City
	tempLocRes.State = locRes.State
	tempLocRes.Zip = locRes.Zip
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&tempLocRes)

	if err != nil {
		fmt.Errorf("Error in Input: %v", err)
	}

	address := tempLocRes.Address + " " + tempLocRes.City + " " + tempLocRes.State + " " + tempLocRes.Zip
	address = strings.Replace(address, " ", "%20", -1)
	locationDetails := getGoogleLoc(address)
	tempLocRes.Coordinate.Lat = locationDetails.Results[0].Geometry.Location.Lat
	tempLocRes.Coordinate.Lng = locationDetails.Results[0].Geometry.Location.Lng
	err = collection.UpdateId(id, tempLocRes)
	if err != nil {
		fmt.Printf("Got error while updating documment %v\n")
	}

	err = collection.FindId(id).One(&locRes)
	if err != nil {
		fmt.Printf("Got error while searching document %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(201)
	json.NewEncoder(rw).Encode(locRes)
}

func deleteLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id := bson.ObjectIdHex(p.ByName("locationID"))
	err := collection.RemoveId(id)
	if err != nil {
		fmt.Printf("Got error while deleting document %v\n")
	}
	rw.WriteHeader(200)
}
func AddTrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	fmt.Println("Hello Arpit!!!")
	tripRequest := TripRequest{}
	locationDetails := LocationRes{}
	tripDetails := TripDetails{}
	locations := make(map[string]LocationRes)
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&tripRequest)
	if err != nil {
		fmt.Errorf("Error in Input JSON: %v", err)
	}

	url := fmt.Sprintf("http://localhost:8080/locations/%s", tripRequest.StartingFromLocationID)
	client := http.Client{Timeout: timeout}

	res, err := client.Get(url)
	if err != nil {
		fmt.Errorf("Not able to read localhost LocationsAPI: %v", err)
	}
	defer res.Body.Close()
	decoder = json.NewDecoder(res.Body)

	err = decoder.Decode(&locationDetails)
	if err != nil {
		fmt.Errorf("Error in Location JSON: %v", err)
	}
	fmt.Println(locationDetails.Coordinate.Lat)

	locations[tripRequest.StartingFromLocationID] = locationDetails

	//Push the rest of the IDs' location details to map by iteration
	for _, value := range tripRequest.LocationIds {
		url = fmt.Sprintf("http://localhost:8080/locations/%s", value)
		client = http.Client{Timeout: timeout}

		res, err = client.Get(url)
		if err != nil {
			fmt.Errorf("Not able to read localhost LocationsAPI: %v", err)
		}
		defer res.Body.Close()
		decoder = json.NewDecoder(res.Body)

		err = decoder.Decode(&locationDetails)
		if err != nil {
			fmt.Errorf("Error in Location JSON: %v", err)
		}
		locations[value] = locationDetails
	}
	startID := tripRequest.StartingFromLocationID
	startLat := locations[startID].Coordinate.Lat
	originLat := startLat
	startLng := locations[startID].Coordinate.Lng
	originLng := startLng
	nextID := tripRequest.StartingFromLocationID
	var lowPrice int
	var duration int
	var distance float64
	minPrice := 10000
	minduration := 0
	mindistance := 0.0
	totalCost := 0
	totalUberDuration := 0
	totalDistance := 0.0
	locationOrder := 0

	for len(locations) > 1 {
		for key, value := range locations {
			if key != startID {

				lowPrice, duration, distance = PriceEstimate(startLat, startLng, value.Coordinate.Lat, value.Coordinate.Lng)

				if lowPrice < minPrice {
					minPrice = lowPrice
					minduration = duration
					mindistance = distance
					nextID = key

				}
			}
		}
		totalCost += lowPrice
		totalUberDuration += minduration
		totalDistance += mindistance
		delete(locations, startID)
		startID = nextID
		startLat = locations[startID].Coordinate.Lat
		startLng = locations[startID].Coordinate.Lng
		tripRequest.LocationIds[locationOrder] = nextID
		locationOrder++
		minPrice = 1000000.0
		//	minduration := 0
		//	mindistance := 0.0
	}

	lowPrice, duration, distance = PriceEstimate(originLat, originLng, locations[nextID].Coordinate.Lat, locations[nextID].Coordinate.Lng)
	totalCost += lowPrice
	totalUberDuration += duration
	totalDistance += distance

	tripDetails.BestRouteLocationIds = tripRequest.LocationIds
	tripDetails.StartingFromLocationID = tripRequest.StartingFromLocationID
	tripDetails.Status = "planning..."
	tripDetails.TotalDistance = distance
	tripDetails.TotalUberDuration = duration
	tripDetails.TotalUberCosts = totalCost

	tripDetails.ID = bson.NewObjectId()
	err = collection.Insert(tripDetails)
	if err != nil {
		fmt.Printf("Not able to insert document: %v\n", err)
	}
	err = collection.FindId(tripDetails.ID).One(&tripDetails)
	if err != nil {
		fmt.Printf("Got error While Searching document %v\n")
	}

	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(200)
	json.NewEncoder(rw).Encode(tripDetails)
	fmt.Println(tripRequest.LocationIds)
}

func GetUberProductID(latitude float64, longitude float64) string {

	uberDetails := UberProducts{}
	url := fmt.Sprintf("%s/products?latitude=%f&longitude=%f", UBER_URL, latitude, longitude)
	client := http.Client{Timeout: timeout}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", UBER_SERVER_TOKEN)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		fmt.Errorf("Not able to read UBER API: %v", err)
	}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&uberDetails)
	if err != nil {
		fmt.Errorf("Error in Google JSON: %v", err)
	}
	return uberDetails.Products[0].ProductID
}

func PriceEstimate(start_latitude float64, start_longitude float64, end_latitude float64, end_longitude float64) (int, int, float64) {

	uberPriceEstimates := UberPriceEstimates{}
	url := fmt.Sprintf("%s/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f", UBER_URL, start_latitude, start_longitude, end_latitude, end_longitude)
	client := http.Client{Timeout: timeout}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", UBER_SERVER_TOKEN)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		fmt.Errorf("Not able to read UBER API: %v", err)
	}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&uberPriceEstimates)
	if err != nil {
		fmt.Errorf("Error in Google JSON: %v", err)
	}

	return uberPriceEstimates.Prices[0].LowEstimate, uberPriceEstimates.Prices[0].Duration, uberPriceEstimates.Prices[0].Distance

}

func UberRideRequest(start_latitude float64, start_longitude float64, end_latitude float64, end_longitude float64) string {

	rideRequest := RideRequest{}
	rideResponse := RideResponse{}
	rideRequest.ProductID = GetUberProductID(start_latitude, start_longitude)
	rideRequest.StartLatitude = fmt.Sprintf("%.6f", start_latitude)
	rideRequest.StartLongitude = fmt.Sprintf("%.6f", start_longitude)
	rideRequest.EndLatitude = fmt.Sprintf("%.6f", end_latitude)
	rideRequest.EndLongitude = fmt.Sprintf("%.6f", end_longitude)

	url := fmt.Sprintf("%s/requests", UBER_URL)
	client := http.Client{Timeout: timeout}
	b, err := json.Marshal(rideRequest)
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, &b)
	req, _ := http.NewRequest("POST", url, buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", UBER_ACCESS_TOKEN)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		fmt.Errorf("Error in UBER_RideRequest API: %v", err)
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&rideResponse)
	if err != nil {
		fmt.Errorf("Error in UBER_RideRequest JSON: %v", err)
	}
	eta := strconv.Itoa(rideResponse.Eta)
	return eta

}

func GetTrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	tripDetails := TripDetails{}
	id := bson.ObjectIdHex(p.ByName("trip_id"))
	err := collection.FindId(id).One(&tripDetails)
	if err != nil {
		fmt.Printf("got error While Searching document %v\n")
	}
	if tripDetails.ID == "" {
		fmt.Fprintf(rw, "LocationID not valid")
	} else {
		tripDetails.NextDestinationLocationID = ""
		tripDetails.UberWaitTime = ""
		rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.WriteHeader(200)
		json.NewEncoder(rw).Encode(tripDetails)
	}
}

func UpdateTrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	tripDetails := TripDetails{}
	locationDetails := LocationRes{}
	id := bson.ObjectIdHex(p.ByName("trip_id"))
	err := collection.FindId(id).One(&tripDetails)
	if err != nil {
		fmt.Printf("got error while searching a trip %v\n")
	}
	currentLocationID := tripDetails.StartingFromLocationID
	if tripDetails.NextDestinationLocationID == tripDetails.StartingFromLocationID {
		tripDetails.Status = "trip completed"
		tripDetails.NextDestinationLocationID = ""
		tripDetails.StartingFromLocationID = ""
		tripDetails.UberWaitTime = ""
	} else {

		if tripDetails.Status == "requesting..." {
			if len(tripDetails.BestRouteLocationIds) > 1 {
				currentLocationID = tripDetails.BestRouteLocationIds[0]
				x := tripDetails.BestRouteLocationIds[1:len(tripDetails.BestRouteLocationIds)]
				tripDetails.BestRouteLocationIds = x
				tripDetails.NextDestinationLocationID = tripDetails.BestRouteLocationIds[0]
			} else {
				tripDetails.BestRouteLocationIds = nil
				currentLocationID = tripDetails.NextDestinationLocationID
				tripDetails.NextDestinationLocationID = tripDetails.StartingFromLocationID
			}
		} else if tripDetails.Status == "planning..." {
			tripDetails.NextDestinationLocationID = tripDetails.BestRouteLocationIds[0]
			tripDetails.Status = "requesting..."
		}

		url := fmt.Sprintf("http://localhost:8080/locations/%s", currentLocationID)
		client := http.Client{Timeout: timeout}

		res, err := client.Get(url)
		if err != nil {
			fmt.Errorf("Cannot read localhost LocationsAPI: %v", err)
		}
		defer res.Body.Close()
		decoder := json.NewDecoder(res.Body)

		err = decoder.Decode(&locationDetails)
		if err != nil {
			fmt.Errorf("Error in Google Location JSON: %v", err)
		}
		startLat := locationDetails.Coordinate.Lat
		startLng := locationDetails.Coordinate.Lng

		url = fmt.Sprintf("http://localhost:8080/locations/%s", tripDetails.NextDestinationLocationID)
		client = http.Client{Timeout: timeout}

		res, err = client.Get(url)
		if err != nil {
			fmt.Errorf("Error in localhost Google LocationsAPI: %v", err)
		}
		defer res.Body.Close()
		decoder = json.NewDecoder(res.Body)

		err = decoder.Decode(&locationDetails)
		if err != nil {
			fmt.Errorf("Error in Google Location JSON: %v", err)
		}
		tripDetails.UberWaitTime = UberRideRequest(startLat, startLng, locationDetails.Coordinate.Lat, locationDetails.Coordinate.Lng)

		if len(tripDetails.BestRouteLocationIds) == 0 {
			tripDetails.NextDestinationLocationID = tripDetails.StartingFromLocationID
		}

	}

	//update the request in database
	err = collection.UpdateId(id, tripDetails)
	if err != nil {
		fmt.Printf("got error while updating document %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(200)
	json.NewEncoder(rw).Encode(tripDetails)

	fmt.Println("Hiiii!!!!!")
}

func main() {
	mux := httprouter.New()
	mux.GET("/locations/:locationID", getLocation)
	mux.POST("/locations", addLocation)
	mux.PUT("/locations/:locationID", updateLocation)
	mux.DELETE("/locations/:locationID", deleteLocation)
	mux.GET("/trips/:trip_id", GetTrip)
	mux.POST("/trips", AddTrip)
	mux.PUT("/trips/:trip_id/request", UpdateTrip)
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	connectMongo()
	server.ListenAndServe()
}
