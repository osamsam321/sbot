package main

import (
	"bytes"
    "bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"github.com/joho/godotenv"
)

var debug_enabled = flag.Bool("d", false, "enable debug mode")

func main() {
    InitFlags()
}

func InitFlags(){
    DebugPrint("starting init")
    DebugPrint("Base dir is " + GetBaseDir())

    user_query:=                        flag.String("q", "", "add your query here")

    user_selected_prompt:=              flag.String("p", "", "your prompt alias")

    execute_last_command :=             flag.Bool("l", false, "run last command that exist in the local sbot history file")

    show_history_command :=             flag.Bool("y", false, "show local history")

    flag.Parse()

    if(*execute_last_command){
        execute_command(last_command_in_history(filepath.Join(GetBaseDir(), "sbot_command_history.txt")))
    }else if *show_history_command{
        ShowHistory()
    }else if len(*user_query) > 0 || StdinExist(){

            if StdinExist(){
                input, err := getStdinAsString()
                if err != nil{
                    os.Exit(1)
                }
                result:= input + *user_query
                user_query=&result
            }

            prompts, err:= getAndLoadPrompts()
            starting_prompt:=""
            if err != nil{
                fmt.Println("could not load any prompts. Exiting!")
                os.Exit(1)
            }

            // start execution here now

            if len(*user_selected_prompt) > 0{
                prompt_aliases:= []string{}
                prompt_aliases,err=getPromptAliases(prompts)
                if len(prompt_aliases) <= 0{
                    println("you have no prompts in the prompt folder. ")
                    os.Exit(1)
                }
                if len(prompt_aliases) > 0 {
                    if(slices.Contains(prompt_aliases, *user_selected_prompt)){
                        //filepath, err:= getABSPathFromAlias(prompts, prompt_alias)
                        prompt, err:=getPromptFromAlias(prompts, *user_selected_prompt)
                        if err != nil{
                            println("Unable to get prompt. Exiting!")
                            os.Exit(1)
                        }
                        DebugPrint("Aliases found and being used " + prompt.Alias)
                        // this will combine the initialial content field with the user prompt using the template type specfied in the prompt file
                        for i, msg:=range prompt.ChatRequestBody.Messages{
                            if msg.Role=="user"{
                               starting_prompt=msg.Content
                               complete_prompt:=strings.ReplaceAll(starting_prompt, prompt.PlaceholderType, *user_query)
                               prompt.ChatRequestBody.Messages[i].Content=complete_prompt
                               DebugPrint("The complete prompt " + complete_prompt)
                            }
                        }
                        if starting_prompt == ""{
                            fmt.Println("You are either missing the user field or user prompt is empty. Please fix your prompt. ")
                            os.Exit(1)

                        }
                        SendOpenRouterPostRequest(GetAPIKey(), prompt.ChatRequestBody, true)
                    }
                    if err!= nil{
                        fmt.Println("options are empty from site")
                    }
                }

            }else{
                DebugPrint("Prompt option not supplied. Using default prompt by selecting the prompt with the lowest id value")
                prompt:=PromptOption{}
                smallest_val:=prompts[0].ID;
                for _, p := range prompts{
                    if p.ID <= smallest_val{
                       DebugPrint("prompt id " + string(p.ID))
                       prompt = p
                    }
                }
                DebugPrint("The select prompt alias is " + prompt.Alias)
                for i, msg:=range prompt.ChatRequestBody.Messages{
                    if msg.Role=="user"{
                        starting_prompt=msg.Content
                        complete_prompt:=strings.ReplaceAll(starting_prompt, prompt.PlaceholderType, *user_query)
                        prompt.ChatRequestBody.Messages[i].Content=complete_prompt
                        DebugPrint("The complete prompt " + complete_prompt)
                    }
                }
                if starting_prompt == ""{
                    fmt.Println("You are either missing the user field or user prompt is empty. Please fix your prompt. ")
                    os.Exit(1)
                }
                SendOpenRouterPostRequest(GetAPIKey(), prompt.ChatRequestBody, true)
            }
        }else{
            fmt.Println("Please input the correct options.")
        }
}

func printList(element []string){
    for _, element:=range element{
        println(element)
    }
}

func getPromptFromAlias(prompts []PromptOption, prompt_alias string) (PromptOption, error){
    for _, prompt:= range prompts{
        if prompt.Alias == prompt_alias{
            return prompt, nil
        }
    }
    return PromptOption{}, fmt.Errorf("Prompt type was not obtainable from prompt alias")
}

