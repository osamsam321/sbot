package main

import (
    "path/filepath"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
    "os/exec"
    "strings"

)

var debug_enabled = flag.Bool("d", false, "enable debug mode")
var openai_api_key_file_name = "../openai_api_key.txt"

func DebugPrint(msg string){
    if *debug_enabled{
        fmt.Println(msg)
    }
}
func DebugPrintf(msg string){
    if *debug_enabled{
        fmt.Printf(msg)
    }
}

func main() {
    InitFlags()
}

func getBaseDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	baseDir := filepath.Dir(ex)
	return baseDir, nil
}

func InitFlags(){

    env_help_variable_flag:=            flag.Bool("env",false,"help env for flag ")

    openai_api_key:=                    flag.String("s", "", "set your openai api key")

    user_openai_unix_prompt:=           flag.String("q", "", "Ask a basic unix shell query and get a command back")

    user_openai_general_command:=       flag.String("g", "", "Ask a general GPT question")

    stdin_filter_command:=                     flag.String("i", "", "filter or combine query with stdin")

    user_openai_unix_explain_prompt:=   flag.String("e", "", "explain what a command does")

    execute_last_command :=             flag.Bool("l", false, "run last command that exist in the local sbot history file")

    show_history_command :=             flag.Bool("y", false, "show local history")
    flag.Parse()

    if(*env_help_variable_flag){
        set_env_variable_instructions()
    } else if(len(*openai_api_key) > 0){
        SetOpenAIAPIKey(*openai_api_key)
        fmt.Print("adding your new openai api key")
    } else if(len(*user_openai_unix_prompt) > 0){
        options,err := GetOpenAIAPIBodyOptions("../prompts/openai_prompt_style_unix.json")
        if err!= nil{
            panic("openai options are empty ")
        }
        ExecuteOpenAIQuery(GetOpenAIAPIKey(),  options, *user_openai_unix_prompt)
    }else if(len(*user_openai_unix_explain_prompt) > 0){
        options,err := GetOpenAIAPIBodyOptions("../prompts/openai_prompt_style_explain.json")
        if err!= nil{
            panic("openai options are empty ")
        }
        ExecuteOpenAIQuery(GetOpenAIAPIKey(),  options, *user_openai_unix_explain_prompt)
    }else if(len(*user_openai_general_command) > 0){
        options,err := GetOpenAIAPIBodyOptions("../prompts/openai_prompt_style_general.json")
        if err!=nil{
            panic("openai options are empty ")
        }

        ExecuteOpenAIQuery(GetOpenAIAPIKey(),  options, *user_openai_general_command)

    }else if(len(*stdin_filter_command) >= 0){
	    stdin, _ := io.ReadAll(os.Stdin)
        stdin_filter_prompt:=string(stdin)+*stdin_filter_command
        options,err := GetOpenAIAPIBodyOptions("../prompts/openai_prompt_style_general.json")
        if err!=nil{
            panic("openai options are empty ")
        }
        ExecuteOpenAIQuery(GetOpenAIAPIKey(), options, stdin_filter_prompt)
    }else if(*execute_last_command){
        execute_command(last_command_in_history("../sbot_command_history.txt"))
    }else if(*show_history_command){
        ShowHistory()
    }else{
        fmt.Println("please input the correct options")
    }


}


// Creation and Sending OpenAI prompt Section
func ExecuteOpenAIQuery(api_key string,  options OpenAIBodyOptions, user_prompt string){
    new_content:=options.Messages[1].Content + user_prompt
    options.Messages[1].Content = new_content
    SendOpenAIQuery(api_key, options)
}

func ExecuteOpenAIUnixExplainQuery(api_key string,  options OpenAIBodyOptions, user_prompt string){
    new_content:=options.Messages[1].Content + user_prompt
    options.Messages[1].Content = new_content
    SendOpenAIQuery(api_key, options)
}

func GetOpenAIAPIBodyOptions(file_name string) (OpenAIBodyOptions, error){
    file_content,err := os.ReadFile(file_name)
    if(err != nil){
        fmt.Print("An error came up reading the file ", err)
    }

    //json_value, err :=json.Marshal(file_content)
    var open_ai_options OpenAIBodyOptions

    if err := json.Unmarshal(file_content, &open_ai_options); err != nil{
        panic("could not encode json "  )
        //return OpenAIBodyOptions{}, fmt.Errorf("could not Unmarshal json")
    }

    return open_ai_options, nil
}

func SendOpenAIQuery(api_key string, openai_body OpenAIBodyOptions){

    post_body, err :=json.Marshal(openai_body)
    if err != nil{
        panic("could not encode json")
    }

    response_body := bytes.NewBuffer(post_body)
    request, err:=http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", response_body)
    if err != nil{
        panic("An issue occured attempting to reach the openapi api url")
    }

    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", "Bearer " + api_key)
    client:=&http.Client{}
    response, error := client.Do(request)
    if error != nil{
        panic("there was an issue with do the request")
    }
    defer request.Body.Close()

    response_bytes,err:= io.ReadAll(response.Body)
    if err != nil{
        panic("could not interprit the response data")
    }

    var response_json ChatCompletion

    err = json.Unmarshal(response_bytes, &response_json)

    if err != nil{
        panic(err)
    }
    command:=response_json.Choices[0].Message.Content
    DebugPrint("response from openai " + string(response_bytes))
    fmt.Println()
    fmt.Println(command)
    WriteAppendToLocalCommandHistory("../sbot_command_history.txt", command, 700)
}

