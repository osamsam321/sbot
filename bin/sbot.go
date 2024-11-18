package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var debug_enabled =                    flag.Bool("d", false, "enable debug mode")
var execute_current_command =          flag.Bool("x", false, "execute command")
var execute_last_command =             flag.Bool("l", false, "run last command that exist in the local sbot history file")
var selected_chat_template_name=       flag.String("t", "", "your chat template name")
var user_query=                        flag.String("q", "", "add your query here")
var show_history_command =             flag.Bool("y", false, "show local history")
var all_chat_template_names =          flag.Bool("a", false, "list all chat template names")

func main() {
    handleFlags()
}

func handleFlags(){
    DebugPrintln("starting init", InfoLog)
    DebugPrintln("Base dir is " + GetBaseDir(), InfoLog)

    flag.Parse()

    if(*execute_last_command){
        executeCommand(lastCommandInHistory(filepath.Join(GetBaseDir(), "sbot_command_history.txt")))
    }else if *show_history_command{
        showHistory()
    }else if *all_chat_template_names{
        listChatTemplateNames()
    }else if len(*user_query) > 0 || StdinExist(){

            if StdinExist(){
                DebugPrintln("Stdin found", InfoLog)
                input, err := GetStdinAsString()
                if err != nil{
                    fmt.Println("There was an issue collecting standard input.")
                    os.Exit(1)
                }
                result:= input + *user_query
                user_query=&result
            }
            handleUserQuery(user_query)
    }else{
        fmt.Println("Please input the correct options.")
    }
}

func handleUserQuery(user_query *string){
    chat_templates, err:= getAndLoadChatTemplates()
    starting_prompt:=""
    if err != nil{
        fmt.Println("could not load any prompts. Exiting!")
        os.Exit(1)
    }
    // start execution here now

    if len(*selected_chat_template_name) > 0{
        chat_template_names:= []string{}
        chat_template_names,err=getChatTemplatename(chat_templates)
        if err != nil{
            fmt.Println("you have no chat template in the chat_template folder. ")
            os.Exit(1)
        }
        if len(chat_template_names) > 0 {
                chat_template, err:=getChatTemplateFromName(chat_templates, *selected_chat_template_name)
                if err != nil{
                    fmt.Println(err)
                    os.Exit(1)
                }
                DebugPrintln("names found and being used " + chat_template.Name, InfoLog)
                // this will combine the initialial content field with the user prompt using the template type specfied in the chat template file
                for i, msg:=range chat_template.ChatRequestBody.Messages{
                    if msg.Role=="user"{
                       starting_prompt=msg.Content
                       complete_prompt:=strings.ReplaceAll(starting_prompt, chat_template.PlaceholderType, *user_query)
                       chat_template.ChatRequestBody.Messages[i].Content=complete_prompt
                       DebugPrintln("The complete prompt " + complete_prompt, InfoLog)
                    }
                }
                if starting_prompt == ""{
                    fmt.Println("You are either missing the user field or user prompt is empty. Please fix your chat template prompt. ")
                    os.Exit(1)
                }
                api_response := ExecuteOpenRouterRequest(GetAPIKey("OPENROUTER_API_KEY"), chat_template.ChatRequestBody)

                if  api_response.Error != nil && api_response.Error.Code != 200{
                    DebugPrintln("api message error code: " + string(api_response.Error.Code ), InfoLog)
                    fmt.Println("Could not execute command. The API site returned the following message: \n")
                    fmt.Println(api_response.Error.Metadata.Raw)
                }else{
                    api_msg_content:=api_response.Choices[0].Message.Content
                    fmt.Println(api_msg_content)
                    writeAppendToLocalCommandHistory(filepath.Join(GetBaseDir(), "sbot_command_history.txt"), api_msg_content, 700)
                    if *execute_current_command{
                        executeCommand(api_msg_content)
                    }
                }
            if err!= nil{
                fmt.Println("options are empty from site")
            }
        }

    }else{
        DebugPrintln("Chat template option not supplied. Using default chat template body found in setting.json", InfoLog)
        commonSettings, err:=getCommonSettingsConfig("setting.json")
        if err != nil{
            fmt.Println("could not open setting file")
            os.Exit(0)
        }
        chat_template, err := getChatTemplateFromName(chat_templates, commonSettings.DefaultChatTemplate)
        if err != nil {
            fmt.Print(err)
            fmt.Print(" or valid in your settings.json file")
            os.Exit(1)
        }
        DebugPrintln("The select chat template name is " + chat_template.Name, InfoLog)
        for i, msg:=range chat_template.ChatRequestBody.Messages{
            if msg.Role=="user"{
                starting_prompt=msg.Content
                complete_prompt:=strings.ReplaceAll(starting_prompt, chat_template.PlaceholderType, *user_query)
                chat_template.ChatRequestBody.Messages[i].Content=complete_prompt
                DebugPrintln("The complete prompt " + complete_prompt, InfoLog)
            }
        }
        if starting_prompt == ""{
            fmt.Println("You are either missing the user field or user prompt is empty. Please fix your prompt. ")
            os.Exit(1)
        }
        api_response:=ExecuteOpenRouterRequest(GetAPIKey("OPENROUTER_API_KEY"), chat_template.ChatRequestBody)

        if  api_response.Error != nil && api_response.Error.Code!= 200{
            DebugPrintln("api message error code: " + string(api_response.Error.Code ), WarningLog)
            fmt.Println("Could not execute command. The API site returned the following message: \n")
            fmt.Println(api_response.Error.Metadata.Raw)
        }else{
            api_msg_content:=api_response.Choices[0].Message.Content
            fmt.Println(api_msg_content)
            writeAppendToLocalCommandHistory(filepath.Join(GetBaseDir(), "sbot_command_history.txt"), api_msg_content, 700)
            if *execute_current_command{
                executeCommand(api_msg_content)
            }
        }
    }

}

