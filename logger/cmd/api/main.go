package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"logger/data"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoUrl = "mongodb://mongo:27017"
	grpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	//connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Fatal(err)
	}
	client = mongoClient

	// create a context to disconnect (it is needed by mongo)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	//close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := &Config{
		Models: data.New(client),
	}

	//registering rpc server
	err = rpc.Register(new(RPCserver))

	go app.rpcListen() //start rpc server
	app.serve()        //web server
}

//serve will start a webserver
func (c *Config) serve() {
	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: c.routes(),
	}

	log.Println("LOGGER server starting on port " + webPort)
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

//rpcListen will start a rpc server
func (c *Config) rpcListen() {
	log.Println("RPC server starting on port " + rpcPort)
	listen, err := net.Listen("tcp", "0.0.0.0:"+rpcPort)
	if err != nil {
		log.Println("impossible to start rpc server")
		log.Panic(err)
	}

	defer listen.Close()

	for {
		rpcConnection, err := listen.Accept()
		if err != nil {
			log.Println("error accepting rpc connection")
			continue
		}

		go rpc.ServeConn(rpcConnection)
	}
}

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoUrl)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})
	//connect
	connection, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("error connecting to mongodb", err)
		return nil, err
	}

	return connection, nil
}
