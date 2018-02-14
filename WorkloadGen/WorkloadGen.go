package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func main() {
	endPoint := os.Args[1]
	workloadFile := os.Args[2]
	fmt.Printf("Testing %v on endpoint %v\n", workloadFile, endPoint)

	users := splitUsersFromFile(workloadFile)

	for userName, user := range users {
		fmt.Printf("Running user %v's commands...\n", userName)
		go runUserRequests(endPoint, user)
	}
	defer postDumpLog()
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

func runUserRequests(endpoint string, commands []string) {
	for _, command := range commands {
		fmt.Println(command)
	}
}

func postDumpLog() {
	runUserRequests("DUMPLOG", []string{"workloadgen_results.xml"})
}