// functions in alphabetical order

func executeCommand(command string)(string, string, error){
    //!TODO sanitize any special chracters
    //re := regexp.MustCompile(`^[a-zA-Z0-9\s\.\|\'\"\-\/\_]+$`)
    // Check if the command matches the allowed pattern
    // read from settings file and unmarshal as shown
    // TODO! Check for the matching characters
   // if !re.MatchString(command) {
   //     fmt.Println("command contains dangerous special characters ")
   //     return "", "", nil
   // }
    common_settings,err := getCommonSettingsConfig("setting.json")
    if err != nil{
        fmt.Println("could not open settings file")
        os.Exit(0)
    }

    var commandStringOption []string
    shellToUse := strings.ToLower(common_settings.Shell)
    shellBaseName:= path.Base(shellToUse)
    if MatchesAny(shellBaseName, []string{"bash","ksh", "fish", "sh", "zsh"}) {
        commandStringOption=[]string{"-c"}
    }else if MatchesAny(shellBaseName, []string{"cmd"}) {
        commandStringOption=[]string{"/C"}
    }else if MatchesAny(shellBaseName, []string{"nushell"}) {
        commandStringOption=[]string{"-e"}
    }else if MatchesAny(shellBaseName, []string{"powershell", "pwsh"}){
        commandStringOption=[]string{"-Command"}
    }else{
        fmt.Println("The shell type provided in setting.json is unknown or not supported in sbot")
        os.Exit(1)
    }
    DebugPrintln("Using the following command string option " + commandStringOption[0], InfoLog)
    dangerous_commands:= common_settings.DangerousCommands
    if !common_settings.AllowDangerousCommands{
        for i:=0;i< len(dangerous_commands);i++ {
           if strings.Contains(command, dangerous_commands[i])    ||
              strings.Contains(shellToUse, dangerous_commands[i]) ||
              strings.Contains(commandStringOption[0], dangerous_commands[i]) {

               DebugPrintln("command: " + command + " dangerous command: " + dangerous_commands[i] + " i " + string(i), WarningLog)
               fmt.Println("your command is a 'Dangerous command type'. Please enable this in the setting file or try a different command")
               return "", "", nil
           }
        }
    }

    var stdout bytes.Buffer
    var stderr bytes.Buffer
    DebugPrintln("going to run the command with the follow components " + shellToUse + " " + commandStringOption[0] + " " + command, InfoLog)
    cmd := exec.Command(shellToUse, append(commandStringOption, command)...)
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    DebugPrintln("Command to execute >>> " + cmd.String(), InfoLog)
    err = cmd.Run()
    if err != nil{
		os.Stderr.WriteString(err.Error() + "\n")
    }
    if cmd.Stdout != nil{
        //fmt.Println()
        //fmt.Println(cmd.Stdout)
    }
    return stdout.String(), stderr.String(), err
}

