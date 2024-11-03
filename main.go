package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
)

type ViaCEPResponse struct {
	CEP        string `json:"cep"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Erro       bool   `json:"erro"`
}

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/temperatura/", handleTemperature)
	log.Printf("Servidor iniciado na port: %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleTemperature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}

	cep := r.URL.Path[len("/temperatura/"):]

	if !isValidCEP(cep) {
		respondWithError(w, http.StatusUnprocessableEntity, "CEP invalido")
		return
	}

	location, err := getLocationByCEP(cep)
	if err != nil {
		if err == ErrCEPNotFound {
			respondWithError(w, http.StatusNotFound, "CEP nao encontrado")
			return
		}
		respondWithError(
			w,
			http.StatusInternalServerError,
			"erro ao obter localidade",
		)
		return
	}

	temp, err := getTemperature(location)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"erro ao obter temperatura",
		)
		return
	}

	response := TemperatureResponse{
		TempC: temp,
		TempF: celsiusToFahrenheit(temp),
		TempK: celsiusToKelvin(temp),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func isValidCEP(cep string) bool {
	matched, _ := regexp.MatchString(`^\d{8}$`, cep)
	return matched
}

var (
	ErrCEPNotFound = fmt.Errorf("CEP nao encontrado")
	viaCEPBaseURL  = "https://viacep.com.br"
)

func getLocationByCEP(cep string) (string, error) {
	url := fmt.Sprintf("%s/ws/%s/json/", viaCEPBaseURL, cep)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Erro ao consultar API: %v", err)
	}
	defer resp.Body.Close()

	var viaCEPResp ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viaCEPResp); err != nil {
		return "", fmt.Errorf("Erro ao decodificar resposta: %v", err)
	}

	if viaCEPResp.Erro || viaCEPResp.Localidade == "" {
		return "", ErrCEPNotFound
	}

	return fmt.Sprintf("%s,%s", viaCEPResp.Localidade, viaCEPResp.UF), nil
}

func getTemperature(location string) (float64, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	url := fmt.Sprintf(
		"http://api.weatherapi.com/v1/current.json?key=%s&q=%s",
		apiKey,
		location,
	)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return 0, err
	}

	return weatherResp.Current.TempC, nil
}

func celsiusToFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}

func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}
