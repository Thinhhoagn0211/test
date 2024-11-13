package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	db "training/10-file-index/db/sqlc"
	"training/10-file-index/pb"
	"training/10-file-index/service"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// loadTLSCredentials loads TLS credentials from the files
func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}
	// Create a new cert pool and add cert client CA
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server-s certificate and private key
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server on port %d", *port)

	tlsCredential, err := loadTLSCredentials()
	if err != nil {
		log.Fatal(err)
	}
	// Open a database connection
	conn, err := sql.Open("postgres", "postgresql://root:secret@localhost:5432/everything_pg?sslmode=disable")
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	// Initialize the store
	store := db.NewStore(conn)
	// Create a new file store
	fileStore := service.NewInMemoryFileStore(store)
	// Register the file store with the file discovery server
	fileDiscoveryServer := service.NewFileDiscoveryServer(fileStore)
	// Create a new gRPC server
	grpcServer := grpc.NewServer(grpc.Creds(tlsCredential))
	pb.RegisterFileIndexServer(grpcServer, fileDiscoveryServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
