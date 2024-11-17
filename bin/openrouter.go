package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func ExecuteOpenRouterRequest(api_key string, chat_request_body ChatRequestBody) (ChatCompletion){
    post_body, err :=json.Marshal(chat_request_body)
    if err != nil{
        fmt.Println("could not encode json")
    }

    response_body := bytes.NewBuffer(post_body)
    request, err:=http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", response_body)
    if err != nil{
        fmt.Println("An issue occured attempting to reach the api url")
    }

    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", "Bearer " + api_key)
    client:=&http.Client{}
    response, error := client.Do(request)
    if error != nil{
        fmt.Println("There was an error in the response >> ", error);
    }
    defer request.Body.Close()

        if response.StatusCode != 200{
            fmt.Printf("\nThere was an error from the response. Status code is %d\n\n", response.StatusCode)
            fmt.Printf("Please check your API key or any other common issues \n\n")
            responseBytes, err := io.ReadAll(response.Body)
            if err != nil {
                fmt.Println("Could not interpret the response data")
            }
            DebugPrintln("Full error log from response body: " + string(responseBytes), ErrorLog)
            os.Exit(1)
        }

        response_bytes,err:= io.ReadAll(response.Body)
        if err != nil{
            fmt.Println("could not interprit the response data")
        }

        var response_json ChatCompletion
        err = json.Unmarshal(response_bytes, &response_json)

        if err != nil{
            fmt.Println(err)
        }
        DebugPrintln("Status of request >> " + response.Status, InfoLog);
        DebugPrintln("response from site >>" + string(response_bytes), InfoLog)
        return response_json
}


type ChatTemplate struct {
    ChatRequestBody ChatRequestBody `json:"chat_request_body"`
    Name           string          `json:"name"`
    PlaceholderType string          `json:"placeholder_type"`
}

type ChatTemplateList struct {
    Templates []ChatTemplate `json:"chat_templates"`
}

type ChatRequestBody struct {
    Model            string    `json:"model"`
    Messages         []Message `json:"messages"`
    Temperature      float64   `json:"temperature"`
    MaxTokens        int       `json:"max_tokens"`
    TopP             float64   `json:"top_p"`
    FrequencyPenalty float64   `json:"frequency_penalty"`
    PresencePenalty  float64   `json:"presence_penalty"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
type ChatAIMessages struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
type ChatCompletion struct {
    ID                string    `json:"id"`
    Object            string    `json:"object"`
    Created           int64     `json:"created"`
    Model             string    `json:"model"`
    Choices           []Choice  `json:"choices"`
    Error             *Error     `json:"error"`
    Usage             Usage     `json:"usage"`
    SystemFingerprint *string   `json:"system_fingerprint"` // Use a pointer to handle null values
}
type Choice struct {
    Index        int      `json:"index"`
    Message      Message  `json:"message"`
    Logprobs     *string  `json:"logprobs"` // Use a pointer to handle null values
    FinishReason string   `json:"finish_reason"`
}
type Error struct {
    Code    int       `json:"code"`
    Message string    `json:"message"`
    Metadata Metadata `json:"metadata"`
}
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
type Metadata struct {
	Raw          string `json:"raw"`
	ProviderName string `json:"provider_name"`
}
