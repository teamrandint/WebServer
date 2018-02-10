package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var channel = make(chan string, 1000)
var users = 0
var hostName = ""

func makeHTTPRequest(command string) {
	tokens := strings.Split(command, ",")
	cmdType := strings.Replace(tokens[0], " ", "", -1)
	userID := strings.Replace(tokens[1], " ", "", -1)
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

	endpointURL := hostName + cmdType + "/"
	resp, err := http.PostForm(endpointURL, url.Values{"username": {userID}, "stock": {stock}, "amount": {amount}})

	if err != nil {
		fmt.Println("REQUEST ERROR OCCURED!!")
	} else {
		// fmt.Println(resp.StatusCode)
		if resp.StatusCode == 400 {
			// fmt.Println(endpointURL)
		}
		// Always close the response-body, even if content not required
		defer resp.Body.Close()
	}
}

// Special dumplog request method for when end of requests is reached.
func dumplog(filename string) {
	endpointURL := hostName + "DUMPLOG/"
	resp, err := http.PostForm(endpointURL, url.Values{"filename": {filename}})
	if err != nil {
		fmt.Println("REQUEST ERROR OCCURED!!")
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 404 {
			fmt.Println(endpointURL)
		}
	}
	// Close connection
	defer resp.Body.Close()
}

func makeUserRequests(commands []string) {
	for _, command := range commands {
		makeHTTPRequest(command)
	}
	channel <- "done"
}

// Processes the specified input file and makes async requests for each user.
// outputs the filename of the dumpfile to be written at the end as the final command.
func processFile(address string, port string, filename string) string {
	workloadFile, err := os.Open(filename)
	var outfileName string
	if err != nil {
		log.Fatal(err)
	}
	defer workloadFile.Close()

	scanner := bufio.NewScanner(workloadFile)

	var userCommands = make(map[string][]string)
	userID := ""

	for scanner.Scan() {
		line := scanner.Text()
		params := strings.Split(line, ",")
		re := regexp.MustCompile("^\\[(\\d)+\\]\\s")

		// Id contains a space in the file, remove it
		userID = strings.Replace(params[1], " ", "", -1)

		// user id is the filename for dumplog commands
		if userID[0] == '.' {
			outfileName = strings.Replace(params[1], " ", "", -1)
		} else {
			fullCommand := re.ReplaceAllString(line, "")
			userCommands[userID] = append(userCommands[userID], fullCommand)
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
	fmt.Print(outfileName)
	return outfileName
}

func listenForCompleted() {
	for i := 0; i < users; i++ {
		status := <-channel
		fmt.Println(status)
	}
	fmt.Println("finished users requests")
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
	outFileName := processFile(address, port, filename)
	listenForCompleted()
	dumplog(outFileName)
}
