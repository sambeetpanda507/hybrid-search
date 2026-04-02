package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const SEMANTIC_SEARCH string = "symantic_search"
const LEXICAL_SEARCH string = "lexical_search"
const HYBRID_SEARCH string = "hybrid_search"

type symanticSearchPayload struct {
	Query string `json:"query"`
	Page  string `json:"page"`
	Limit string `json:"limit"`
}

type lexicalSearchPayload struct {
	Query string `json:"query"`
	Page  *int   `json:"page"`
	Limit *int   `json:"limit"`
}

type hybridSearchPayload struct {
	Query string `json:"query"`
	Page  *int   `json:"page"`
	Limit *int   `json:"limit"`
}

type requestBody struct {
	Action                string                `json:"action"`
	SymanticSearchPayload symanticSearchPayload `json:"symanticSearchPayload"`
	LexicalSearchPayload  lexicalSearchPayload  `json:"lexicalSearchPayload"`
	HybridSearchPayload   hybridSearchPayload   `json:"hybridSearchPayload"`
}

func handleSemanticSearch(w http.ResponseWriter, payload symanticSearchPayload) {
	baseURL := "http://search-service/symantic-search"
	queryParams := url.Values{}
	queryParams.Add("page", payload.Page)
	queryParams.Add("limit", payload.Limit)
	encodedQuery := queryParams.Encode()
	url := fmt.Sprintf("%s?%s", baseURL, encodedQuery)
	requestPayload := struct {
		Query string `json:"query"`
	}{
		Query: payload.Query,
	}

	requestPayloadBytes, err := json.MarshalIndent(requestPayload, "", "\t")
	if err != nil {
		errorJSON(w, "Unable to marshal the payload", err, http.StatusInternalServerError)
		return
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(requestPayloadBytes))
	if err != nil {
		errorJSON(w, err.Error(), err, res.StatusCode)
		return
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		errorJSON(w, "Something went worng", nil, res.StatusCode)
		return
	}

	var responsePayload interface{}
	if err := json.NewDecoder(res.Body).Decode(&responsePayload); err != nil {
		errorJSON(w, err.Error(), err, http.StatusInternalServerError)
		return
	}

	responseJSON(w, "ok", responsePayload, res.StatusCode)
}

func handleLexicalSearch(w http.ResponseWriter, payload lexicalSearchPayload) {
	baseURL := "http://search-service/lexical-search"
	requestBody, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		errorJSON(w, err.Error(), err, http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		errorJSON(w, err.Error(), nil, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errorJSON(w, "Something went worng", resp.Body, resp.StatusCode)
		return
	}

	var responseBody interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		errorJSON(w, err.Error(), err, http.StatusInternalServerError)
		return
	}

	responseJSON(w, "ok", responseBody, resp.StatusCode)
}

func handleHybridSearch(w http.ResponseWriter, payload hybridSearchPayload) {
	baseURL := "http://search-service/hybrid-search"
	requestBody, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		errorJSON(w, err.Error(), err, http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		errorJSON(w, err.Error(), nil, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errorJSON(w, "Something went worng", resp.Body, resp.StatusCode)
		return
	}

	var responseBody interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		errorJSON(w, err.Error(), err, http.StatusInternalServerError)
		return
	}

	responseJSON(w, "ok", responseBody, resp.StatusCode)
}

func responseJSON(w http.ResponseWriter, msg string, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	var responseBody struct {
		Message string `json:"message"`
		Data    any    `json:"data,omitempty"`
	}

	responseBody.Message = msg
	responseBody.Data = data
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func (app *Config) handleBroker(w http.ResponseWriter, r *http.Request) {
	var requestBody requestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		errorJSON(w, "Unable to parse request payload", err, http.StatusInternalServerError)
		return
	}

	switch requestBody.Action {
	case SEMANTIC_SEARCH:
		handleSemanticSearch(w, requestBody.SymanticSearchPayload)
	case LEXICAL_SEARCH:
		handleLexicalSearch(w, requestBody.LexicalSearchPayload)
	case HYBRID_SEARCH:
		handleHybridSearch(w, requestBody.HybridSearchPayload)
	default:
		errorJSON(w, "Invalid action", nil, http.StatusBadRequest)
	}
}

func errorJSON(w http.ResponseWriter, msg string, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	var responseBody struct {
		Message string `json:"message"`
		Data    any    `json:"data,omitempty"`
	}

	responseBody.Message = msg
	responseBody.Data = data
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(responseBody)
}
