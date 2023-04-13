package assets

import "C"
import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/kycaid"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Kyc - The Kyc struct is used to store information related to Know Your Customer (KYC) processes. It contains an API key, as
// well as three structs (S, P, and C) that store information related to the documents required to complete a KYC
// process. These documents may include a passport, proof of address, and proof of identity. Each struct contains a key
// and multiplication value, which is likely used to specify the number of documents that need to be submitted for the
// KYC process. The Kyc struct also contains a RedirectUrl field, which is likely used to redirect the user to a specific
// page after the KYC process is completed.
type Kyc struct {
	ApiKey string
	Forms  struct {
		S struct {
			Key            string
			Multiplication int32
		}
		P struct {
			Key            string
			Multiplication int32
		}
		C struct {
			Key            string
			Multiplication int32
		}
	}
	RedirectUrl string
}

// Smtp - The type Smtp struct is used to store information about a Simple Mail Transfer Protocol (SMTP) connection. It contains
// fields for the SMTP host name, sender address, password, and port number. This information can be used to establish an
// SMTP connection, send emails, and receive emails.
type Smtp struct {
	Host, Sender, Password string
	Port                   int
}

// Server - The type Server struct is a data structure in the Go programming language that holds two strings, Host and Proxy. It
// is used to represent a server with both a host name and a proxy name. It can be used to store configuration details
// for a server, such as host and proxy settings. It can also be used to store information about the server such as its
// address, port, and other settings.
type Server struct {
	Host, Proxy string
}

// Redis - The type Redis struct is a data structure used to store information about a Redis server. It contains fields to store the
// host address, password, and the database number that is being accessed. This data structure is used to establish a
// connection to a Redis server and can be used to store and retrieve data.
type Redis struct {
	Host, Password string
	DB             int
}

// Rabbitmq - The type Rabbitmq struct is a data structure used to store the information needed to connect to a RabbitMQ server. It stores
// the host URL, username, password, and a boolean value that indicates whether the connection should be
// established with a clean session. This information is then used by clients to connect and interact with RabbitMQ.
type Rabbitmq struct {
	Host, Username, Password string
	CleanSession             bool
}

// The Credentials struct is used to store authentication credentials such as a certificate, secret key, and override. It
// allows the data to be organized and accessed more easily.
type Credentials struct {
	Crt, Key, Override string
}

type Context struct {

	// Development bool is a boolean value used to check if the current environment is a development environment or not.
	// This is commonly used in software development to distinguish between the environment used for development and the
	// environment used for production. It is used to ensure that certain features are enabled or disabled depending on the environment.
	Development bool

	// Logger *logrus.Logger is a type of logger that is used to record events and errors in a program. It helps developers
	// to track down issues, monitor application performance, and audit user activity. It can also be used to provide
	// detailed debugging information, which is useful for troubleshooting issues.
	Logger *logrus.Logger

	// A mutex is a synchronization mechanism used to control access to shared resources in a multi-threaded environment. It
	// is used to ensure that only one thread can access a resource at a given time, thus preventing race conditions and
	// data corruption. The sync.Mutex object allows threads to take ownership of the resource, thus ensuring that no other
	// threads can access it until the owner thread has finished its work.
	Mutex sync.Mutex

	// The purpose of the above code is to create an array of strings called Secrets. This array can be used to store
	// various secrets, such as passwords, encryption keys, or other sensitive information.
	Secrets []string

	// StoragePath string is a variable used to store the path of a directory or file in a string format. It can be used to
	// access the contents of the file or directory.
	StoragePath string

	// The PostgresConnect string is a connection string used to establish a connection to a PostgreSQL database. It
	// typically contains information such as the server name, port, database name, and authentication credentials.
	PostgresConnect string

	// Timezones is a string that is used to store information about the different time zones that a person may need to
	// interact with. It can be used to help convert between different time zones and to manage scheduling of events.
	Timezones string

	// KYC: KYC (Know Your Customer) is a process used by organizations to verify the identity of their customers and assess their suitability for doing business.
	// SMTP: Simple Mail Transfer Protocol (SMTP) is a protocol for sending emails. It is used to transfer emails from one server to another over the Internet.
	// Server: A server is a computer that provides services to other computers or clients on a network. It can host websites, applications and other services.
	// Redis: Redis is an in-memory data structure store that is used as a database, cache and message broker. It supports data structures such as strings, hashes, lists, sets, sorted sets with range queries and bitmaps.
	// Rabbitmq: RabbitMQ is an open source message broker software that implements the Advanced Message Queuing Protocol (AMQP). It is used to send, receive and store messages between applications.
	// Credentials: Credentials are information used to authenticate a user or entity to gain access to a system or service. Examples include usernames, passwords, and security tokens.
	// RabbitmqClient: MQTT.Client is a client library for the Message Queue Telemetry
	// RedisClient: This is a Redis client which is used for caching and storing data in a NoSQL key-value store.
	// GrpcClient: This is a gRPC client which is used for communicating with a remote server using a high-performance RPC protocol.
	// Db: This is a SQL database which is used for storing and managing relational data.
	// KycProvider: This is a KYC provider which is used to verify the identity of users for compliance with anti-money laundering regulations.

	Kyc            *Kyc
	Smtp           *Smtp
	Server         *Server
	Redis          *Redis
	Rabbitmq       *Rabbitmq
	Credentials    *Credentials
	RabbitmqClient MQTT.Client
	RedisClient    *redis.Client
	GrpcClient     *grpc.ClientConn
	Db             *sql.DB
	KycProvider    *kycaid.Api
}

