package api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	db "training/db/sqlc"
	"training/file-search/util"

	"training/file-index/pb"
	"training/file-search/token"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	docs "training/file-search/docs/swagger"

	"github.com/casbin/casbin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config             util.Config
	store              db.Store
	tokenMaker         token.Maker
	router             *gin.Engine
	fileSearcherClient pb.FileIndexClient
	enforcer           *casbin.Enforcer
	role               string
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA
	pemServerCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	enforcer := casbin.NewEnforcer("rbac_model.conf", "rbac_policy.csv")
	// Explicitly load policies from the adapter if necessary
	err = enforcer.LoadPolicy()
	fmt.Println("Loaded policies: ", enforcer.GetPolicy())

	if err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		enforcer:   enforcer,
	}

	tlsCredential, err := loadTLSCredentials()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.Dial(config.GRPCServerAddress, grpc.WithTransportCredentials(tlsCredential))
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	server.fileSearcherClient = pb.NewFileIndexClient(conn)
	fmt.Println(server.fileSearcherClient)
	server.setupRouter()
	return server, nil
}

// @title File Searcher Web Server API
// @version 2.0
// @description Server for file searching and managing users
// @host localhost:8082
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func (server *Server) setupRouter() {
	router := gin.Default()
	server.router = router
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		c.Header("Access-Control-Allow-Origin", origin) // Set the received Origin as the allowed origin
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, crossdomain")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// @Security BearerAuth
	router.POST("/login", server.loginUser)

	authRoutes := router.Group("/api/v1").Use(server.authMiddleware(server.tokenMaker, server.enforcer))
	{
		// Use Authorize middleware to check if user has permission to access the route
		authRoutes.GET("/files", server.getFileSearcher)
		authRoutes.GET("/users", server.getUsers)
		authRoutes.POST("/users", server.createUser)
		authRoutes.PATCH("/users", server.updateUser)
		authRoutes.DELETE("/users/:id", server.deleteUser)
		authRoutes.GET("/users/:id", server.getUserById)
	}

	docs.SwaggerInfo.Host = server.config.HTTPServerAddress
	// Register swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	address = strings.Split(address, ":")[1]
	return server.router.Run(":" + address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
