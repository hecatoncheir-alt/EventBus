package engine

import (
	"encoding/json"
	"fmt"
	"github.com/hecatoncheir/Broker"
	"net"
	"os"
)

type Socket struct {
	APIVersion, Host string
	Port             int
	TCPHost          *net.TCPAddr

	ConnectedClients []*Client
}

func New(APIVersion string) *Socket {
	engine := Socket{APIVersion: APIVersion}
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
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	println(fmt.Sprintf("Socket server listen on: %v:%v", socket.Host, socket.Port))

	for {
		connection, err := listener.Accept()
		if err != nil {
			continue
		}

		client := NewClient(connection)

		socket.ConnectedClients = append(socket.ConnectedClients, client)

		println(fmt.Sprintf("Client: %v connected. Connected clients: %v", client.ID, len(socket.ConnectedClients)))

		socket.SubscribeOnClientEvents(client)
	}
}

func (socket *Socket) SubscribeOnClientEvents(client *Client) {
	for event := range client.InputChannel {
		if event.Message == "Connection closed" {
			var index int

			for i,connectedClient := range socket.ConnectedClients{
				if connectedClient.ID == client.ID {
					index = i
					break
				}
			}

			clients := socket.ConnectedClients
			socket.ConnectedClients = append(clients[:index], clients[index+1:]...)

			break
		}

		if event.Message == "Need APIVersion" {
			eventData := map[string]string{"APIVersion": socket.APIVersion}
			bytes, err := json.Marshal(eventData)
			if err != nil {
				println(fmt.Sprintf("Error marshal event data: %v for client: %v",
					eventData, client.ID))
				continue
			}

			client.Write(broker.EventData{Message: "APIVersion ready", Data: string(bytes)})
		}
	}
}
