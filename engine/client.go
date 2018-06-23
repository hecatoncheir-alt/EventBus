package engine

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hecatoncheir/Broker"
	"net"
)

type Client struct {
	ID            string
	Connection    net.Conn
	InputChannel  chan broker.EventData
	OutputChannel chan broker.EventData
}

func NewClient(connection net.Conn) *Client {
	id := uuid.New().String()

	client := Client{
		ID:            id,
		Connection:    connection,
		InputChannel:  make(chan broker.EventData),
		OutputChannel: make(chan broker.EventData)}

	go func() {
		for output := range client.OutputChannel {
			client.write(output)
		}
	}()

	//client.Connection.SetReadDeadline(time.Now().Add(2 * time.Minute))
	go client.SubscribeOnEvents()
	return &client
}

func (client *Client) SubscribeOnEvents() {
	request := make([]byte, 128)

	defer client.Connection.Close()

	for {
		lengthOfBytes, err := client.Connection.Read(request)

		if err != nil || lengthOfBytes == 0 {
			client.InputChannel <- broker.EventData{Message: "Connection closed"}
			break // connection already closed by client
		}

		event := broker.EventData{}
		err = json.Unmarshal(request[:lengthOfBytes], &event)
		if err != nil {
			println(fmt.Sprintf("Error by unmarshal event: %v", request[:lengthOfBytes]))
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