func getAndLoadChatTemplates() (ChatTemplateList, error)  {
    DebugPrintln("Now attempting to load prompt files", InfoLog)
    chat_template_list := ChatTemplateList{}
    chat_templates_dir := "chat_templates"
    files, err := os.ReadDir(filepath.Join(GetBaseDir(), chat_templates_dir))
    if err != nil {
        abs_path, err := filepath.Abs(chat_templates_dir)
        if err != nil {
            fmt.Println("Could not read directory " + abs_path)
            return ChatTemplateList{}, err
        }
        fmt.Printf("There was an issue trying to read the directory %s\n", abs_path)
        return ChatTemplateList{}, err
    }

    for _, chat_templates_file := range files {
        if chat_templates_file.Type().IsRegular() {
            chat_template_abs_path := filepath.Join(GetBaseDir(), chat_templates_dir, chat_templates_file.Name())
            DebugPrintln("chat template file path being loaded: " + chat_template_abs_path, InfoLog)

            // Read the file content
            file_content, err := os.ReadFile(chat_template_abs_path)
            if err != nil {
                fmt.Printf("There was an issue attempting to open %s\n", chat_template_abs_path)
                continue // Skip to the next file if there's an error reading this one
            }

            // Initialize a new chat_template for each file
            var chat_templates_capture ChatTemplateList
            err = json.Unmarshal(file_content, &chat_templates_capture)
            if err != nil {
                fmt.Printf("Warning! There is a syntax issue with this JSON chat_template file: %s\n", chat_template_abs_path)
                fmt.Println("Checking other files...")
                continue // Skip to the next file if there's a JSON syntax issue
            }
            //append both arrays here
            chat_template_list.Templates = append(chat_templates_capture.Templates,chat_template_list.Templates...)
            //chat_template_list.Templates = append(chat_template_list.Templates, chat_templates_content.Templates...)
        }
    }

    return chat_template_list, nil
}

func getChatTemplateFromName(chat_templates ChatTemplateList, prompt_name string) (ChatTemplate, error){
    for _, prompt:= range chat_templates.Templates{
        if prompt.Name == prompt_name{
            return prompt, nil
        }
    }
    return ChatTemplate{}, fmt.Errorf("chat template name was not found")
}

func getChatTemplatename(chat_template_list ChatTemplateList) ([]string, error){
    chat_template_names:=[]string{}
    for _, chat_template := range chat_template_list.Templates{
            DebugPrintln("chat template name found " +  chat_template.Name, InfoLog)
            if len(chat_template.Name) > 0{
                chat_template_names = append(chat_template_names, chat_template.Name)
            }
    }
    return chat_template_names,nil
}

func getCommonSettingsConfig(fpath string) (CommonSettings, error){
    settings_file_content,err:=os.ReadFile(filepath.Join(GetBaseDir(), fpath))
    if err != nil{
        fmt.Println(err)
        return CommonSettings{}, err
    }

    var common_settings CommonSettings
    err=json.Unmarshal(settings_file_content, &common_settings )
    if err != nil{
        fmt.Println(err)
        return CommonSettings{}, err
    }
    return common_settings,nil
}
func lastCommandInHistory(file_name string) string {
    DebugPrintln("checking last command", InfoLog);
    last_command:=""
    bytesRead,err := os.ReadFile(file_name)
    if(err != nil){
        fmt.Println("could not read or open history file")
        fmt.Println(err)
    }
    fileContent := string(bytesRead)
    lines := strings.Split(fileContent, "\n")
    last_command = lines[len(lines) - 1]
    if len(last_command) <= 0 {
        fmt.Println("command history is empty.")
    }

    DebugPrintln("last command in history is " + last_command, InfoLog);
    return last_command
}

func listChatTemplateNames(){
    chat_templates,err := getAndLoadChatTemplates()
    if err != nil{
        fmt.Println(err)
        return
    }
    chat_templates_names, err := getChatTemplatename(chat_templates)
    if err != nil{
        fmt.Println(err)
        os.Exit(1)
    }

    for _, name := range chat_templates_names{
        fmt.Println(name)
    }
}

func showHistory(){
    content, err:= os.ReadFile(filepath.Join(GetBaseDir(), "sbot_command_history.txt" ) )
    if err != nil{
        fmt.Println(err)
    }
    fmt.Println(string(content))
}

func writeAppendToLocalCommandHistory(file_name string, content_passed string, perm int){
    content_from_file, err:=os.ReadFile(file_name)
    space:= "\n"
    if err != nil{
       fmt.Println("Could read not read from the history file. Please see if it exists.")
    }
    if(len(content_from_file) <= 0){
        space=""
    }
    content_to_add := string(content_from_file) + space + content_passed
    os.WriteFile(file_name, []byte(content_to_add), 700)
}

type CommonSettings struct{
    AllowDangerousCommands bool     `json:"allow_dangerous_commands"`
    Shell                  string   `json:"shell"`
    DangerousCommands      []string `json:"dangerous_commands"`
    DefaultChatTemplate    string   `json:"default_chat_template"`
}