func getAndLoadPrompts() ([]PromptOption, error) {
    DebugPrint("Now attempting to load prompt files")
    prompts := []PromptOption{}
    prompt_dir := "prompts"
    files, err := os.ReadDir(filepath.Join(GetBaseDir(), prompt_dir))
    if err != nil {
        abs_path, err := filepath.Abs(prompt_dir)
        if err != nil {
            fmt.Println("Could not read directory " + abs_path)
            return nil, err
        }
        fmt.Printf("There was an issue trying to read the directory %s\n", abs_path)
        return nil, err
    }

    for _, prompt_file := range files {
        if prompt_file.Type().IsRegular() {
            prompt_abs_path := filepath.Join(GetBaseDir(), prompt_dir, prompt_file.Name())
            DebugPrint("Prompt file path being loaded: " + prompt_abs_path)

            // Read the file content
            file_content, err := os.ReadFile(prompt_abs_path)
            if err != nil {
                fmt.Printf("There was an issue attempting to open %s\n", prompt_abs_path)
                continue // Skip to the next file if there's an error reading this one
            }

            // Initialize a new PromptOption for each file
            var prompt_content PromptOption
            err = json.Unmarshal(file_content, &prompt_content)
            if err != nil {
                fmt.Printf("There is a syntax issue with this JSON prompt file: %s\n", prompt_abs_path)
                fmt.Println("Checking other files...")
                continue // Skip to the next file if there's a JSON syntax issue
            }

            prompts = append(prompts, prompt_content)
        }
    }

    return prompts, nil
}

func getPromptAliases(prompts []PromptOption) ([]string, error){
    prompt_aliases:=[]string{}
    for _, prompt := range prompts{
        DebugPrint("Prompt alias found " + prompt.Alias)
        if len(prompt.Alias) > 0{
            prompt_aliases = append(prompt_aliases, prompt.Alias)
        }
    }
    return prompt_aliases,nil
}

func SendOpenRouterPostRequest(api_key string, chat_request_body ChatRequestBody, add_to_history bool){

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
            DebugPrint("Full error log from response body: " + string(responseBytes))
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
        DebugPrint("response from site >>" + string(response_bytes))
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
    if !common_settings.AllowDangerousCommands{
        for i:=0;i< len(dangerous_commands);i++ {
           if strings.Contains(command, dangerous_commands[i]) {
               DebugPrint("command: " + command + " dangerous command: " + dangerous_commands[i] + " i " + string(i))
               fmt.Println("your command is a 'Dangerous command type'. Please enable this in the setting file or try a different command")
               return "", "", nil
           }
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

func GetAPIKey() (string){
    err := godotenv.Load(filepath.Join(GetBaseDir(),".env"))
    DebugPrint("grabbing api key from " + filepath.Join(GetBaseDir(), ".env"))
    if err != nil{
        fmt.Println("Error loading .env file")
    }
    return os.Getenv("OPENROUTER_API_KEY")
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
	return base_dir
}

func getStdinAsString() (string, error){
    scanner := bufio.NewScanner(os.Stdin)
    if scanner.Scan() {
        input := scanner.Text()
        return input, nil
    }

    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading from stdin:", err)
        return "", nil
    }
    return "", nil
}

func StdinExist() bool{
    fi, err := os.Stdin.Stat()
    if err != nil {
        fmt.Println(err)
    }
    if fi.Mode()&os.ModeCharDevice == 0 {
		DebugPrint("There is stdin input available.")
        return true
	}
    return false
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

type PromptOption struct {
    ChatRequestBody   ChatRequestBody   `json:"chat_request_body"`
    Alias             string            `json:"alias"`
    ID                int16             `json:"id"`
    PlaceholderType   string            `json:"placeholder_type"`
}
type ChatRequestBody struct {
    Model              string           `json:"model"`
    Messages           []ChatAIMessages `json:"messages"`
    Temperature        float64          `json:"temperature"`
    MaxTokens          int              `json:"max_tokens"`
    TopP               float64          `json:"top_p"`
    FrequencyPenalty   float64          `json:"frequency_penalty"`
    PresencePenalty    float64          `json:"presence_penalty"`
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
    Shell                  string   `json:"shell"`
    DangerousCommands      []string `json:"dangerous_commands"`
}

