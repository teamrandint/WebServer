package transmitter

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

type Transmitters interface {
	MakeRequest() string
}

type Transmitter struct {
	address    string
	port       string
	connection net.Conn
}

func NewTransmitter(addr string, prt string) *Transmitter {
	transmitter := new(Transmitter)
	transmitter.address = addr
	transmitter.port = prt
  
	// Create a connection to the specified server
	var conn net.Conn
	var err error
	for err != nil {
		conn, err = net.Dial("tcp", addr+":"+prt)
		time.Sleep(time.Millisecond * 30)
		log.Fatal(err)
	}

	transmitter.connection = conn

	return transmitter
}

func (trans *Transmitter) MakeRequest(transNum int,message string) string {
	prefix := strconv.Itoa(transNum)
	message = prefix + ";" + message
	message += "\n"
	conn, err := net.Dial("tcp", trans.address+":"+trans.port)

	if err != nil {
		// Error in connection
		log.Fatal(err)
	} else {
		trans.connection = conn
	}

	// fmt.Println("Making request to transaction server")
	fmt.Fprintf(trans.connection, message)
	// fmt.Println("Waiting for response from transaction server")
	reply, _ := bufio.NewReader(trans.connection).ReadString('\n')
	// trans.connection.Close()
	return reply
}
