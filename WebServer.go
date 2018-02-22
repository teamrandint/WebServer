package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"./Commands"
	"sync/atomic"

	"./UserSessions"
	"./logger"
	"./transmitter"
)

type WebServer struct {
	Name              string
	transactionNumber int64
	userSessions      map[string]*usersessions.UserSession
	transmitter       *transmitter.Transmitter
	logger            logger.Logger
	validPath         *regexp.Regexp
}

func (webServer *WebServer) makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		m := webServer.validPath.FindStringSubmatch(request.URL.Path)
		if m == nil {
			http.NotFound(writer, request)
			panic(request)
			return
		}
		fn(writer, request, m[1])
	}
}

// Garuntees that the user exists in the session cache for managing operations
func (webServer *WebServer) userLogin(id string) {
	if webServer.userSessions[id] == nil {
		webServer.createUserSession(id)
	}
}

// Adds the specified user to the sessions list.
func (webServer *WebServer) createUserSession(id string) {
	webServer.userSessions[id] = usersessions.NewUserSession(id)
}

func (webServer *WebServer) addHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "ADD", username, nil, nil, amount)
	// TODO : generic login once for each user.
	webServer.userLogin(username)

	resp := webServer.transmitter.MakeRequest(currTransNum ,"ADD," + username + "," + amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "ADD",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
	}
}

func (webServer *WebServer) quoteHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "QUOTE",
		username, stock, nil, nil)

	webServer.userLogin(username)

	resp := webServer.transmitter.MakeRequest(currTransNum, "QUOTE," + username + "," + stock)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "QUOTE",
			username, stock, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
	}
}

func (webServer *WebServer) buyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")
	command := commands.NewCommand("BUY", username, []string{stock, amount})

	webServer.logger.UserCommand(webServer.Name, currTransNum, "BUY",
		username, stock, nil, amount)

	webServer.userLogin(username)

	resp := webServer.transmitter.MakeRequest(currTransNum, "BUY," + username + "," + stock + "," + amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "BUY",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}

	// Append buy to pendingBuys list
	webServer.userSessions[username].PendingBuys = append(webServer.userSessions[username].PendingBuys, command)
}

func (webServer *WebServer) commitBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	webServer.userLogin(username)

	webServer.logger.UserCommand(webServer.Name, currTransNum, "COMMIT_BUY",
		username, nil, nil, nil)

	if !webServer.userSessions[username].HasPendingBuys() {
		// No pendings buys, return error
		//fmt.Printf("No buys to commit for user %s\n", username)
		webServer.logger.SystemError(webServer.Name, currTransNum, "COMMIT_BUY",
			username, nil, nil, nil, "No pending buys to commit")
		http.Error(writer, "Invalid request", 400)
		return
	}

	command := webServer.userSessions[username].PendingBuys[0]
	var resp string
	if command.HasTimeElapsed() {
		// Time has elapsed on Buy, automatically cancel request
		resp = webServer.transmitter.MakeRequest(currTransNum, "CANCEL_BUY," + username)
		webServer.logger.SystemError(webServer.Name, currTransNum, "COMMIT_BUY",
			username, nil, nil, nil, "Time elapsed on most recent buy request")
		// TODO: Ensure that this is not finishing early, and actually popping off results
		http.Error(writer, "Invalid request", 400)
		//fmt.Printf("Time has elapsed on last buy for user %s\n", username)
	} else {
		resp = webServer.transmitter.MakeRequest(currTransNum, "COMMIT_BUY," + username)
	}

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "COMMIT_BUY",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	// Pop last sell off the pending list.
	webServer.userSessions[username].PendingBuys = webServer.userSessions[username].PendingBuys[1:]
}

func (webServer *WebServer) cancelBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "CANCEL_BUY",
		username, nil, nil, nil)

	webServer.userLogin(username)
	if !webServer.userSessions[username].HasPendingBuys() {
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_BUY",
			username, nil, nil, nil, "No pending buys to cancel")
		http.Error(writer, "Invalid request", 400)
		//fmt.Printf("No buys to cancel for user %s\n", username)
		return
	}

	resp := webServer.transmitter.MakeRequest(currTransNum, "CANCEL_BUY," + username)

	if resp == "-1" {	
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_BUY",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	// Pop last sell off the pending list.
	webServer.userSessions[username].PendingBuys = webServer.userSessions[username].PendingBuys[1:]
}

