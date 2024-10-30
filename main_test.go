package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupMockViaCEP() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cep := r.URL.Path[len("/ws/"):]
			cep = cep[:8]

			w.Header().Set("Content-Type", "application/json")

			if cep == "99999999" {
				json.NewEncoder(w).Encode(ViaCEPResponse{Erro: true})
				return
			}

			json.NewEncoder(w).Encode(ViaCEPResponse{
				CEP:        "12345678",
				Localidade: "São Paulo",
				UF:         "SP",
				Erro:       false,
			})
		}),
	)
}

func TestHandleTemperature(t *testing.T) {
	mockServer := setupMockViaCEP()
	defer mockServer.Close()

	tests := []struct {
		name           string
		method         string
		cep            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Método inválido",
			method:         http.MethodPost,
			cep:            "12345678",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Método não permitido",
		},
		{
			name:           "Formato inválido para o CEP",
			method:         http.MethodPost,
			cep:            "123",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "CEP inválido",
		},
		{
			name:           "CEP Não existe",
			method:         http.MethodPost,
			cep:            "99999999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "CEP não encontrado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/temperatura/"+tt.cep, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleTemperature)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			var errorResp ErrorResponse
			err = json.NewDecoder(rr.Body).Decode(&errorResp)
			if err != nil {
				t.Fatal(err)
			}

			if errorResp.Message != tt.expectedError {
				t.Errorf(
					"handler returned unexpected error message: got %v want %v",
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
