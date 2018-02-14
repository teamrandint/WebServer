package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sync"
)

func main() {
	endPoint := os.Args[1]
	workloadFile := os.Args[2]
	fmt.Printf("Testing %v on endpoint %v\n", workloadFile, endPoint)

	users := splitUsersFromFile(workloadFile)
	fmt.Printf("Found %d users...\n", len(users))

	var wg sync.WaitGroup
	for userName, commands := range users {
		fmt.Printf("Running user %v's commands...\n", userName)
		wg.Add(1)
		go func(commands []string) {
			for _, command := range commands {
				//command = command + "a"
				fmt.Println(command)
			}

			wg.Done()
		}(commands)
	}

	wg.Wait()
	fmt.Printf("Done!\n")
	// TODO run dumplog
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
