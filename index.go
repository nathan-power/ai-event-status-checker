package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	N           int           `json:"n,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   OpenAIUsage            `json:"usage"`
}

type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	processRecords("urls.csv")
}

func processRecords(csvFileName string) {
	file, err := os.Open(csvFileName)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %s", err)
	}
	defer file.Close()

	totalRecords, err := lineCounter(csvFileName)
	if err != nil {
		log.Fatalf("Failed to count lines in CSV file: %s", err)
	}

	if totalRecords <= 1 { // 1 line for the header
		log.Fatalf("CSV file does not contain any records to process.")
	}

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		log.Fatalf("Failed to read headers from CSV file: %s", err)
	}

	nameIndex := findIndex(headers, "name")
	if nameIndex == -1 {
		log.Fatalf("CSV does not contain required 'name' column")
	}

	urlIndex := findIndex(headers, "url")
	if urlIndex == -1 {
		log.Fatalf("CSV does not contain required 'url' column")
	}

	totalRecords -= 1 // Adjust for header line
	processEachRecord(reader, nameIndex, urlIndex, totalRecords)
}

func processEachRecord(reader *csv.Reader, nameIndex, urlIndex, totalRecords int) {
	spinner := []string{"|", "/", "-", "\\"}
	var processedRecords int

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read a record from CSV file: %s", err)
		}

		processedRecords++
		displayProgress(processedRecords, totalRecords, spinner)

		name := record[nameIndex]
		url := record[urlIndex]

		pageContent, err := fetchPageContent(url)
		if err != nil {
			fmt.Printf("\r\033[KFailed to fetch data from %s: %s\n", url, err)
			continue
		}

		responseData := queryModel(pageContent)

		if strings.TrimSpace(responseData) == "yes" {
			fmt.Printf("\r\033[KEvent at \"%s\" is currently accepting applications.\n", name)
		} else if strings.TrimSpace(responseData) == "no" {
			fmt.Printf("\r\033[KEvent at \"%s\" is not accepting applications.\n", name)
		} else {
			fmt.Printf("\r\033[KCould not determine if event at \"%s\" is accepting applications.\n", name)
		}
	}

	fmt.Printf("\r\033[KProcessing complete.\n")
}

func displayProgress(current, total int, spinner []string) {
	if total <= 0 {
		fmt.Printf("\r\033[KInvalid progress calculation: total records must be greater than 0.\n")
		return
	}
	spinnerIndex := (current - 1) % len(spinner)
	fmt.Printf("\r\033[KProcessing... %d%% complete %s ", (current*100)/total, spinner[spinnerIndex])
}

func queryModel(pageContent string) string {
	prompt := fmt.Sprintf("Based on the following webpage content, determine if the event is currently accepting applications. Answer only with 'yes' or 'no'. Content:\n%s", pageContent)

	request := ChatCompletionRequest{
		Model:       "gpt-4o-mini",
		Messages:    []ChatMessage{{Role: "user", Content: prompt}},
		MaxTokens:   5,
		Temperature: 0,
		N:           1,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Failed to encode request to JSON: %s", err)
	}

	return postData(jsonData)
}

func lineCounter(fileName string) (int, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return 0, err
	}
	return bytes.Count(data, []byte("\n")), nil
}

func findIndex(headers []string, column string) int {
	for i, header := range headers {
		if header == column {
			return i
		}
	}
	return -1
}

func fetchPageContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(bodyBytes), nil
}

func postData(data []byte) string {
	url := "https://api.openai.com/v1/chat/completions"
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send HTTP request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("OpenAI API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Fatalf("Failed to decode response: %s", err)
	}

	if len(res.Choices) == 0 {
		log.Fatal("No choices returned from OpenAI API")
	}

	return strings.TrimSpace(res.Choices[0].Message.Content)
}
