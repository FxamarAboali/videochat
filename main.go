package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/centrifugal/centrifuge"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nkonev/videochat/client"
	. "github.com/nkonev/videochat/logger"
	"github.com/nkonev/videochat/utils"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"net/http"
	"strings"
	"time"
)

type staticMiddleware echo.MiddlewareFunc
type authMiddleware echo.MiddlewareFunc

func main() {
	configFile := utils.InitFlag("./config-dev/config.yml")
	utils.InitViper(configFile)

	app := fx.New(
		fx.Logger(Logger),
		fx.Provide(
			configureMongo,
			client.NewRestClient,
			configureCentrifuge,
			configureEcho,
			configureStaticMiddleware,
			configureAuthMiddleware,
		),
		fx.Invoke(runCentrifuge, runEcho),
	)
	app.Run()

	Logger.Infof("Exit program")
}

func handleLog(e centrifuge.LogEntry) {
	Logger.Printf("%s: %v", e.Message, e.Fields)
}

func configureCentrifuge(lc fx.Lifecycle) *centrifuge.Node {
	// We use default config here as starting point. Default config contains
	// reasonable values for available options.
	cfg := centrifuge.DefaultConfig
	// In this example we want client to do all possible actions with server
	// without any authentication and authorization. Insecure flag DISABLES
	// many security related checks in library. This is only to make example
	// short. In real app you most probably want authenticate and authorize
	// access to server. See godoc and examples in repo for more details.
	cfg.ClientInsecure = false
	// By default clients can not publish messages into channels. Setting this
	// option to true we allow them to publish.
	cfg.Publish = true

	// Centrifuge library exposes logs with different log level. In your app
	// you can set special function to handle these log entries in a way you want.
	cfg.LogLevel = centrifuge.LogLevelDebug
	cfg.LogHandler = handleLog

	// Node is the core object in Centrifuge library responsible for many useful
	// things. Here we initialize new Node instance and pass config to it.
	node, _ := centrifuge.New(cfg)

	// ClientConnected node event handler is a point where you generally create a
	// binding between Centrifuge and your app business logic. Callback function you
	// pass here will be called every time new connection established with server.
	// Inside this callback function you can set various event handlers for connection.
	node.On().ClientConnected(func(ctx context.Context, client *centrifuge.Client) {
		// Set Subscribe Handler to react on every channel subscribtion attempt
		// initiated by client. Here you can theoretically return an error or
		// disconnect client from server if needed. But now we just accept
		// all subscriptions.
		client.On().Subscribe(func(e centrifuge.SubscribeEvent) centrifuge.SubscribeReply {
			Logger.Printf("client id=%v, userId=%v subscribes on channel %s", client.ID(), client.UserID(), e.Channel)
			return centrifuge.SubscribeReply{}
		})

		// Set Publish Handler to react on every channel Publication sent by client.
		// Inside this method you can validate client permissions to publish into
		// channel. But in our simple chat app we allow everyone to publish into
		// any channel.
		client.On().Publish(func(e centrifuge.PublishEvent) centrifuge.PublishReply {
			Logger.Printf("client publishes into channel %s: %s", e.Channel, string(e.Data))
			return centrifuge.PublishReply{}
		})

		// Set Disconnect Handler to react on client disconnect events.
		client.On().Disconnect(func(e centrifuge.DisconnectEvent) centrifuge.DisconnectReply {
			Logger.Printf("client disconnected")
			return centrifuge.DisconnectReply{}
		})

		// In our example transport will always be Websocket but it can also be SockJS.
		transportName := client.Transport().Name()
		// In our example clients connect with JSON protocol but it can also be Protobuf.
		transportEncoding := client.Transport().Encoding()

		Logger.Printf("client connected via %s (%s)", transportName, transportEncoding)
	})

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// do some work on application stop (like closing connections and files)
			Logger.Infof("Stopping centrifuge")
			return node.Shutdown(ctx)
		},
	})

	return node
}

func runCentrifuge(node *centrifuge.Node) {
	// Run node.
	Logger.Infof("Starting centrifuge...")
	go func() {
		if err := node.Run(); err != nil {
			Logger.Fatalf("Error on start centrifuge: %v", err)
		}
	}()
	Logger.Info("Centrifuge started.")
}

type authResult struct {
	userId int
	userLogin string
}

