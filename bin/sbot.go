package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/joho/godotenv"
)

var debug_enabled = flag.Bool("d", false, "enable debug mode")

func main() {
    InitFlags()
}

func InitFlags(){

    user_openai_unix_prompt:=           flag.String("q", "", "Ask a basic unix shell query and get a command back")

    user_openai_general_command:=       flag.String("g", "", "Ask a general GPT question")

    stdin_filter_command:=              flag.String("i", "", "filter or combine query with stdin")

    user_openai_unix_explain_prompt:=   flag.String("e", "", "explain what a command does")

    execute_last_command :=             flag.Bool("l", false, "run last command that exist in the local sbot history file")

    show_history_command :=             flag.Bool("y", false, "show local history")

    flag.Parse()

    if len(*user_openai_unix_prompt) > 0{
        options,err := GetOpenAIAPIBodyOptions(filepath.Join(GetBaseDir(), "prompts/openai_prompt_style_unix.json"))
        if err!= nil{
           fmt.Println("openai options are empty ")
        }
        ExecuteOpenAIQuery(GetOpenAIAPIKey(),  options, *user_openai_unix_prompt, true)
    }else if len(*user_openai_unix_explain_prompt) > 0{
        options,err := GetOpenAIAPIBodyOptions(filepath.Join(GetBaseDir(), "prompts/openai_prompt_style_explain.json"))
        if err!= nil{
            fmt.Println("openai options are empty ")
        }
        ExecuteOpenAIQuery(GetOpenAIAPIKey(),  options, *user_openai_unix_explain_prompt, false)
    }else if len(*user_openai_general_command) > 0{
        options,err := GetOpenAIAPIBodyOptions(filepath.Join(GetBaseDir(),"prompts/openai_prompt_style_general.json"))
        if err!=nil{
            fmt.Println("openai options are empty ")
        }

        ExecuteOpenAIQuery(GetOpenAIAPIKey(),  options, *user_openai_general_command, false)
    }else if len(*stdin_filter_command) >= 0 && StdinExist(){
        DebugPrint("stdin value was added")
	    stdin, _ := io.ReadAll(os.Stdin)
        stdin_filter_prompt:=string(stdin)+*stdin_filter_command
        options,err := GetOpenAIAPIBodyOptions(filepath.Join(GetBaseDir(),"prompts/openai_prompt_style_general.json"))
        if err!=nil{
            fmt.Println("openai options are empty ")
        }
        ExecuteOpenAIQuery(GetOpenAIAPIKey(), options, stdin_filter_prompt, false)

    }else if(*execute_last_command){
        execute_command(last_command_in_history(filepath.Join(GetBaseDir(), "sbot_command_history.txt")))
    }else if *show_history_command{
        ShowHistory()
    }else{
        fmt.Println("please input the correct options")
    }
}

// Creation and Sending OpenAI prompt Section
func ExecuteOpenAIQuery(api_key string,  options OpenAIBodyOptions, user_prompt string, add_to_history bool){
    new_content:=options.Messages[1].Content + user_prompt
    options.Messages[1].Content = new_content
    SendOpenAIQuery(api_key, options, add_to_history)
}

func ExecuteOpenAIUnixExplainQuery(api_key string,  options OpenAIBodyOptions, user_prompt string){
    new_content:=options.Messages[1].Content + user_prompt
    options.Messages[1].Content = new_content
    SendOpenAIQuery(api_key, options, false)
}

