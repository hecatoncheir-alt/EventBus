package engine

import (
	"encoding/json"
	"fmt"
	"github.com/hecatoncheir/Broker"
	"github.com/hecatoncheir/Configuration"
	"net"
	"testing"
	"time"
)

func TestSocket_SetUp(t *testing.T) {
	config := configuration.New()
	engine := New(config.APIVersion)

	if engine.APIVersion != config.APIVersion {
		t.Fail()
	}

	err := engine.SetUp(config.Development.EventBus.Host, config.Development.EventBus.Port)
	if err != nil {
		t.Error(err)
	}

	if engine.Host != "localhost" {
		t.Fail()
	}

	if engine.Port != 8282 {
		t.Fail()
	}

	if engine.TCPHost.Port != 8282 {
		t.Fail()
	}
}

func TestSocket_SubscribeOnClientEvents(t *testing.T) {
	config := configuration.New()
	engine := New(config.APIVersion)

	err := engine.SetUp(config.Development.EventBus.Host, config.Development.EventBus.Port)
	if err != nil {
		t.Error(err)
	}

	go engine.Listen()

	address := fmt.Sprintf("%v:%v", config.Development.EventBus.Host, config.Development.EventBus.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		t.Error(err)
	}

	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second * 1)
	if len(engine.ConnectedClients) != 1 {
		t.Fail()
	}

	event := broker.EventData{Message: "Need APIVersion"}
	encodedEvent, err := json.Marshal(event)
	if err != nil {
		t.Error(err)
	}

	_, err = connection.Write(encodedEvent)
	if err != nil {
		t.Error(err)
	}

	request := make([]byte, 128)

	for {
		length, err := connection.Read(request)
		if err != nil {
			t.Error(err)
			break
		}

		if length == 0 {
			t.Error(err)
			break
		}

		decodedEvent := broker.EventData{}

		err = json.Unmarshal(request[:length], &decodedEvent)
		if err != nil {
			t.Error(err)
			break
		}

		if decodedEvent.Message == "APIVersion ready" {
			decodedData := map[string]string{}

			err = json.Unmarshal([]byte(decodedEvent.Data), &decodedData)
			if err != nil {
				t.Error(err)
				break
			}

			if decodedData["APIVersion"] != "1.0.0" {
				t.Error(err)
				break
			}

			break
		}
	}

	connection.Close()

	if len(engine.ConnectedClients) != 0 {
		t.Fail()
	}
}

func TestSocket_WriteToAllConnectedClients(t *testing.T) {
	config := configuration.New()
	engine := New(config.APIVersion)

	err := engine.SetUp(config.Development.EventBus.Host, config.Development.EventBus.Port)
	if err != nil {
		t.Error(err)
	}

	go engine.Listen()

	address := fmt.Sprintf("%v:%v", config.Development.EventBus.Host, config.Development.EventBus.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		t.Error(err)
	}

	firstClientConnection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		t.Error(err)
	}

	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		t.Error(err)
	}

	event := broker.EventData{Message: "Test message for other connected client"}
	encodedEvent, err := json.Marshal(event)
	if err != nil {
		t.Error(err)
	}

	_, err = connection.Write(encodedEvent)
	if err != nil {
		t.Error(err)
	}

	request := make([]byte, 1024)

	for {
		length, err := firstClientConnection.Read(request)
		if err != nil {
			t.Error(err)
			break
		}

		if length == 0 {
			t.Error(err)
			break
		}

		decodedEvent := broker.EventData{}

		err = json.Unmarshal(request[:length], &decodedEvent)
		if err != nil {
			t.Error(err)
			break
		}

		if decodedEvent.Message == "Test message for other connected client" {
			break
		}

		if decodedEvent.Message != "Test message for other connected client" {
			t.Fail()
			break
		}
	}

	connection.Close()
	firstClientConnection.Close()
}
