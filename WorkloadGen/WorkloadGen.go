package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	endPoint := os.Args[1]
	workloadFile := os.Args[2]
	fmt.Printf("Testing %v on endpoint %v", workloadFile, endPoint)

	users := splitUsersFromFile(workloadFile)

	for _, user := range users {
		go runUserRequests(endPoint, user)
	}
	defer postDumpLog()
}

func splitUsersFromFile(filename string) map[string][]string {
	commands := []string{
		"ADD,userid,amount",
		"QUOTE",
		"BUY",
		"COMMIT_BUY",
		"CANCEL_BUY",
		"SELL",
		"COMMIT_SELL",
		"CANCEL_SELL",
		"SET_BUY_AMOUNT",
		"CANCEL_SET_BUY",
		"SET_BUY_TRIGGER",
		"SET_SELL_AMOUNT",
		"SET_SELL_TRIGGER",
		"CANCEL_SET_SELL",
		"DUMPLOG",
		"DUMPLOG",
		"DISPLAY_SUMMARY",
	}

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	//for _, line := range workLoad.
	x := make(map[string][]string)

	x["key"] = append(x["key"], "value")
	x["key"] = append(x["key"], "value1")

	fmt.Println(commands)
	return x
}

func runUserRequests(endpoint string, commands []string) {
	for _, command := range commands {
		fmt.Println(command)
	}
}

func postDumpLog() {

}