// This function is used to set up the application context. It locks the mutex, reads the configuration file, sets the
// timezone, initializes the logger, opens the PostgresQL and Redis databases, and connects to the RabbitMQ broker.
// Finally, it unlocks the mutex and returns the application context.
func (app *Context) Write() *Context {

	// The purpose of app.Mutex.Lock() is to prevent multiple threads or processes from accessing a shared resource
	// simultaneously. This is done by obtaining a lock that allows only one thread to access the resource at a time. By
	// doing this, it ensures that data is not corrupted or overwritten by multiple threads accessing it simultaneously.
	app.Mutex.Lock()

	// This code is used to read a file located at the location stored in the app.ConfigPath() variable. It uses the
	// ioutil.ReadFile() function to read the contents of the file as a byte array and store it in the serialize variable.
	// If there is an error encountered while reading the file, to err variable will contain the error and the
	// logrus.Fatal() function is used to log the error and terminate the program.
	serialize, err := ioutil.ReadFile(app.ConfigPath())
	if err != nil {
		logrus.Fatal(err)
	}

	// This code is used to convert the serialized data in the variable "serialize" into the "app" variable using the
	// json.Unmarshal method. If an error occurs, it is logged using logrus.Fatal and the program will terminate.
	if err = json.Unmarshal(serialize, &app); err != nil {
		logrus.Fatal(err)
	}

	// This code is loading a timezone from the app.Timezones variable and setting it as the local timezone. If an error
	// occurs when loading the location, it will log the error and exit the program.
	loc, err := time.LoadLocation(app.Timezones)
	if err != nil {
		logrus.Fatal(err)
	}
	time.Local = loc

	// The app.Logger = logrus.New() statement assigns an instance of a logrus logger to the app.Logger variable. This
	// allows you to use the logrus logger to log messages from your application.
	app.Logger = logrus.New()

	// This code sets the formatter for the Logger to logrus.TextFormatter. The ForceColors option will force the logger to
	// output logs with colors, even if they are not displayed in a terminal.
	app.Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	// The purpose of this code is to open the file "writer.log" with read-write permissions, create it if it does not
	// exist, and append to it if it does exist. If there is an error opening the file, the app.Logger.Fatalf function is
	// called to log the error.
	writer, err := os.OpenFile(app.StoragePath+"/writer.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		app.Logger.Fatalf("error opening file: %v", err)
	}

	// This statement sets the output of the app.Logger object to both os.Stdout and writer. This means that any log
	// messages will be printed to the standard output and written to the writer object.
	app.Logger.SetOutput(io.MultiWriter(os.Stdout, writer))

	// The purpose of app.Logger.SetLevel is to set the logging level for the application. This enables the application to
	// log different types of messages depending on their severity. The different log levels are ErrorLevel, FatalLevel, and
	// WarnLevel. ErrorLevel will log errors, FatalLevel will log fatal errors, and WarnLevel will log warnings.
	app.Logger.SetLevel(logrus.ErrorLevel)
	app.Logger.SetLevel(logrus.FatalLevel)
	app.Logger.SetLevel(logrus.WarnLevel)

	// The purpose of the code is to set the log level for the application depending on if it is in development or not. If
	// the application is in development, the log level is set to InfoLevel and DebugLevel, otherwise it is set to whatever the default log level is.
	if app.Development {
		app.Logger.SetLevel(logrus.InfoLevel)
		app.Logger.SetLevel(logrus.DebugLevel)
	}

	// This code is used to open a connection to a PostgreSQL database using the app.PostgresConnect parameter for the
	// connection string. The connection is stored in a variable called app.Db and is used for subsequent database
	// operations. If an error occurs during the opening of the connection, it is logged using the logrus library and the
	// program exits with a fatal error.
	app.Db, err = sql.Open("postgres", app.PostgresConnect)
	if err != nil {
		logrus.Fatal(err)
	}

	// Ping the database to ensure a connection is successful.
	err = app.Db.Ping()
	if err != nil {
		logrus.Fatal(err)
	}

	// The code above is creating a new redis client connection with the specified Redis host, password, and DB from the
	// app. It allows the app to interact with Redis and perform operations such as retrieving or setting data.
	app.RedisClient = redis.NewClient(&redis.Options{
		Addr:     app.Redis.Host,
		Password: app.Redis.Password,
		DB:       app.Redis.DB,
	})

	// This code is establishing a connection to a RabbitMQ server with the given credentials and settings. The purpose of
	// this is to allow for communication between the RabbitMQ server and the application. The code also checks to see if
	// the connection was successful, and if not, it prints an error message.
	app.RabbitmqClient = MQTT.NewClient(MQTT.NewClientOptions().
		AddBroker(app.Rabbitmq.Host).
		SetUsername(app.Rabbitmq.Username).
		SetPassword(app.Rabbitmq.Password).
		SetCleanSession(app.Rabbitmq.CleanSession).
		SetKeepAlive(2 * time.Second).
		SetPingTimeout(1 * time.Second))
	if connect := app.RabbitmqClient.Connect(); connect.Wait() && connect.Error() != nil {
		logrus.Fatal(connect.Error())
	}

	// This code snippet is creating a new API instance for the app.KycProvider variable. It is using the kycaid.NewApi()
	// function to do this. If an error occurs, the logrus.Fatal() function is called to log the error and terminate the program.
	app.KycProvider, err = kycaid.NewApi(app.Kyc, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	// App.Mutex.Unlock() is a function that unlocks a mutex, which is a synchronization primitive that allows only one
	// thread to access a shared resource at a time. It is used to ensure that multiple threads do not access a shared
	// resource simultaneously, which can cause unexpected results.
	app.Mutex.Unlock()

	return app
}

// Auth - This function is used to authenticate users in a context-based application. It uses JWT to parse the authorization
// token from the incoming context and uses the secret key stored in the application's Secrets to validate the token.
// Once the token is validated, it returns the user's personal data that was previously encoded.
func (app *Context) Auth(ctx context.Context) (int64, error) {

	// The purpose of this code is to extract the metadata from the incoming context (ctx) and assign it to the meta
	// variable. Metadata is a key-value map containing information about the context, such as details about the request, the user, etc.
	meta, _ := metadata.FromIncomingContext(ctx)

	// This code checks if the "authorization" field of the "meta" object has a length of 1 and is not nil. If both of these
	// conditions are not met, the code returns an error with code 10010 and the message "missing metadata". This is likely
	// used to ensure that the "authorization" field is set correctly and is not empty before continuing with further processing.
	if len(meta["authorization"]) != 1 && meta["authorization"] == nil {
		return 0, status.Error(10010, "missing metadata")
	}

	// This line of code is used to parse a JWT token from an authorization header. It takes the authorization header value
	// and splits it into two parts, taking the second part as the token. It then uses the token and the app.Secrets[0] byte
	// array to parse the token. If the token is valid, it will return the token, otherwise it will return an error.
	token, err := jwt.Parse(strings.Split(meta["authorization"][0], "Bearer ")[1], func(token *jwt.Token) (interface{}, error) {
		return []byte(app.Secrets[0]), nil
	})
	if err != nil {
		return 0, err
	}

	// This code is used to extract a key from a JSON Web Token (JWT). The code is checking if the claims are of type
	// jwt.MapClaims and that the token is valid. If so, it extracts the value of the "sub" key and converts it to an int64.
	// The purpose of this code is to get a user's ID from the JWT so that the application can identify the user and grant
	// them access to the appropriate resources.
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int64(claims["sub"].(float64)), nil
	}

	return 0, nil
}

