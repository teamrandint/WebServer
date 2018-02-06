package transmitter

import (
	"fmt"
	"net"
	"log"
	"bufio"
)

type Transmitters interface {
	MakeRequest() string

}

type Transmitter struct {
	address string
	port string
	connection net.Conn
}

func NewTransmitter(addr string, prt string) *Transmitter {
	transmitter := new(Transmitter)
	transmitter.address = addr
	transmitter.port = prt
	// Create a connection to the specified server
	conn, err := net.Dial("tcp", addr + ":" + prt)

	if err != nil {
		// Error in connection 
		log.Fatal(err)
	} else {
		transmitter.connection = conn
	}

	return transmitter
}

func (trans *Transmitter) MakeRequest(message string) string {
	// Send through socket.
	fmt.Fprintf(trans.connection, message)
	reply, _ := bufio.NewReader(trans.connection).ReadString('\n')

	// TODO: Process response and append failed responses to some kind of list
	// fmt.Print("Message from server: "+ reply)

	return reply
}