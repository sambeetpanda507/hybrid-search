package server

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pgvector/pgvector-go"
	"github.com/sambeetpanda507/advance-search/search-service/internal/database"
	"gorm.io/gorm"
)

func readCSVFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}

func stringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func parseRows(rows [][]string) []database.Job {
	workers := 5
	job := make(chan []string, len(rows))
	results := make(chan database.Job)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for row := range job {
			var parsedRow database.Job
			parsedRow.Title = row[0]
			exp, err := stringToInt(row[1])
			if err != nil {
				continue
			}

			parsedRow.Experience = exp
			parsedRow.EducationLevel = row[2]
			skillCount, err := stringToInt(row[3])
			if err != nil {
				continue
			}

			parsedRow.SkillsCount = skillCount
			parsedRow.Industry = row[4]
			parsedRow.CompanySize = row[5]
			parsedRow.Location = row[6]
			parsedRow.RemoteWork = row[7]
			cert, err := stringToInt(row[8])
			if err != nil {
				continue
			}

			parsedRow.Certifications = cert
			salary, err := stringToInt(row[9])
			if err != nil {
				continue
			}

			parsedRow.Salary = salary
			results <- parsedRow
		}
	}

	for range workers {
		wg.Add(1)
		go worker()
	}

	for _, row := range rows {
		job <- row
	}
	close(job)

	go func() {
		wg.Wait()
		close(results)
	}()

	output := []database.Job{}
	for r := range results {
		output = append(output, r)
	}

	return output
}

