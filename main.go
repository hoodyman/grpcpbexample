package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/hoodyman/grpcpbexample/GrpcPbExmpl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

type SrvImpl struct {
	GrpcPbExmpl.UnimplementedGrpcPbExmplServer
}

func (s *SrvImpl) Ping(ctx context.Context, req *GrpcPbExmpl.Message) (*GrpcPbExmpl.Message, error) {
	x := new(GrpcPbExmpl.Message)
	x.MessageData = req.MessageData
	return x, nil
}

func main() {

	/// SERVER PART

	srv_cert, err := tls.LoadX509KeyPair("srv-cert.pem", "srv-key.pem")
	if err != nil {
		grpclog.Errorln("srv load cert:", err)
		return
	}

	srv_config := &tls.Config{}
	srv_config.Certificates = []tls.Certificate{srv_cert}

	srv_creds := credentials.NewTLS(srv_config)

	srv := grpc.NewServer(grpc.Creds(srv_creds))
	chCheckStopping := make(chan int)
	defer func() {
		srv.Stop()
		<-chCheckStopping
	}()

	go func() {
		defer func() { chCheckStopping <- 0 }()
		l, err := net.Listen("tcp", ":6060")
		if err != nil {
			grpclog.Errorln("srv listen:", err)
			return
		}
		GrpcPbExmpl.RegisterGrpcPbExmplServer(srv, &SrvImpl{})
		err = srv.Serve(l)
		if err != nil && err != grpc.ErrServerStopped {
			grpclog.Errorln("srv serve:", err)
		}
	}()

	/// CLIENT PART

	cli_cert, err := ioutil.ReadFile("srv-cert.pem")
	if err != nil {
		grpclog.Errorln("cli read cert:", err)
		return
	}

	cli_cert_pool := x509.NewCertPool()
	if !cli_cert_pool.AppendCertsFromPEM(cli_cert) {
		grpclog.Errorln("cli pool:", err)
		return
	}

	cli_config := &tls.Config{}
	cli_config.RootCAs = cli_cert_pool

	cli_conn, err := grpc.Dial(":6060", grpc.WithTransportCredentials(credentials.NewTLS(cli_config)))
	if err != nil {
		grpclog.Errorln("cli dial:", err)
		return
	}

	cli := GrpcPbExmpl.NewGrpcPbExmplClient(cli_conn)
	mess, err := cli.Ping(context.Background(), &GrpcPbExmpl.Message{MessageData: "Hello GRPCPB!"})
	if err != nil {
		grpclog.Errorln("cli ping:", err)
		return
	}
	fmt.Println(mess.MessageData)

}
