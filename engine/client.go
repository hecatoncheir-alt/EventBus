package engine

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hecatoncheir/Broker"
	"log"
	"net"
	"os"
)

type Client struct {
	ID            string
	Connection    net.Conn
	InputChannel  chan broker.EventData
	OutputChannel chan broker.EventData
	Log           *log.Logger
}

func NewClient(connection net.Conn) *Client {
	id := uuid.New().String()

	client := Client{
		ID:            id,
		Connection:    connection,
		InputChannel:  make(chan broker.EventData),
		OutputChannel: make(chan broker.EventData)}

	logPrefix := fmt.Sprintf("Client: %v", client.ID)
	client.Log = log.New(os.Stdout, logPrefix, 3)

	go func() {
		for outputEvent := range client.OutputChannel {
			err := client.write(outputEvent)
			if err != nil {
				client.Log.Printf("Error write event: %v to client", outputEvent)
			}
		}
	}()

	//client.Connection.SetReadDeadline(time.Now().Add(2 * time.Minute))
	go client.SubscribeOnEvents(connection)
	return &client
}

func (client *Client) SubscribeOnEvents(connection net.Conn) {
	request := make([]byte, 1024)

	defer connection.Close()

	for {
		lengthOfBytes, err := connection.Read(request)

		if err != nil || lengthOfBytes == 0 {
			client.InputChannel <- broker.EventData{Message: "Connection closed"}
			break // connection already closed by client
		}

		event := broker.EventData{}
		err = json.Unmarshal(request[:lengthOfBytes], &event)
		if err != nil {
			client.Log.Printf("Error by unmarshal event: %v", request[:lengthOfBytes])
			continue
		}

		event.ClientID = client.ID

		client.InputChannel <- event
	}
}

func (client *Client) write(data broker.EventData) error {
	encodedData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = client.Connection.Write(encodedData)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) Write(data broker.EventData) {
	client.OutputChannel <- data
}
