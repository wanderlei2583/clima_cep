package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleTemperature(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cep := strings.TrimPrefix(r.URL.Path, "/ws/")
			cep = strings.TrimSuffix(cep, "/json/")

			switch cep {
			case "99999999":
				json.NewEncoder(w).Encode(ViaCEPResponse{Erro: true})
			case "01001000":
				json.NewEncoder(w).Encode(ViaCEPResponse{
					CEP:        "01001000",
					Localidade: "São Paulo",
					UF:         "SP",
				})
			default:
				http.Error(w, "Not found", http.StatusNotFound)
			}
		}),
	)
	defer ts.Close()

	originalViaCEPURL := viaCEPBaseURL
	viaCEPBaseURL = ts.URL
	defer func() { viaCEPBaseURL = originalViaCEPURL }()

	tests := []struct {
		name           string
		cep            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid CEP Format",
			cep:            "123",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "CEP invalido",
		},
		{
			name:           "Non-existent CEP",
			cep:            "99999999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "CEP nao encontrado",
		},
		{
			name:           "Valid CEP",
			cep:            "01001000",
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/temperatura/"+tt.cep, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleTemperature)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf(
					"handler retornou codigo de resposta inesperado: obteve %v desejado %v",
					status,
					tt.expectedStatus,
				)
			}

			var errorResp ErrorResponse
			err = json.NewDecoder(rr.Body).Decode(&errorResp)
			if err != nil {
				t.Fatal(err)
			}

			if errorResp.Message != tt.expectedError {
				t.Errorf(
					"handler retornou um erro inesperado: obteve %v desejado %v",
					errorResp.Message,
					tt.expectedError,
				)
			}
		})
	}
}

func TestIsValidCEP(t *testing.T) {
	tests := []struct {
		cep      string
		expected bool
	}{
		{"12345678", true},
		{"1234567", false},
		{"123456789", false},
		{"1234567a", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isValidCEP(tt.cep)
		if result != tt.expected {
			t.Errorf(
				"isValidCEP(%s) = %v; want %v",
				tt.cep,
				result,
				tt.expected,
			)
		}
	}
}