// Seting and Getting openai_api_key
func SetOpenAIAPIKey(openai_api_key string){
    api_key_content := []byte (openai_api_key)
    err:= os.WriteFile(openai_api_key_file_name, api_key_content, 700)

    if err != nil{
        panic("failed to create or write to the openai_api_key file")
    }
}

func GetOpenAIAPIKey() (string){

    value, err := os.ReadFile(openai_api_key_file_name)
    if(err != nil){
        panic(err)
    }
    return string(value)
}
func execute_command(command string)(string, string, error){
    //!TODO sanitize any special chracters
    //re := regexp.MustCompile(`^[a-zA-Z0-9\s\.\|\'\"\-\/\_]+$`)
    settings_file_content,err:=os.ReadFile("../setting.json")
    if err != nil{
        panic(err)
    }

    var common_settings CommonSettings
    err=json.Unmarshal(settings_file_content, &common_settings )
    if err != nil{
        panic(err)
    }
    // Check if the command matches the allowed pattern
    // read from settings file and unmarshal as shown
    // TODO! Check for the matching characters
   // if !re.MatchString(command) {
   //     fmt.Println("command contains dangerous special characters ")
   //     return "", "", nil
   // }

    dangerous_commands:= common_settings.DangerousCommands;

    for i:=0;i< len(dangerous_commands);i++ {
       if strings.Contains(command, dangerous_commands[i]) {
           DebugPrint("command: " + command + " dangerous command: " + dangerous_commands[i] + " i " + string(i))
           fmt.Println("your command is a 'Dangerous command type'. Please enable this in the setting file or try a different command")
           return "", "", nil
       }
    }

    // now execute command
    const shell_to_use = "bash"
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    DebugPrint("going to run the command with the follow components " + shell_to_use + " " + "-c " + command)
    cmd :=exec.Command(shell_to_use, "-c", command)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    fmt.Println("Command to execute >>> " + cmd.String())
    err = cmd.Run()
    if err != nil{
        DebugPrint("An error occured while running command")
        fmt.Print("An error occured >>> ")
        fmt.Println(cmd.Stderr)
    }
    if cmd.Stdout != nil{
        //fmt.Println()
        fmt.Println(cmd.Stdout)
    }
    return stdout.String(), stderr.String(), err
}

func last_command_in_history(file_name string) string {
    last_command:=""
    bytesRead,err := os.ReadFile(file_name)
    if(err != nil){
        panic(err)
    }
    fileContent := string(bytesRead)
    lines := strings.Split(fileContent, "\n")
    last_command = lines[len(lines) - 1]
    if len(last_command) <= 0 {
        panic("command history is empty")
    }
    return last_command
}

func set_env_variable_instructions(){
fmt.Println(`
To set variable only for current shell:

VARNAME="my value"

To set it for current shell and all processes started from current shell:

export VARNAME="my value"      # shorter, less portable version

To set it permanently for all future bash sessions add such line to your .bashrc file in your $HOME directory.

To set it permanently, and system wide (all users, all processes) add set variable in /etc/environment:

sudo -H gedit /etc/environment

This file only accepts variable assignments like:

VARNAME="my value"

Do not use the export keyword here.

Use source ~/.bashrc in your terminal for the changes to take place immediately.
`)

}

func WriteAppendToLocalCommandHistory(file_name string, content_passed string, perm int){
    content_from_file, err:=os.ReadFile(file_name)
    space:= "\n"
    if err != nil{
        println("could read not read from the history file. Please see if it exists")
    }
    if(len(content_from_file) <= 0){
        space=""
    }
    content_to_add := string(content_from_file) + space + content_passed
    os.WriteFile(file_name, []byte(content_to_add), 700)
}

func ShowHistory(){
    content, err:= os.ReadFile("../sbot_command_history.txt")
    if err != nil{
        panic(err)
    }
    fmt.Println(string(content))
}

type OpenAIBodyOptions struct {
    Model    string           `json:"model"`
    Messages []OpenAIMessages `json:"messages"`
    Temperature      float64 `json:"temperature"`
    MaxTokens        int     `json:"max_tokens"`
    TopP             float64 `json:"top_p"`
    FrequencyPenalty float64 `json:"frequency_penalty"`
    PresencePenalty  float64 `json:"presence_penalty"`
}

type OpenAIMessages struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
type ChatCompletion struct {
    ID                string    `json:"id"`
    Object            string    `json:"object"`
    Created           int64     `json:"created"`
    Model             string    `json:"model"`
    Choices           []Choice  `json:"choices"`
    Usage             Usage     `json:"usage"`
    SystemFingerprint *string   `json:"system_fingerprint"` // Use a pointer to handle null values
}

type Choice struct {
    Index        int      `json:"index"`
    Message      Message  `json:"message"`
    Logprobs     *string  `json:"logprobs"` // Use a pointer to handle null values
    FinishReason string   `json:"finish_reason"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
type CommonSettings struct{
    AllowDangerousCommands bool     `json:"allow_dangerous_commands"`
    DangerousCommands      []string `json:"dangerous_commands"`
}
