package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"
)

// go run WorkloadGen.go serverAddr:port workloadfile
func main() {
	serverAddr := os.Args[1]
	workloadFile := os.Args[2]
	fmt.Printf("Testing %v on serverAddr %v\n", workloadFile, serverAddr)

	users := splitUsersFromFile(workloadFile)
	fmt.Printf("Found %d users...\n", len(users))

	var wg sync.WaitGroup
	for userName, commands := range users {
		fmt.Printf("Running user %v's commands...\n", userName)

		wg.Add(1)
		go func(commands []string) {
			for _, command := range commands {
				// username, stock, amount, filename
				endpoint, values := parseCommand(command)
				time.Sleep(time.Second) // ADJUST THIS TO CHANGE DELAY
				http.PostForm(serverAddr+endpoint, values)
			}

			wg.Done()
		}(commands)
	}

	wg.Wait()
	http.PostForm(serverAddr+"/DUMPLOG/", url.Values{"filename": {"./output.xml"}})
	fmt.Printf("Done!\n")
}

func splitUsersFromFile(filename string) map[string][]string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	//https://regex101.com/r/O6xaTp/3
	re := regexp.MustCompile(`\[\d+\] ((?P<endpoint>\w+),(?P<user>\w+)(,\w+\.*\d*)*)`)
	outputCommands := make(map[string][]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)

		if matches != nil {
			command := matches[1]
			//endpoint := matches[2]
			user := matches[3]
			outputCommands[user] = append(outputCommands[user], command)
		}
	}

	return outputCommands
}

// Parse a single line command into the corresponding endpoint and values
func parseCommand(cmd string) (endpoint string, v url.Values) {
	return "", url.Values{"test": {"test"}}
}
