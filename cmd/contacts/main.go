package main

import (
	"flag"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/atlas-contacts-app/pb/contacts"
	"github.com/infobloxopen/atlas-contacts-app/server/contacts"
)

// Addr is the address and port to listen on
var Addr string

// Dsn is the database connection string
var Dsn string

func main() {
	logger := logrus.New()

	ln, err := net.Listen("tcp", Addr)
	if err != nil {
		logger.Fatalln(err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer( // middleware chain
				grpc_validator.UnaryServerInterceptor(),                     // validation middleware
				grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logger)), // logging middleware
			),
		),
	)
	s, err := contacts.NewBasicServer(Dsn)
	if err != nil {
		logger.Fatalln(err)
	}
	pb.RegisterContactsServer(server, s)

	server.Serve(ln)
}

func init() {
	flag.StringVar(&Addr, "listen", "0.0.0.0:9091", "")
	flag.StringVar(&Dsn, "dsn", "", "")
	flag.Parse()
}
