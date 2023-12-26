package client

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/pkg/errors"

	"playGround/utils/network"
	"playGround/utils/protocol"
	"playGround/utils/sha256"
	"playGround/utils/tracing"
)

type Config struct {
	serverAddr *net.TCPAddr
	workersNum int
}

func (c *Config) Load(ctx context.Context) error {
	span, _ := tracing.NewSpan(ctx, "client.Config.Load")
	defer span.Close()

	const serverAddrEnvVarName = "WISDOM_CLIENT_SERVER_ADDR"
	serverAddrValue := os.Getenv(serverAddrEnvVarName)
	if serverAddrValue == "" {
		return errors.Errorf("empty env var (%s) value", serverAddrEnvVarName)
	}
	parsedServerAddr, err := net.ResolveTCPAddr(protocol.TCPType, serverAddrValue)
	if err != nil {
		return errors.Wrapf(err, "parsing value (%s) from env var (%s)", serverAddrValue, serverAddrEnvVarName)
	}
	c.serverAddr = parsedServerAddr

	const workersNumEnvVarName = "WISDOM_CLIENT_WORKERS_NUM"
	workersNumValue := os.Getenv(workersNumEnvVarName)
	if workersNumValue == "" {
		return errors.Errorf("empty env var (%s) value", workersNumValue)
	}
	if c.workersNum, err = strconv.Atoi(workersNumValue); err != nil {
		return errors.Wrapf(err, "parsing value (%s) from env var (%s)", workersNumValue, workersNumEnvVarName)
	}

	return nil
}

func New(ctx context.Context, config *Config) *Client {
	span, _ := tracing.NewSpan(ctx, "client.New")
	defer span.Close()

	return &Client{
		config:   config,
		stopChan: make(chan struct{}),
		logger:   log.New(os.Stdout, "client", log.Llongfile),
	}
}

type Client struct {
	config   *Config
	stopChan chan struct{}
	logger   *log.Logger
}

func (c *Client) Start(ctx context.Context) {
	span, _ := tracing.NewSpan(ctx, "client.Client.Start")
	defer span.Close()

	for i := 0; i < c.config.workersNum; i++ {
		go c.processor(ctx)
	}
}

func (c *Client) Stop(ctx context.Context) {
	span, _ := tracing.NewSpan(ctx, "client.Client.Stop")
	defer span.Close()

	close(c.stopChan)
}

func (c *Client) processor(ctx context.Context) {
	span, _ := tracing.NewSpan(ctx, "client.Client.processor")
	defer span.Close()

	for {
		select {
		case <-c.stopChan:
		default:
			if err := c.query(ctx); err != nil {
				c.logger.Println("query execution",
					"error", err,
				)
			}

		}
	}
}

func (c *Client) query(ctx context.Context) error {
	span, _ := tracing.NewSpan(ctx, "client.Client.query")
	defer span.Close()

	conn, err := net.DialTCP("tcp", nil, c.config.serverAddr)
	if err != nil {
		return errors.Wrap(err, "create connection")
	}
	defer conn.Close()

	if err := network.Send(ctx, conn, protocol.ClientWelcomeRequest{
		Message: protocol.Message{
			Type: protocol.MessageTypeClientWelcome,
		},
	}); err != nil {
		return errors.Wrap(err, "sending initial request")
	}

	serverQuestionRequest, err := network.Receive[protocol.ServerQuestionRequest](ctx, conn)
	if err != nil {
		return errors.Wrap(err, "receiving server question request")
	}
	if serverQuestionRequest.Type != protocol.MessageTypeServerQuestion {
		return errors.Errorf("server question requrest: received wrong message (%s)", serverQuestionRequest)
	}

	resultNonce := sha256.Find(ctx, serverQuestionRequest.Prefix, serverQuestionRequest.Difficulty)
	if err := network.Send(ctx, conn, protocol.ClientAnswerResponse{
		Message: protocol.Message{
			Type: protocol.MessageTypeClientAnswer,
		},
		Nonce:      resultNonce,
		Prefix:     serverQuestionRequest.Prefix,
		Difficulty: serverQuestionRequest.Difficulty,
	}); err != nil {
		return errors.Wrap(err, "sending client answer response")
	}

	serverResultResponse, err := network.Receive[protocol.ServerResultResponse](ctx, conn)
	if err != nil {
		return errors.Wrap(err, "receiving server result response")
	}
	if serverResultResponse.Type != protocol.MessageTypeServerResult {
		return errors.Errorf("server result response: received wrong message (%s)", serverResultResponse)
	}
	if !serverResultResponse.Success {
		log.Print("server error response",
			" error: ", serverResultResponse.Payload,
		)

		return nil
	}

	log.Print("server success response",
		" response: ", serverResultResponse.Payload,
	)

	return nil
}
