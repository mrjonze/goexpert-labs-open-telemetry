package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
)

type Response struct {
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
	City  string  `json:"city"`
}

var postData struct {
	Cep string `json:"cep"`
}

func main() {
	http.HandleFunc("/", SearchCepHandler)
	http.ListenAndServe(":8080", nil)
}

func SearchCepHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error reading request body"))
		return
	}

	err = json.Unmarshal(body, &postData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error unmarshalling request body"))
		return
	}

	if postData.Cep == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cep parameter is required"))
		return
	}

	validate := regexp.MustCompile(`^[0-9]{8}$`)
	if !validate.MatchString(postData.Cep) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("invalid zipcode"))
		return
	}

	temperature, err := CallServiceB(postData.Cep)

	if err != nil {
		errorStr := err.Error()
		if errorStr == "can not find zipcode" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errorStr))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error while searching for cep: " + errorStr))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorStr := err.Error()
		w.Write([]byte("error while searching for temperature: " + errorStr))
		return
	}
	if temperature != nil {
		json.NewEncoder(w).Encode(temperature)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("can not find temperature"))
	}
}

func CallServiceB(cep string) (*Response, error) {
	req, err := http.Get("http://goapp-service-b:8081/?cep=" + cep)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	if req.StatusCode == http.StatusNotFound {
		return nil, errors.New("can not find zipcode")
	}

	res, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var data Response
	err = json.Unmarshal(res, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