func authorize(request *http.Request, httpClient client.RestClient) (*authResult, bool, error) {
	whitelistStr := viper.GetStringSlice("auth.exclude")
	whitelist := utils.StringsToRegexpArray(whitelistStr)
	if utils.CheckUrlInWhitelist(whitelist, request.RequestURI) {
		return nil, true, nil
	}

	sessionCookie, err := request.Cookie(utils.SESSION_COOKIE)
	if err != nil {
		Logger.Infof("Error get '%v' cookie: %v", utils.SESSION_COOKIE, err)
		return nil, false, nil
	}

	authUrl := viper.GetString(utils.AUTH_URL)
	// check cookie
	req, err := http.NewRequest(
		"GET", authUrl, nil,
	)
	if err != nil {
		Logger.Errorf("Error during create request: %v", err)
		return nil, false, err
	}

	req.AddCookie(sessionCookie)
	req.Header.Add("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		Logger.Errorf("Error during requesting auth backend: %v", err)
		return nil, false, err
	}
	defer resp.Body.Close()

	// put user id, user name to context
	b := resp.Body
	decoder := json.NewDecoder(b)
	var decodedResponse interface{}
	err = decoder.Decode(&decodedResponse)
	if err != nil {
		Logger.Errorf("Error during decoding json: %v", err)
		return nil, false, err
	}

	if resp.StatusCode == 401 {
		return nil, false, nil
	} else if resp.StatusCode == 200 {
		dto := decodedResponse.(map[string]interface{})
		i, ok := dto["id"].(float64)
		if !ok {
			Logger.Errorf("Error during casting to int")
			return nil, false, errors.New("Error during casting to int")
		}
		str := fmt.Sprintf("%v", dto["login"])
		return &authResult{userId: int(i), userLogin: str}, false, nil
	} else {
		Logger.Errorf("Unknown auth status %v", resp.StatusCode)
		return nil, false, errors.New(fmt.Sprintf("Unknown auth status %v", resp.StatusCode))
	}

}

func configureAuthMiddleware(httpClient client.RestClient) authMiddleware {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authResult, whitelist, err := authorize(c.Request(), httpClient)
			if err != nil {
				return err
			} else if whitelist {
				return next(c)
			} else if authResult == nil {
				return c.JSON(http.StatusUnauthorized, &utils.H{"status": "unauthorized"})
			} else {
				c.Set(utils.USER_PRINCIPAL_DTO, authResult)
				return next(c)
			}
		}
	}
}

func centrifugeAuthMiddleware(h http.Handler, httpClient client.RestClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authResult, _, err := authorize(r, httpClient)
		if err != nil {
			Logger.Errorf("Error during try to authenticate centrifuge request: %v", err)
			return
		} else if authResult == nil {
			Logger.Errorf("Not authenticated centrifuge request")
			return
		} else {
			ctx := r.Context()
			newCtx := centrifuge.SetCredentials(ctx, &centrifuge.Credentials{
				UserID: fmt.Sprintf("%v", authResult.userId),
				ExpireAt: time.Now().Unix() + 10,
				Info:     []byte(fmt.Sprintf("{\"login\": \"%v\"}", authResult.userLogin)),
			})
			r = r.WithContext(newCtx)
			h.ServeHTTP(w, r)
		}
	})
}

func configureEcho(staticMiddleware staticMiddleware, lc fx.Lifecycle, node *centrifuge.Node, httpClient client.RestClient) *echo.Echo {
	bodyLimit := viper.GetString("server.body.limit")

	e := echo.New()
	e.Logger.SetOutput(Logger.Writer())

	e.Pre(echo.MiddlewareFunc(staticMiddleware))

	accessLoggerConfig := middleware.LoggerConfig{
		Output: Logger.Writer(),
		Format: `"remote_ip":"${remote_ip}",` +
			`"method":"${method}","uri":"${uri}",` +
			`"status":${status},"error":"${error}","latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out},"user_agent":"${user_agent}"` + "\n",
	}
	e.Use(middleware.LoggerWithConfig(accessLoggerConfig))
	e.Use(middleware.Secure())
	e.Use(middleware.BodyLimit(bodyLimit))

	e.GET("/connection/websocket", convert(centrifugeAuthMiddleware(centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{}), httpClient)))

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// do some work on application stop (like closing connections and files)
			Logger.Infof("Stopping server")
			return e.Shutdown(ctx)
		},
	})

	return e
}

func convert(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func configureStaticMiddleware() staticMiddleware {
	box := rice.MustFindBox("static").HTTPBox()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqUrl := c.Request().RequestURI
			if reqUrl == "/" || reqUrl == "/index.html" || reqUrl == "/favicon.ico" || strings.HasPrefix(reqUrl, "/build") || strings.HasPrefix(reqUrl, "/assets") {
				http.FileServer(box).
					ServeHTTP(c.Response().Writer, c.Request())
				return nil
			} else {
				return next(c)
			}
		}
	}
}

func configureMongo(lc fx.Lifecycle) *mongo.Client {
	mongoClient := utils.GetMongoClient()
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// do some work on application stop (like closing connections and files)
			Logger.Infof("Stopping mongo client")
			return mongoClient.Disconnect(ctx)
		},
	})

	return mongoClient
}

// rely on viper import and it's configured by
func runEcho(e *echo.Echo) {
	address := viper.GetString("server.address")

	Logger.Info("Starting server...")
	// Start server in another goroutine
	go func() {
		if err := e.Start(address); err != nil {
			Logger.Infof("server shut down: %v", err)
		}
	}()
	Logger.Info("Server started. Waiting for interrupt (2) (Ctrl+C)")
}
