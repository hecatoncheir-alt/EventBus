package engine

import (
	"encoding/json"
	"fmt"
	"github.com/hecatoncheir/Broker"
	"log"
	"net"
	"os"
)

type Socket struct {
	APIVersion, Host string
	Port             int
	TCPHost          *net.TCPAddr
	Log              *log.Logger

	ConnectedClients []*Client
}

func New(APIVersion string) *Socket {
	engine := Socket{APIVersion: APIVersion}

	logPrefix := fmt.Sprintf("Server")
	engine.Log = log.New(os.Stdout, logPrefix, 3)

	return &engine
}

func (socket *Socket) SetUp(ip string, port int) error {
	address := fmt.Sprintf("%v:%v", ip, port)
	tcpAddress, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return err
	}

	socket.Host = ip
	socket.Port = port
	socket.TCPHost = tcpAddress
	return nil
}

func (socket *Socket) Listen() {
	listener, err := net.ListenTCP("tcp", socket.TCPHost)
	if err != nil {
		socket.Log.Fatalf("Fatal error: %s", err.Error())
		os.Exit(1)
	}

	socket.Log.Printf("Socket server listen on: %v:%v", socket.Host, socket.Port)

	for {
		connection, err := listener.Accept()
		if err != nil {
			continue
		}

		client := NewClient(connection)

		socket.ConnectedClients = append(socket.ConnectedClients, client)

		socket.Log.Printf("Client: %v connected. Connected clients: %v",
			client.ID, len(socket.ConnectedClients))

		go socket.SubscribeOnClientEvents(client)
	}
}

func (socket *Socket) RemoveConnectedClient(client *Client) {
	var index int

	for i, connectedClient := range socket.ConnectedClients {
		if connectedClient.ID == client.ID {
			index = i
			break
		}
	}

	clients := socket.ConnectedClients
	socket.ConnectedClients = append(clients[:index], clients[index+1:]...)
}

func (socket *Socket) SubscribeOnClientEvents(client *Client) {
	for event := range client.InputChannel {
		if event.Message == "Connection closed" {
			socket.RemoveConnectedClient(client)
			break
		}

		if event.Message == "Need APIVersion" {
			eventData := map[string]string{"APIVersion": socket.APIVersion}
			bytes, err := json.Marshal(eventData)
			if err != nil {
				socket.Log.Printf("Error marshal event data: %v for client: %v",
					eventData, client.ID)
				continue
			}

			client.Write(broker.EventData{Message: "APIVersion ready", Data: string(bytes)})

			continue
		}

		socket.WriteToAllConnectedClients(event)
	}
}

func (socket *Socket) WriteToAllConnectedClients(data broker.EventData) {
	for _, client := range socket.ConnectedClients {
		if client.ID != data.ClientID {
			client.Write(data)
		}
	}
}