func (app *Config) handleReadCSV(w http.ResponseWriter, r *http.Request) {
	rows, err := readCSVFile("internal/assets/job_salary_prediction_dataset.csv")
	if err != nil {
		errorJSON(w, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	parsedData := parseRows(rows)
	responseJSON(w, "Data parsed successfully", parsedData, http.StatusOK)
}

func storeRows(db *gorm.DB, rows [][]string) error {
	workers := 5
	batchSize := 200
	maxRows := min(10000, len(rows))

	for start := 1; start < maxRows; start += batchSize {
		end := min(start+batchSize, len(rows))
		job := make(chan []string, len(rows))
		result := make(chan database.Job)
		var wg sync.WaitGroup

		worker := func() {
			defer wg.Done()
			for row := range job {
				var parsedRow database.Job
				parsedRow.Title = row[0]
				exp, err := stringToInt(row[1])
				if err != nil {
					continue
				}

				parsedRow.Experience = exp
				parsedRow.EducationLevel = row[2]
				skillCount, err := stringToInt(row[3])
				if err != nil {
					continue
				}

				parsedRow.SkillsCount = skillCount
				parsedRow.Industry = row[4]
				parsedRow.CompanySize = row[5]
				parsedRow.Location = row[6]
				parsedRow.RemoteWork = row[7]
				cert, err := stringToInt(row[8])
				if err != nil {
					continue
				}

				parsedRow.Certifications = cert
				salary, err := stringToInt(row[9])
				if err != nil {
					continue
				}

				parsedRow.Salary = salary
				parsedRow.CreatedAt = time.Now()
				parsedRow.UpdatedAt = time.Now()
				query := fmt.Sprintf("Job Title: %s., Education: %s., Industry: %s., Company Size: %s.", parsedRow.Title, parsedRow.EducationLevel, parsedRow.Industry, parsedRow.CompanySize)
				embeddings, err := text2Vector(query)
				if err != nil {
					continue
				}

				parsedRow.Embedding = pgvector.NewVector(embeddings)
				result <- parsedRow
			}
		}

		// Start the workers
		for range workers {
			wg.Add(1)
			go worker()
		}

		// Send job to worker
		for i := start; i < end; i++ {
			if i == 0 {
				continue
			}

			row := rows[i]
			job <- row
		}
		close(job)

		// Consume the result
		go func() {
			wg.Wait()
			close(result)
		}()

		parsedRows := []database.Job{}
		for r := range result {
			parsedRows = append(parsedRows, r)
		}

		// Write db
		if err := db.CreateInBatches(&parsedRows, 1000).Error; err != nil {
			return err
		}
	}

	return nil
}

func (app *Config) handleStoreCSV(w http.ResponseWriter, r *http.Request) {
	rows, err := readCSVFile("internal/assets/job_salary_prediction_dataset.csv")
	if err != nil {
		errorJSON(w, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if err := storeRows(app.db, rows); err != nil {
		errorJSON(w, err.Error(), err, http.StatusInternalServerError)
		return
	}

	responseJSON(w, "Data stored successfully", nil, http.StatusCreated)
}

func text2Vector(text string) ([]float32, error) {
	url := os.Getenv("OLLAMA_EMBEDDING_URL")
	if url == "" {
		return nil, errors.New("Ollama embedding API url missing")
	}

	requestBody := struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}{
		Model:  "all-minilm",
		Prompt: text,
	}

	reqBytes, err := json.MarshalIndent(requestBody, "", "\t")
	if err != nil {
		return nil, errors.New(err.Error())
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Unable to generate vector")
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Embedding, nil
}

func (app *Config) handleGetVector(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Text string `json:"text"`
	}

	// Read the body
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		errorJSON(w, "Error reading request body", err, http.StatusInternalServerError)
		return
	}

	embedding, err := text2Vector(requestBody.Text)
	if err != nil {
		errorJSON(w, err.Error(), err, http.StatusInternalServerError)
		return
	}

	respBody := struct {
		Embedding []float32 `json:"embedding"`
	}{
		Embedding: embedding,
	}

	responseJSON(w, "Embedding generated", respBody, http.StatusCreated)
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

func isDBEmpty(db *gorm.DB) (bool, error) {
	var count int64
	result := db.Model(&database.Job{}).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count == 0, nil
}

func (app *Config) handleSymanticSearch(w http.ResponseWriter, r *http.Request) {
	isEmpty, err := isDBEmpty(app.db)
	if err != nil {
		errorJSON(w, err.Error(), err, http.StatusInternalServerError)
		return
	}

	if isEmpty {
		errorJSON(w, "Database is empty. Hit /store-csv to populated the database", nil, http.StatusNotFound)
		return
	}

	var requestPayload struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		errorJSON(w, "Unable to parse request payload", err, http.StatusInternalServerError)
		return
	}

	var pageStr = r.URL.Query().Get("page")
	page, err := stringToInt(pageStr)
	if err != nil {
		page = 0
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := stringToInt(limitStr)
	if err != nil {
		limit = 10
	}

	offset := page * limit
	embedding, err := text2Vector(requestPayload.Query)
	if err != nil {
		errorJSON(w, "Unable to generate vector embedding from query", err, http.StatusInternalServerError)
		return
	}

	sqlQuery := `
		WITH
			SEARCH AS (
				SELECT
					ID,
					TITLE,
					EXPERIENCE,
					EDUCATION_LEVEL,
					SKILLS_COUNT,
					INDUSTRY,
					EMBEDDING <=> ? AS DISTANCE
				FROM
					JOBS
			)
		SELECT
			*
		FROM
			SEARCH
		WHERE
			DISTANCE < 5
		ORDER BY
			DISTANCE ASC	
		LIMIT ?
		OFFSET ?;
	`

	type ResponsePayload struct {
		ID             int     `json:"id"`
		Title          string  `json:"title"`
		Experience     int     `json:"experience"`
		EducationLevel string  `json:"educationLevel"`
		SkillsCount    int     `json:"skillCount"`
		Industry       string  `json:"industry"`
		DISTANCE       float32 `json:"distance" gorm:"column:distance"`
	}

	var responsePayload []ResponsePayload
	result := app.db.Raw(sqlQuery, pgvector.NewVector(embedding), limit, offset).Scan(&responsePayload)
	if result.Error != nil {
		errorJSON(w, result.Error.Error(), nil, http.StatusInternalServerError)
		return
	}

	responseJSON(w, "Ok", responsePayload, http.StatusOK)
}

func (app *Config) handleLexicalSearch(w http.ResponseWriter, r *http.Request) {
	// Parse the search query from the r.Body
	var requestPayload struct {
		Query string `json:"query"`
		Page  *int   `json:"page"`
		Limit *int   `json:"Limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		errorJSON(w, "Error decoding request", err, http.StatusInternalServerError)
		return
	}

	// Validate
	queryStr := strings.Trim(requestPayload.Query, " ")
	if queryStr == "" {
		errorJSON(w, "Query can't left empty", nil, http.StatusUnprocessableEntity)
		return
	}

	var page int
	if requestPayload.Page == nil {
		page = 0
	} else {
		page = *requestPayload.Page
	}

	var limit int
	if requestPayload.Limit == nil {
		limit = 10
	} else {
		limit = *requestPayload.Limit
	}

	// Query DB
	type ResponesPayload struct {
		Title          string  `json:"title"`
		EducationLevel string  `json:"educationLevel"`
		Industry       string  `json:"industry"`
		CompanySize    string  `json:"companySize"`
		Location       string  `json:"location"`
		Score          float64 `json:"score"`
	}

	var responsePayload []ResponesPayload
	sqlQuery := `
		select 
			title,
			education_level,
			industry,
			company_size,
			location,
			pdb.score(id) as score
		from 
			jobs
		where
			title ||| ? or
			education_level ||| ? or
			industry ||| ? or
			company_size ||| ? or
			location ||| ?
		order by
			score desc
		limit 
			?
		offset
			?;
	`

	result := app.db.Raw(
		sqlQuery,
		queryStr,
		queryStr,
		queryStr,
		queryStr,
		queryStr,
		limit,
		page*limit,
	).Scan(&responsePayload)

	if result.Error != nil {
		errorJSON(w, result.Error.Error(), result.Error, http.StatusInternalServerError)
		return
	}

	// Return respones
	responseJSON(w, "Ok", responsePayload, http.StatusOK)
}

func (app *Config) handleHybridSearch(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var requestBody struct {
		Query string `json:"query"`
		Page  *int   `json:"page"`
		Limit *int   `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		errorJSON(w, "Unalbe to parse request body", err, http.StatusInternalServerError)
		return
	}

	var p, l int
	if requestBody.Page == nil {
		p = 0
	}

	if requestBody.Limit == nil {
		l = 10
	}

	fmt.Println(p, l)

	// Generate embedding
	embedding, err := text2Vector(requestBody.Query)
	if err != nil {
		errorJSON(w, "Unable to generate vector embedding from query", err, http.StatusInternalServerError)
		return
	}

	// Sql query
	sqlQuery := `
		WITH
			FULLTEXT AS (
				SELECT
					ID,
					PDB.SCORE (ID) AS SCORE,
					ROW_NUMBER() OVER (
						ORDER BY
							PDB.SCORE (ID) DESC
					) AS R
				FROM
					JOBS
				WHERE
					TITLE ||| ?
					OR EDUCATION_LEVEL ||| ?
					OR INDUSTRY ||| ?
					OR COMPANY_SIZE ||| ?
					OR LOCATION ||| ?
				LIMIT
					20
			),
			SEMANTIC AS (
				SELECT
					ID,
					ROW_NUMBER() OVER (
						ORDER BY
							EMBEDDING <=> ? ASC
					) AS R
				FROM
					JOBS
				LIMIT
					20
			),
			RRF AS (
				SELECT
					ID,
					1.0 / (60 + R) AS S
				FROM
					FULLTEXT
				UNION ALL
				SELECT
					ID,
					1.0 / (60 + R) AS S
				FROM
					SEMANTIC
				ORDER BY
					S DESC
			)
		SELECT
			J.ID,
			J.TITLE,
			J.EDUCATION_LEVEL,
			J.INDUSTRY,
			J.COMPANY_SIZE,
			J.LOCATION,
			SUM(RRF.S) AS SCORE
		FROM
			RRF
			INNER JOIN JOBS AS J USING (ID)
		GROUP BY
			J.ID,
			RRF.S
		ORDER BY
			RRF.S DESC;
	`

	type ResponseBody struct {
		ID             int     `json:"id"`
		Title          string  `json:"title"`
		EducationLevel string  `json:"educationLevel"`
		Industry       string  `json:"industry"`
		CompanySize    string  `json:"companySize"`
		Location       string  `json:"location"`
		Score          float64 `json:"score"`
	}

	var responseBody []ResponseBody

	// Query database
	result := app.db.Raw(
		sqlQuery,
		requestBody.Query,
		requestBody.Query,
		requestBody.Query,
		requestBody.Query,
		requestBody.Query,
		pgvector.NewVector(embedding),
	).Scan(&responseBody)

	if result.Error != nil {
		errorJSON(w, result.Error.Error(), result.Error, http.StatusInternalServerError)
		return
	}

	responseJSON(w, "Ok", responseBody, http.StatusOK)
}
