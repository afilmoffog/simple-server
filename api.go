package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
)

type Time struct {
	T string `json:"time"`
}

type Sort struct {
	Array []int `json:"array"`
	Uniq  bool  `json:"uniq"`
}

func getTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var t Time
	t.T = time.Now().UTC().Format("15:04:05")
	json.NewEncoder(w).Encode(t)
}

func postArray(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var array Sort
	_ = json.NewDecoder(r.Body).Decode(&array)
	if len(array.Array) > 100 || len(array.Array) == 0 {
		returnCode400(w, r)
	} else {
		sortArray(array.Array)
		if array.Uniq {
			array.Array = checkUniq(array.Array)
		}
		mapA := map[string][]int{"array": array.Array}
		mapB, _ := json.Marshal(mapA)
		fmt.Fprintf(w, string(mapB))
	}
}

func sortArray(array []int) {
	end := len(array) - 1
	for {
		if end == 0 {
			break
		}
		for i := 0; i < len(array)-1; i++ {
			if array[i] > array[i+1] {
				array[i], array[i+1] = array[i+1], array[i]
			}
		}
		end--
	}
	return
}

func checkUniq(array []int) []int {
	encountered := map[int]bool{}
	result := []int{}
	for v := range array {
		if encountered[array[v]] == true {
		} else {
			encountered[array[v]] = true
			result = append(result, array[v])
		}
	}
	return result
}

func getWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	city := r.URL.Query().Get("city")
	if city == "" {
		returnCode400(w, r)
	} else {
		resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city + "&APPID=a8228c3bc9d8aa285bcd3da9b9a127dc")
		if err != nil {
			fmt.Println(w, err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(w, err)
		}
		code := gjson.Get(string(body), "cod").Int()
		if code == 404 {
			returnCode404(w, r)
		} else {
			value := gjson.Get(string(body), "main.temp")
			mapA := map[string]int64{"temp": value.Int() - 273}
			mapB, _ := json.Marshal(mapA)
			fmt.Fprintf(w, string(mapB))
		}
	}
}
func returnCode400(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(" Error 400"))
}

func returnCode404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(" Error 404"))
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/now", getTime).Methods("GET")
	router.HandleFunc("/api/sort", postArray).Methods("POST")
	router.HandleFunc("/api/weather", getWeather).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
