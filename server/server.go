package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
	t "time"

	gRPC "github.com/Luc-13/dsysexcersises/proto"
	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedTempServer

	name string
	port string

	incrementValue int64
	mutex          sync.Mutex
}

var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "5400", "Server port")           // set with "-port <port>" in terminal

func main() {
	flag.Parse()
	fmt.Println(".:server is starting:.")

	go launchServer()

	for {
		time.Sleep(time.Second * 5)
	}
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &Server{
		name:           *serverName,
		port:           *port,
		incrementValue: 0, // gives default value, but not sure if it is necessary
	}

	gRPC.RegisterTempServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening on port %s\n", *serverName, *port)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func (s *Server) Increase(ctx context.Context, Amount *gRPC.Amount) (*gRPC.Ack, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.incrementValue += int64(Amount.GetVal())
	return &gRPC.Ack{Time: t.Now().String()}, nil
}

func (s *Server) Greet(msgStream gRPC.Temp_GreetServer) error {
	for {
		msg, err := msgStream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		log.Printf("Recieved msg from %s: %s", msg.Name, msg.Msg)
	}
	ack := &gRPC.MsgBack{Msg: "Bye"}

	msgStream.SendAndClose(ack)

	return nil
}