func (webServer *WebServer) sellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")
	command := commands.NewCommand("SELL", username, []string{stock, amount})

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SELL",
		username, stock, nil, amount)

	webServer.userLogin(username)
	resp := webServer.transmitter.MakeRequest(currTransNum, "SELL," + username + "," + stock + "," + amount)
	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SELL",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	webServer.userSessions[username].PendingSells = append(webServer.userSessions[username].PendingSells, command)
}

func (webServer *WebServer) commitSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "COMMIT_SELL",
		username, nil, nil, nil)
	webServer.userLogin(username)

	if !webServer.userSessions[username].HasPendingSells() {
		// No pendings buys, return error
		webServer.logger.SystemError(webServer.Name, currTransNum, "COMMIT_SELL",
			username, nil, nil, nil, "No pending sells to commit")
		http.NotFound(writer, request)
		//fmt.Printf("No sells to commit for user %s\n", username)
		return
	}

	command := webServer.userSessions[username].PendingSells[0]
	var resp string

	if command.HasTimeElapsed() {
		// Time has elapsed on Buy, automatically cancel request
		resp = webServer.transmitter.MakeRequest(currTransNum, "CANCEL_SELL," + username)
		go webServer.logger.SystemError(webServer.Name, currTransNum, "COMMIT_SELL",
			username, nil, nil, nil, "Time elapsed on most recent sell")
		http.NotFound(writer, request)
		//fmt.Printf("Time has elapsed on last sell for user %s\n", username)
	} else {
		resp = webServer.transmitter.MakeRequest(currTransNum, "COMMIT_SELL," + username)
	}

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "COMMIT_SELL",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	// Pop last sell off the pending list.
	webServer.userSessions[username].PendingSells = webServer.userSessions[username].PendingSells[1:]
}

func (webServer *WebServer) cancelSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	webServer.logger.UserCommand(webServer.Name, currTransNum, "CANCEL_SELL",
		username, nil, nil, nil)
	webServer.userLogin(username)

	if !webServer.userSessions[username].HasPendingSells() {
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_SELL",
			username, nil, nil, nil, "User has no pending sells")
		http.NotFound(writer, request)
		//fmt.Printf("No sells to cancel for user %s\n", username)
		return
	}

	resp := webServer.transmitter.MakeRequest(currTransNum, "CANCEL_SELL," + username)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_SELL",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	// Pop last sell off the pending list.
	webServer.userSessions[username].PendingSells = webServer.userSessions[username].PendingSells[1:]
}

func (webServer *WebServer) setBuyAmountHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SET_BUY_AMOUNT",
		username, stock, nil, amount)

	resp := webServer.transmitter.MakeRequest(currTransNum, "SET_BUY_AMOUNT," + username + "," + stock + "," + amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SET_BUY_AMOUNT",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
}

func (webServer *WebServer) cancelSetBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "CANCEL_SET_BUY",
		username, stock, nil, nil)

	resp := webServer.transmitter.MakeRequest(currTransNum, "CANCEL_SET_BUY," + username + "," + stock)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_SET_BUY",
			username, stock, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
}

func (webServer *WebServer) setBuyTriggerHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SET_BUY_TRIGGER",
		username, stock, nil, amount)

	resp := webServer.transmitter.MakeRequest(currTransNum, "SET_BUY_TRIGGER," + username + "," + stock + "," + amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SET_BUY_TRIGGER",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
}

func (webServer *WebServer) setSellAmountHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SET_SELL_AMOUNT", 
		username, stock, nil, amount)

	resp := webServer.transmitter.MakeRequest(currTransNum, "SET_SELL_AMOUNT," + username + "," + stock + "," + amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SET_SELL_AMOUNT",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
}

func (webServer *WebServer) setSellTriggerHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SET_SELL_TRIGGER",
		username, stock, nil, amount)

	resp := webServer.transmitter.MakeRequest(currTransNum, "SET_SELL_TRIGGER," + username + "," + stock + "," + amount)
	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SET_SELL_TRIGGER",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
}

func (webServer *WebServer) cancelSetSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "CANCEL_SET_SELL",
		username, stock, nil, nil)

	resp := webServer.transmitter.MakeRequest(currTransNum,"CANCEL_SET_SELL," + username + "," + stock)
	if resp == "-1" {	
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_SET_SELL",
			username, stock, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
}

