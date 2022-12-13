package assets

import "C"
import (
	"context"
	"database/sql"
	"encoding/json"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Context struct {
	Development bool

	Logger *logrus.Logger
	Mutex  sync.Mutex

	StoragePath string

	Postgres string

	GrpcAddress, GrpcGatewayAddress string
	GrpcConnect                     *grpc.ClientConn

	Secrets []string

	IntervalMarket, IntervalChainStatus time.Duration

	CredentialsCrt, CredentialsKey, CredentialsOverride string

	Db *sql.DB

	Timezones string

	SmtpHost, SmtpSender, SmtpPassword string
	SmtpPort                           int

	RedisHost, RedisPassword string
	RedisDB                  int
	RedisClient              *redis.Client

	MarketKey, MarketSecret string
	MarketPairs             []string
	MarketTest              bool

	BrokerClient                               MQTT.Client
	BrokerHost, BrokerUsername, BrokerPassword string
	BrokerCleanSession                         bool
}

func (app *Context) Write() *Context {

	app.Mutex.Lock()

	serialize, err := ioutil.ReadFile(app.ConfigPath())
	if err != nil {
		logrus.Fatal(err)
	}

	if err = json.Unmarshal(serialize, &app); err != nil {
		logrus.Fatal(err)
	}

	// Convert time between different timezones.
	loc, err := time.LoadLocation(app.Timezones)
	if err != nil {
		logrus.Fatal(err)
	}
	time.Local = loc

	app.Logger = logrus.New()

	// Log as JSON instead of the default ASCII formatter.
	app.Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	writer, err := os.OpenFile("./writer.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		app.Logger.Fatalf("error opening file: %v", err)
	}

	// Output to stdout instead of the default stderr.
	// Can be any io.Writer, see below for File example.
	app.Logger.SetOutput(io.MultiWriter(os.Stdout, writer))

	// Only log the warning severity or above.
	app.Logger.SetLevel(logrus.ErrorLevel)
	app.Logger.SetLevel(logrus.FatalLevel)
	app.Logger.SetLevel(logrus.WarnLevel)

	if app.Development {
		app.Logger.SetLevel(logrus.InfoLevel)
		app.Logger.SetLevel(logrus.DebugLevel)
	}

	// PostgresQL connect and open.
	app.Db, err = sql.Open("postgres", app.Postgres)
	if err != nil {
		logrus.Fatal(err)
	}

	app.RedisClient = redis.NewClient(&redis.Options{
		Addr:     app.RedisHost,
		Password: app.RedisPassword,
		DB:       app.RedisDB,
	})

	options := MQTT.NewClientOptions().
		AddBroker(app.BrokerHost).
		SetUsername(app.BrokerUsername).
		SetPassword(app.BrokerPassword).
		SetCleanSession(app.BrokerCleanSession).
		SetKeepAlive(2 * time.Second).
		SetPingTimeout(1 * time.Second)
	app.BrokerClient = MQTT.NewClient(options)
	if connect := app.BrokerClient.Connect(); connect.Wait() && connect.Error() != nil {
		logrus.Fatal(connect.Error())
	}

	app.Mutex.Unlock()

	return app
}

// Auth - ensure valid token ensures a valid token exists within a request's metadata.
// If the token is missing or invalid, the interceptor blocks execution of the
// handler and returns an error.
func (app *Context) Auth(ctx context.Context) (int64, error) {

	// Metadata from incoming context.
	meta, _ := metadata.FromIncomingContext(ctx)
	if len(meta["authorization"]) != 1 && meta["authorization"] == nil {
		return 0, status.Error(10010, "missing metadata")
	}

	token, err := jwt.Parse(strings.Split(meta["authorization"][0], "Bearer ")[1], func(token *jwt.Token) (interface{}, error) {
		return []byte(app.Secrets[0]), nil
	})
	if err != nil {
		return 0, err
	}

	// Returns to the personal data the previous look, that were previously encoded.
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int64(claims["sub"].(float64)), nil
	}

	return 0, nil
}

// Error - caller.
func (app *Context) Error(err error) error {

	if _, file, line, ok := runtime.Caller(1); ok {
		app.Logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Error(err)
	}

	return err
}

// Publish - pusher send to socket.
func (app *Context) Publish(data interface{}, topic string, channel ...string) error {

	type Marshal struct {
		Channel string `json:"channel"`
		Data    string `json:"data"`
	}

	for i := 0; i < len(channel); i++ {

		serialize, err := json.Marshal(data)
		if err != nil {
			return err
		}

		serialize, err = json.Marshal(Marshal{
			Channel: channel[i],
			Data:    string(serialize),
		})
		if err != nil {
			return err
		}

		app.BrokerClient.Publish(topic, byte(2), false, string(serialize))
	}

	return nil
}

// Recovery - panic recovery.
func (app *Context) Recovery(expr interface{}) error {
	return status.Errorf(codes.Internal, "Unexpected error: (%+v)", expr)
}

// Debug - with debug caller.
func (app *Context) Debug(expr interface{}) bool {

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

// ConfigPath - config path.
func (app *Context) ConfigPath() string {

	dir, _ := os.Getwd()
	if strings.Contains(dir, "cross") {
		return "../config.json"
	}

	return "./config.json"
}