// Publish - This function is used to publish data to a specific topic on a given channel.
// It takes in a data interface, a topic string, and a variable list of channel strings.
// It uses the json package to marshal the data interface into a string.
// It then iterates through the channel list and uses the RabbitmqClient to publish the data string to the given topic on the given channel.
// Finally, it returns nil if the publication is successful.
func (app *Context) Publish(data interface{}, topic string, channel ...string) error {

	// The Marshal struct is used to store data in a standardized format which is capable of being encoded and decoded as
	// JSON. The structure contains two fields: Channel and Data. The Channel field is a string that identifies the source
	// or destination of the data, while the Data field is a string containing the actual data. This structure can be used
	// to easily pass data between different applications or services, as it is a commonly accepted data format.
	type Marshal struct {
		Channel string `json:"channel"`
		Data    string `json:"data"`
	}

	// This code is used in a for loop to iterate through the elements of a channel. The for loop sets the variable 'i' to 0
	// and then checks to see if 'i' is less than the length of the channel. If it is, it will then execute the code within
	// the loop and then increment 'i' by 1. The loop continues to run until 'i' is no longer less than the length of the channel.
	for i := 0; i < len(channel); i++ {

		// This code is attempting to marshal (convert) a data object into a JSON object. The json.Marshal function will return
		// the serialized JSON object as the first return value and any errors that occurred as the second. If an error
		// occurred, the code will return the error to the caller.
		serialize, err := json.Marshal(data)
		if err != nil {
			return err
		}

		// This code is serializing a struct of type Marshal, which contains two fields, Channel and Data. The Channel field is
		// set to the value of the i-th element of the channel array, and the Data field is set to the value of the serialize
		// variable. If any errors occur while serializing, the code returns an error.
		serialize, err = json.Marshal(Marshal{
			Channel: channel[i],
			Data:    string(serialize),
		})
		if err != nil {
			return err
		}

		// The purpose of this code is to publish a message to a given topic using the app.RabbitmqClient. The message is
		// serialized as a byte(2) and the boolean false indicates that the message is not retained by the broker. The
		// string(serialize) is used to indicate the serialized message that needs to be published.
		app.RabbitmqClient.Publish(topic, byte(2), false, string(serialize))
	}

	return nil
}