func (webServer *WebServer) dumplogHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	filename := request.FormValue("filename")
	message := ""

	if len(username) == 0 {
		message = "DUMPLOG," + filename
		webServer.logger.UserCommand(webServer.Name, currTransNum, "DUMPLOG",
			nil, nil, filename, nil)
	} else {
		message = "DUMPLOG," + username + "," + filename
		webServer.logger.UserCommand(webServer.Name, currTransNum, "DUMPLOG",
			username, nil, filename, nil)
	}

	resp := webServer.transmitter.MakeRequest(currTransNum, message)
	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "DUMPLOG",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
}

func (webServer *WebServer) displaySummaryHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "DISPLAY_SUMMARY",
		username, nil, nil, nil)

	resp := webServer.transmitter.MakeRequest(currTransNum,"DISPLAY_SUMMARY," + username)
	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "DISPLAY_SUMMARY", 
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
}

func (webServer *WebServer) genericHandler(writer http.ResponseWriter, request *http.Request, title string) {
	fmt.Fprintf(writer, "Hello from end point %s!", request.URL.Path[1:])
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Please enter a valid server address and port number.")
		return
	}

	address := os.Args[1]
	port := os.Args[2]

	serverAddress := string(address) + ":" + string(port)
	auditAddr := "http://localhost:8080"

	webServer := &WebServer{
		Name:              "webserver",
		transactionNumber: 0,
		userSessions:      make(map[string]*usersessions.UserSession),
		transmitter:       transmitter.NewTransmitter("localhost", "8000"),
		logger:            logger.AuditLogger{Addr: auditAddr},
		validPath:         regexp.MustCompile("^/(ADD|QUOTE|BUY|COMMIT_BUY|CANCEL_BUY|SELL|COMMIT_SELL|CANCEL_SELL|SET_BUY_AMOUNT|CANCEL_SET_BUY|SET_BUY_TRIGGER|SET_SELL_AMOUNT|SET_SELL_TRIGGER|CANCEL_SET_SELL|DUMPLOG|DISPLAY_SUMMARY)/$"),
	}

	http.HandleFunc("/", webServer.makeHandler(webServer.genericHandler))
	http.HandleFunc("/ADD/", webServer.makeHandler(webServer.addHandler))
	http.HandleFunc("/QUOTE/", webServer.makeHandler(webServer.quoteHandler))
	http.HandleFunc("/BUY/", webServer.makeHandler(webServer.buyHandler))
	http.HandleFunc("/COMMIT_BUY/", webServer.makeHandler(webServer.commitBuyHandler))
	http.HandleFunc("/CANCEL_BUY/", webServer.makeHandler(webServer.cancelBuyHandler))
	http.HandleFunc("/SELL/", webServer.makeHandler(webServer.sellHandler))
	http.HandleFunc("/COMMIT_SELL/", webServer.makeHandler(webServer.commitSellHandler))
	http.HandleFunc("/CANCEL_SELL/", webServer.makeHandler(webServer.cancelSellHandler))
	http.HandleFunc("/SET_BUY_AMOUNT/", webServer.makeHandler(webServer.setBuyAmountHandler))
	http.HandleFunc("/CANCEL_SET_BUY/", webServer.makeHandler(webServer.cancelSetBuyHandler))
	http.HandleFunc("/SET_BUY_TRIGGER/", webServer.makeHandler(webServer.setBuyTriggerHandler))
	http.HandleFunc("/SET_SELL_AMOUNT/", webServer.makeHandler(webServer.setSellAmountHandler))
	http.HandleFunc("/SET_SELL_TRIGGER/", webServer.makeHandler(webServer.setSellTriggerHandler))
	http.HandleFunc("/CANCEL_SET_SELL/", webServer.makeHandler(webServer.cancelSetSellHandler))
	http.HandleFunc("/DUMPLOG/", webServer.makeHandler(webServer.dumplogHandler))
	http.HandleFunc("/DISPLAY_SUMMARY/", webServer.makeHandler(webServer.displaySummaryHandler))

	fmt.Printf("Successfully started server on address: %s, port #: %s\n", address, port)
	http.ListenAndServe(serverAddress, nil)
}
