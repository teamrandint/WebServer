package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sync/atomic"

	"seng468/WebServer/UserSessions"
	"seng468/WebServer/logger"
	"seng468/WebServer/transmitter"
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
	resp := webServer.transmitter.MakeRequest(currTransNum, "ADD,"+username+","+amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "ADD",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) quoteHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "QUOTE", username, stock, nil, nil)
	resp := webServer.transmitter.MakeRequest(currTransNum, "QUOTE,"+username+","+stock)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "QUOTE",
			username, stock, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) buyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "BUY", username, stock, nil, amount)
	resp := webServer.transmitter.MakeRequest(currTransNum, "BUY,"+username+","+stock+","+amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "BUY",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) commitBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "COMMIT_BUY", username, nil, nil, nil)
	resp := webServer.transmitter.MakeRequest(currTransNum, "COMMIT_BUY,"+username)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "COMMIT_BUY",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) cancelBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "CANCEL_BUY", username, nil, nil, nil)
	resp := webServer.transmitter.MakeRequest(currTransNum, "CANCEL_BUY,"+username)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_BUY",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) sellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SELL", username, stock, nil, amount)
	resp := webServer.transmitter.MakeRequest(currTransNum, "SELL,"+username+","+stock+","+amount)
	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SELL",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) commitSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "COMMIT_SELL", username, nil, nil, nil)
	resp := webServer.transmitter.MakeRequest(currTransNum, "COMMIT_SELL,"+username)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "COMMIT_SELL",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) cancelSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	webServer.logger.UserCommand(webServer.Name, currTransNum, "CANCEL_SELL", username, nil, nil, nil)
	resp := webServer.transmitter.MakeRequest(currTransNum, "CANCEL_SELL,"+username)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_SELL",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) setBuyAmountHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SET_BUY_AMOUNT", username, stock, nil, amount)

	resp := webServer.transmitter.MakeRequest(currTransNum, "SET_BUY_AMOUNT,"+username+","+stock+","+amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SET_BUY_AMOUNT",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) cancelSetBuyHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "CANCEL_SET_BUY", username, stock, nil, nil)

	resp := webServer.transmitter.MakeRequest(currTransNum, "CANCEL_SET_BUY,"+username+","+stock)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_SET_BUY",
			username, stock, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) setBuyTriggerHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SET_BUY_TRIGGER",
		username, stock, nil, amount)

	resp := webServer.transmitter.MakeRequest(currTransNum, "SET_BUY_TRIGGER,"+username+","+stock+","+amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SET_BUY_TRIGGER",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) setSellAmountHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SET_SELL_AMOUNT",
		username, stock, nil, amount)

	resp := webServer.transmitter.MakeRequest(currTransNum, "SET_SELL_AMOUNT,"+username+","+stock+","+amount)

	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SET_SELL_AMOUNT",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) setSellTriggerHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")
	amount := request.FormValue("amount")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "SET_SELL_TRIGGER",
		username, stock, nil, amount)

	resp := webServer.transmitter.MakeRequest(currTransNum, "SET_SELL_TRIGGER,"+username+","+stock+","+amount)
	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "SET_SELL_TRIGGER",
			username, stock, nil, amount, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) cancelSetSellHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))
	username := request.FormValue("username")
	stock := request.FormValue("stock")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "CANCEL_SET_SELL",
		username, stock, nil, nil)

	resp := webServer.transmitter.MakeRequest(currTransNum, "CANCEL_SET_SELL,"+username+","+stock)
	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "CANCEL_SET_SELL",
			username, stock, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}

	writer.WriteHeader(http.StatusOK)
	return
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

	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) displaySummaryHandler(writer http.ResponseWriter, request *http.Request, title string) {
	currTransNum := int(atomic.AddInt64(&webServer.transactionNumber, 1))

	username := request.FormValue("username")

	webServer.logger.UserCommand(webServer.Name, currTransNum, "DISPLAY_SUMMARY",
		username, nil, nil, nil)

	resp := webServer.transmitter.MakeRequest(currTransNum, "DISPLAY_SUMMARY,"+username)
	if resp == "-1" {
		webServer.logger.SystemError(webServer.Name, currTransNum, "DISPLAY_SUMMARY",
			username, nil, nil, nil, "Bad response from transactionserv")
		http.Error(writer, "Invalid Request", 400)
		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}

func (webServer *WebServer) genericHandler(writer http.ResponseWriter, request *http.Request, title string) {
	fmt.Fprintf(writer, "Hello from end point %s!", request.URL.Path[1:])
}

func main() {
	serverAddress := os.Getenv("webaddr") + ":" + os.Getenv("webport")
	auditAddr := "http://" + os.Getenv("auditaddr") + ":" + os.Getenv("auditport")

	webServer := &WebServer{
		Name:              "webserver",
		transactionNumber: 0,
		userSessions:      make(map[string]*usersessions.UserSession),
		transmitter:       transmitter.NewTransmitter(os.Getenv("transaddr"), os.Getenv("transport")),
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

	fmt.Printf("Successfully started server on %s\n", serverAddress)
	http.ListenAndServe(serverAddress, nil)
}