func SendOpenAIQuery(api_key string, openai_body OpenAIBodyOptions, add_to_history bool){

    post_body, err :=json.Marshal(openai_body)
    if err != nil{
        fmt.Println("could not encode json")
    }

    response_body := bytes.NewBuffer(post_body)
    request, err:=http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", response_body)
    if err != nil{
        fmt.Println("An issue occured attempting to reach the openapi api url")
    }

    request.Header.Add("Content-Type", "application/json")
    request.Header.Add("Authorization", "Bearer " + api_key)
    client:=&http.Client{}
    response, error := client.Do(request)
    if error != nil{
        fmt.Println("There was an error in openai response >> ", error);
    }
    defer request.Body.Close()

        if response.StatusCode != 200{
            fmt.Printf(" \n There was an error from OPENAI. Status code is %d\n\n", response.StatusCode)
            fmt.Printf("Please check your API key or any other common issues \n\n");
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
        command :=response_json.Choices[0].Message.Content
        DebugPrint("Status of request >> " + response.Status);
        DebugPrint("response from openai >>" + string(response_bytes))

        fmt.Println(command)
        if add_to_history{
            WriteAppendToLocalCommandHistory(filepath.Join(GetBaseDir(), "sbot_command_history.txt"), command, 700)
        }


}

func execute_command(command string)(string, string, error){
    //!TODO sanitize any special chracters
    //re := regexp.MustCompile(`^[a-zA-Z0-9\s\.\|\'\"\-\/\_]+$`)
    settings_file_content,err:=os.ReadFile(filepath.Join(GetBaseDir(), "setting.json"))
    if err != nil{
        fmt.Println(err)
    }

    var common_settings CommonSettings
    err=json.Unmarshal(settings_file_content, &common_settings )
    if err != nil{
        fmt.Println(err)
    }
    // Check if the command matches the allowed pattern
    // read from settings file and unmarshal as shown
    // TODO! Check for the matching characters
   // if !re.MatchString(command) {
   //     fmt.Println("command contains dangerous special characters ")
   //     return "", "", nil
   // }

    dangerous_commands:= common_settings.DangerousCommands

    for i:=0;i< len(dangerous_commands);i++ {
       if strings.Contains(command, dangerous_commands[i]) {
           DebugPrint("command: " + command + " dangerous command: " + dangerous_commands[i] + " i " + string(i))
           fmt.Println("your command is a 'Dangerous command type'. Please enable this in the setting file or try a different command")
           return "", "", nil
       }
    }

    // now execute command
    shell_to_use := common_settings.Shell
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    DebugPrint("going to run the command with the follow components " + shell_to_use + " " + "-c " + command)
    cmd :=exec.Command(shell_to_use, "-c", command)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    fmt.Println("Command to execute >>> " + cmd.String())
    err = cmd.Run()
    if err != nil{
        fmt.Print("An error occured >>> ")
        fmt.Println(cmd.Stderr)
    }
    if cmd.Stdout != nil{
        //fmt.Println()
        fmt.Println(cmd.Stdout)
    }
    return stdout.String(), stderr.String(), err
}
// getting stuff from files
func GetOpenAIAPIBodyOptions(file_path string) (OpenAIBodyOptions, error){
    file_content,err := os.ReadFile(file_path)
    if(err != nil){
        fmt.Print("An error came up reading the file ", file_path, err);
    }
    //json_value, err :=json.Marshal(file_content)
    var open_ai_options OpenAIBodyOptions

    if err := json.Unmarshal(file_content, &open_ai_options); err != nil{
        fmt.Println("could not encode json "  )
        //return OpenAIBodyOptions{}, fmt.Errorf("could not Unmarshal json")
    }

    return open_ai_options, nil
}

func GetOpenAIAPIKey() (string){
    err := godotenv.Load(filepath.Join(GetBaseDir(),".env"))

    if err != nil{
        fmt.Println("Error loading .env file")
    }
    return os.Getenv("OPENAI_API_KEY")
}


func last_command_in_history(file_name string) string {
    DebugPrint("checking last command");
    last_command:=""
    bytesRead,err := os.ReadFile(file_name)
    if(err != nil){
        fmt.Println(err)
    }
    fileContent := string(bytesRead)
    lines := strings.Split(fileContent, "\n")
    last_command = lines[len(lines) - 1]
    if len(last_command) <= 0 {
        fmt.Println("command history is empty")
    }

    DebugPrint("last command in history is " + last_command);
    return last_command
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
    content, err:= os.ReadFile(filepath.Join(GetBaseDir(), "sbot_command_history.txt" ) )
    if err != nil{
        fmt.Println(err)
    }
    fmt.Println(string(content))
}


// util functions
func GetBaseDir() string {
    file_executable_path, err := os.Executable()
    bin_dir := filepath.Dir(file_executable_path)
	if err != nil {
        fmt.Println(err)
	}
    base_dir := filepath.Dir(bin_dir)
    DebugPrint("base dir is " + base_dir )
	return base_dir
}


func StdinExist() bool{
    fi, err := os.Stdin.Stat()
    if err != nil {
        fmt.Println(err)
        return false;
    }
    if fi.Mode() & os.ModeNamedPipe == 0 {
        return false
    } else {
        return true
    }
}

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


//struct sections
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
    Shell string                    `json:"shell"`
    DangerousCommands      []string `json:"dangerous_commands"`
}

