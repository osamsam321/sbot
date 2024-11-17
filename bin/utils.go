package main

import (
	"fmt"
	"os"
    "log"
    "bufio"
	"github.com/joho/godotenv"
	"path/filepath"
    "strings"
)

type LogType int32
var infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var warningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
var errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

const (
    InfoLog LogType=iota
    WarningLog
    ErrorLog
)
func containsAny(target string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(target, substring) {
			return true
		}
	}
	return false
}
func MatchesAny(target string, substrings []string) bool {
	for _, substring := range substrings {
		if target==substring {
            DebugPrintln("target " + target + " matches " + substring, InfoLog)
			return true
		}
	}
	return false
}

func DebugPrintln(msg string, logType LogType){
    if *debug_enabled{
        switch logType{
        case ErrorLog:
            errorLog.Println(msg)
        case WarningLog:
            warningLog.Println(msg)
        default:
            infoLog.Println(msg)
        }
    }
}

func DebugPrintf(msg string){
    if *debug_enabled{
        fmt.Printf(msg)
    }
}

func GetAPIKey(envVariable string) (string){
    err := godotenv.Load(filepath.Join(GetBaseDir(),".env"))
    DebugPrintln("grabbing api key from " + filepath.Join(GetBaseDir(), ".env"), InfoLog)
    if err != nil{
        fmt.Println("Error loading .env file")
    }
    return os.Getenv(envVariable)
}

func GetBaseDir() string {
    file_executable_path, err := os.Executable()
    bin_dir := filepath.Dir(file_executable_path)
	if err != nil {
        fmt.Println(err)
	}
    base_dir := filepath.Dir(bin_dir)
	return base_dir
}

func GetStdinAsString() (string, error) {
    var input string
    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        input += scanner.Text() + "\n"
    }

    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading from stdin:", err)
        return "", err
    }

    return input, nil
}

func StdinExist() bool{
    fi, err := os.Stdin.Stat()
    if err != nil {
        fmt.Println(err)
    }
    if fi.Mode()&os.ModeCharDevice == 0 {
		DebugPrintln("There is stdin input available.", InfoLog)
        return true
	}
    return false
}

