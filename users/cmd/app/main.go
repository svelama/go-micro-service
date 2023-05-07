package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/svelama/go-micro-service/users/pkg/models/mongo"
	mDB "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type application struct {
	errLog  *log.Logger
	infoLog *log.Logger
	users   *mongo.UserModel
}

func main() {

	// define command line flags
	serverAddress := flag.String("serverAddr", "", "Http server network address")
	serverPort := flag.Int("serverPort", 4000, "Http server network port")
	mongoURI := flag.String("mongoURI", "mongo://localhost:27017", "Database hostname url")
	mongoDatabase := flag.String("mongoDB", "users", "DB name")
	enableCreds := flag.Bool("enableCreds", false, "Enable the use of credentials for mongo connection")
	flag.Parse()

	// create a logger for writing information and error messages
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create a mongo client config
	c := options.Client().ApplyURI(*mongoURI)
	if *enableCreds {
		c.Auth = &options.Credential{
			Username: os.Getenv("MONGODB_USERNAME"),
			Password: os.Getenv("MONGODB_PASSWORD"),
		}
	}

	// Establish database connection
	client, err := mDB.NewClient(c)
	if err != nil {
		errLog.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		errLog.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	infoLog.Printf("Database connection established")

	// Initialize a new instance of application containing the dependencies.
	app := &application{
		infoLog: infoLog,
		errLog:  errLog,
		users: &mongo.UserModel{
			C: client.Database(*mongoDatabase).Collection("users"),
		},
	}

	// Initialize a new http.Server struct.
	serverURI := fmt.Sprintf("%s:%d", *serverAddress, *serverPort)
	srv := &http.Server{
		Addr:         serverURI,
		ErrorLog:     errLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", serverURI)
	err = srv.ListenAndServe()
	errLog.Fatal(err)

}
