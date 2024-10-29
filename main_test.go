package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleTemperature(t *testing.T) {
	tests := []struct {
		name           string
		cep            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Formato invalido para o CEP",
			cep:            "123",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "CEP invalido",
		},
		{
			name:           "CEP Não existe",
			cep:            "99999999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "CEP Não encontrado.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/temperature/"+tt.cep, nil)
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