// Recovery - This function is used to recover from unexpected errors in a Go application. It takes an interface as an argument and
// returns an error. It generates an error message with the given expression included, which allows the application to
// identify the source of the unexpected error.
func (app *Context) Recovery(expr interface{}) error {
	return status.Errorf(codes.Internal, "Unexpected error: (%+v)", expr)
}

// Debug - This function is used to debug the application context. It takes an expression as a parameter and logs it based on its
// type. If the expression is an error type, it will log an error level message. Otherwise, it will log a debug level
// message. The function also includes the file and line number from where the logging was called. This helps to trace any issues in the application.
func (app *Context) Debug(expr interface{}) bool {

	// This code is used to log errors and debug information in an application. The runtime.Caller(1) statement allows the
	// application to retrieve the file and line number of the code that is currently running. The switch statement then
	// checks the type of the expression (expr) passed in and either logs an error, returns false if the expression is nil,
	// or logs a debug message and returns true.
	if _, file, line, ok := runtime.Caller(1); ok {
		switch expr.(type) {
		case error:
			app.Logger.WithFields(logrus.Fields{"file": file, "line": line}).Error(expr)
			return true
		case nil:
			return false
		default:
			app.Logger.WithFields(logrus.Fields{"file": file, "line": line}).Debug(expr)
			return true
		}
	}

	return false
}

// ConfigPath - This function is used to find the location of the config.json file. It first checks if the working directory contains the "cross" string, and if it does,
// the path is set to "../config.json" otherwise, it is set to "./config.json".
// If the config.json file is found at either path, the path is returned. If the file is not found, it throws a panic.
func (app *Context) ConfigPath() (path string) {

	// The purpose of the line of code is to define the path of a file named config.json. This file can contain
	// configuration information to be used in the program.
	path = fmt.Sprintf("%v/%v", app.StoragePath, "config.json")

	// This code is intended to check for the existence of a path and panic in case it doesn't exist. The first if statement
	// checks is a path exists and returns it if it does. The else if statement checks if the error is specifically
	// os.ErrNotExist, and panics if it is. The else statement handles any other error and panics with the err as its argument.
	if _, err := os.Stat(path); err == nil {
		return path
	} else if errors.Is(err, os.ErrNotExist) {
		panic("Config not found")
	} else {
		panic(err)
	}
}
