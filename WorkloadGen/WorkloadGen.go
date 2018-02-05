package main

import (
	"fmt"
	"bufio"
	"log"
	"os"
	"strings"
	"regexp"
	"net/http"
	"net/url"
)

var channel = make(chan string, 10)
var users = 0
var hostName = ""

func makeHttpRequest(command string) {
	tokens := strings.Split(command, ",")
	cmdType := strings.Replace(tokens[0], " ", "", -1)
	userId := strings.Replace(tokens[1], " ", "", -1)
	stock := ""
	amount := ""

	if len(tokens) == 3 {
		re := regexp.MustCompile("(\\d)+\\.\\d\\d")
		if re.MatchString(tokens[2]) {
			// third param is amount
			amount = strings.Replace(tokens[2], " ", "", -1)
		} else {
			stock = strings.Replace(tokens[2], " ", "", -1)
		}
	} else if len(tokens) == 4 {
		stock = strings.Replace(tokens[2], " ", "", -1)
		amount = strings.Replace(tokens[3], " ", "", -1)
	}

	endpointUrl := hostName + cmdType + "/"
	resp, err := http.PostForm(endpointUrl, url.Values{"username":{userId}, "stock":{stock}, "amount":{amount}})

	if err != nil {
		fmt.Println("REQUEST ERROR OCCURED!!")
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 404 {
			fmt.Println(endpointUrl)
		}
	}
}

func makeUserRequests(commands []string) {
	for _, command := range commands {
		makeHttpRequest(command)
	}
	channel <- "done"
}

func processFile(address string, port string, filename string) {
	workloadFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	} 
	defer workloadFile.Close()

	scanner := bufio.NewScanner(workloadFile)

	var userCommands = make(map[string][]string)
	userId := ""

	for scanner.Scan() {
		line := scanner.Text()
		params := strings.Split(line, ",")
		re := regexp.MustCompile("^\\[(\\d)+\\]\\s")

		// Id contains a space in the file, remove it
		userId = strings.Replace(params[1], " ", "", -1)
		// user id is blank for dumplog
		if userId[0] == '.' {
			fmt.Println("dumplog command found")
			// Dumplog is the last command sent, so send it specially
		} else {
			fullCommand := re.ReplaceAllString(line, "")
			userCommands[userId] = append(userCommands[userId], fullCommand)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(userCommands))
	for _, v := range userCommands {
		users++
		makeUserRequests(v)
	}
}

func listenForCompleted() {
	for i := 0; i < users; i++ {
		status := <- channel
		fmt.Println(status)
	}
	fmt.Println("finished program")
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Please supply a host name, port number, and filename.")
		return
	}

	address := os.Args[1]
	port := os.Args[2]
	filename := os.Args[3]
	hostName = "http://" + address + ":" + port + "/"
	processFile(address, port, filename)
	listenForCompleted()
}