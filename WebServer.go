package main

import (
	"net/http"
	"fmt"
	"os"
	"regexp"
	"./Commands"
	"./transmitter"
	"./UserSessions"
)

type WebServer struct {
	transactionNumber int
	userSessions map[string]*usersessions.UserSession
	transmitter *transmitter.Transmitter
}

var validPath = regexp.MustCompile("^/(ADD|QUOTE|BUY|COMMIT_BUY|CANCEL_BUY|SELL|COMMIT_SELL|CANCEL_SELL|SET_BUY_AMOUNT|CANCEL_SET_BUY|SET_BUY_TRIGGER|SET_SELL_AMOUNT|SET_SELL_TRIGGER|CANCEL_SET_SELL|DUMPLOG|DISPLAY_SUMMARY)/$")
var webServer = &WebServer{}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		m := validPath.FindStringSubmatch(request.URL.Path)
		if m == nil {
			http.NotFound(writer, request)
			return
		}
		fn(writer, request, m[1])
	}
}

// Garuntees that the user exists in the session cache for managing operations
func userLogin(id string) {
	if webServer.userSessions[id] == nil {
		createUserSession(id)
	}
}

// Adds the specified user to the sessions list.
func createUserSession(id string) {
	webServer.userSessions[id] = usersessions.NewUserSession(id)
}

func addHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	amount := request.FormValue("amount")
	userLogin(username)

	go webServer.transmitter.MakeRequest("ADD," + username + "," + amount)
}

func quoteHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	userLogin(username)

	go webServer.transmitter.MakeRequest("QUOTE," + username + "," + stock)
}

func buyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")
	command := commands.NewCommand("BUY", username, []string{stock, amount})
	userLogin(username)

	go webServer.transmitter.MakeRequest("BUY," + username + "," + stock + "," + amount)

	// Append buy to pendingBuys list
	webServer.userSessions[username].PendingBuys = append(webServer.userSessions[username].PendingBuys, command)
}

func commitBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	userLogin(username)

	if !webServer.userSessions[username].HasPendingBuys() {
		// No pendings buys, return error
		http.NotFound(writer, request)
		fmt.Printf("No buys to commit for user %s\n", username)
		return
	}

	command := webServer.userSessions[username].PendingBuys[0]

	if command.HasTimeElapsed() {
		// Time has elapsed on Buy, automatically cancel request
		go webServer.transmitter.MakeRequest("CANCEL_BUY," + username)
		http.NotFound(writer, request)
		fmt.Printf("Time has elapsed on last buy for user %s\n", username)
	} else {
		go webServer.transmitter.MakeRequest("COMMIT_BUY," + username)
	}

	// TODO: Check if the command was successful on the trans server

	// Pop last sell off the pending list.
	webServer.userSessions[username].PendingBuys  = webServer.userSessions[username].PendingBuys[1:]
}

func cancelBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	userLogin(username)
	if !webServer.userSessions[username].HasPendingBuys() {
		http.NotFound(writer, request)
		fmt.Printf("No buys to cancel for user %s\n", username)
		return
	}

	go webServer.transmitter.MakeRequest("CANCEL_BUY," + username)

	// Pop last sell off the pending list.
	webServer.userSessions[username].PendingBuys = webServer.userSessions[username].PendingBuys[1:]
}

func sellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")
	command := commands.NewCommand("SELL", username, []string{stock, amount})

	userLogin(username)
	go webServer.transmitter.MakeRequest("SELL," + username + "," + stock + "," + amount)

	webServer.userSessions[username].PendingSells = append(webServer.userSessions[username].PendingSells, command)
}

func commitSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	userLogin(username)

	if !webServer.userSessions[username].HasPendingSells() {
		// No pendings buys, return error
		http.NotFound(writer, request)
		fmt.Printf("No sells to commit for user %s\n", username)
		return
	}

	command := webServer.userSessions[username].PendingSells[0]

	if command.HasTimeElapsed() {
		// Time has elapsed on Buy, automatically cancel request
		go webServer.transmitter.MakeRequest("CANCEL_SELL," + username)
		http.NotFound(writer, request)
		fmt.Printf("Time has elapsed on last sell for user %s\n", username)
	} else {
		go webServer.transmitter.MakeRequest("COMMIT_SELL," + username)
	}

	// TODO: Check if the command was successful on the trans server

	// Pop last sell off the pending list.
	webServer.userSessions[username].PendingSells  = webServer.userSessions[username].PendingSells[1:]
}

func cancelSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	userLogin(username)

	if !webServer.userSessions[username].HasPendingSells() {
		http.NotFound(writer, request)
		fmt.Printf("No sells to cancel for user %s\n", username)
		return
	}

	go webServer.transmitter.MakeRequest("CANCEL_SELL," + username)

	// Pop last sell off the pending list.
	webServer.userSessions[username].PendingSells = webServer.userSessions[username].PendingSells[1:]
}


func setBuyAmountHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	go webServer.transmitter.MakeRequest("SET_BUY_AMOUNT," + username + "," + stock + "," + amount)
}

func cancelSetBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")

	go webServer.transmitter.MakeRequest("CANCEL_SET_BUY," + username + "," + stock)
}

func setBuyTriggerHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	go webServer.transmitter.MakeRequest("SET_BUY_TRIGGER," + username + "," + stock + "," + amount)
}

func setSellAmountHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	go webServer.transmitter.MakeRequest("SET_SELL_AMOUNT," + username + "," + stock + "," + amount)

}

func setSellTriggerHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	go webServer.transmitter.MakeRequest("SET_SELL_TRIGGER," + username + "," + stock + "," + amount)
}

func cancelSetSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	stock := request.FormValue("stock")

	go webServer.transmitter.MakeRequest("CANCEL_SET_SELL," + username + "," + stock)
}

func dumplogHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")
	filename := request.FormValue("filename")
	message := ""

	if len(username) == 0 {
		message = "DUMPLOG," + filename
	} else {
		message = "DUMPLOG," + username + "," + filename
	}

	go webServer.transmitter.MakeRequest(message)
}

func displaySummaryHandler(writer http.ResponseWriter, request *http.Request, title string) {
	webServer.transactionNumber++
	username := request.FormValue("username")

	go webServer.transmitter.MakeRequest("DISPLAY_SUMMARY," + username)
}

func genericHandler(writer http.ResponseWriter, request *http.Request, title string) {
	fmt.Fprintf(writer, "Hello from end point %s!", request.URL.Path[1:])
}

func main(){
	if (len(os.Args) < 3) {
		fmt.Println("Please enter a valid server address and port number.")
		return
	}

	address := os.Args[1]
	port := os.Args[2]

	serverAddress := string(address) + ":" + string(port)
	http.HandleFunc("/", makeHandler(genericHandler))
	http.HandleFunc("/ADD/", makeHandler(addHandler))
	http.HandleFunc("/QUOTE/", makeHandler(quoteHandler))
	http.HandleFunc("/BUY/", makeHandler(buyHandler))
	http.HandleFunc("/COMMIT_BUY/", makeHandler(commitBuyHandler))
	http.HandleFunc("/CANCEL_BUY/", makeHandler(cancelBuyHandler))
	http.HandleFunc("/SELL/", makeHandler(sellHandler))
	http.HandleFunc("/COMMIT_SELL/", makeHandler(commitSellHandler))
	http.HandleFunc("/CANCEL_SELL/", makeHandler(cancelSellHandler))
	http.HandleFunc("/SET_BUY_AMOUNT/", makeHandler(setBuyAmountHandler))
	http.HandleFunc("/CANCEL_SET_BUY/", makeHandler(cancelSetBuyHandler))
	http.HandleFunc("/SET_BUY_TRIGGER/", makeHandler(setBuyTriggerHandler))
	http.HandleFunc("/SET_SELL_AMOUNT/", makeHandler(setSellAmountHandler))
	http.HandleFunc("/SET_SELL_TRIGGER/", makeHandler(setSellTriggerHandler))
	http.HandleFunc("/CANCEL_SET_SELL/", makeHandler(cancelSetSellHandler))
	http.HandleFunc("/DUMPLOG/", makeHandler(dumplogHandler))
	http.HandleFunc("/DISPLAY_SUMMARY/", makeHandler(displaySummaryHandler))

	// Connection to the transaction server. 
	// TODO make system args for setting transaction server
	webServer.transmitter = transmitter.NewTransmitter("localhost", "8000")

	// Connection to the Audit server
	// TODO: add connection to Audit server
	
	webServer.userSessions = make(map[string]*usersessions.UserSession)

	fmt.Printf("Successfully started server on address: %s, port #: %s\n", address, port)
	http.ListenAndServe(serverAddress, nil)
}