package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/pkg/errors"

	"playGround/utils/protocol"
	"playGround/utils/sha256"
	"playGround/utils/tracing"
)

type Config struct {
	serverPort int
	maxConns   int
	difficulty int
}

func (c *Config) Load(ctx context.Context) error {
	span, _ := tracing.NewSpan(ctx, "server.Config.Load")
	defer span.Close()

	const serverPortEnvVarName = "WISDOM_SERVER_PORT"
	serverPortValue := os.Getenv(serverPortEnvVarName)
	if serverPortValue == "" {
		return errors.Errorf("empty env var (%s) value", serverPortEnvVarName)
	}
	parsedServerPort, err := strconv.Atoi(serverPortValue)
	if err != nil {
		return errors.Wrapf(err, "parsing value (%s) from env var (%s)", serverPortValue, serverPortEnvVarName)
	}
	c.serverPort = parsedServerPort

	const maxConnsEnvVarName = "WISDOM_SERVER_MAX_CONNS"
	maxConnsValue := os.Getenv(maxConnsEnvVarName)
	if maxConnsValue == "" {
		return errors.Errorf("empty env var (%s) value", maxConnsValue)
	}
	if c.maxConns, err = strconv.Atoi(maxConnsValue); err != nil {
		return errors.Wrapf(err, "parsing value (%s) from env var (%s)", maxConnsValue, maxConnsEnvVarName)
	}

	const difficultyEnvVarName = "WISDOM_SERVER_POW_DIFFICULTY"
	difficultyValue := os.Getenv(difficultyEnvVarName)
	if difficultyValue == "" {
		return errors.Errorf("empty env var (%s) value", difficultyValue)
	}
	if c.difficulty, err = strconv.Atoi(difficultyValue); err != nil {
		return errors.Wrapf(err, "parsing value (%s) from env var (%s)", difficultyValue, difficultyEnvVarName)
	}

	return nil
}

func New(ctx context.Context, config *Config) *Server {
	span, _ := tracing.NewSpan(ctx, "server.New")
	defer span.Close()

	return &Server{
		config:    config,
		stopChan:  make(chan struct{}),
		logger:    log.New(os.Stdout, "server", log.Llongfile),
		quoteList: []string{"example quote one", "example quote two", "example quote three"},
	}
}

type Server struct {
	config   *Config
	stopChan chan struct{}
	logger   *log.Logger
	connCnt  atomic.Int32

	listner   net.Listener
	quoteList []string
}

func (s *Server) Start(ctx context.Context) error {
	span, _ := tracing.NewSpan(ctx, "server.Server.Start")
	defer span.Close()

	var err error
	if s.listner, err = net.Listen(protocol.TCPType, fmt.Sprintf("localhost:%d", s.config.serverPort)); err != nil {
		return errors.Wrap(err, "start listener")
	}

	go s.processor(ctx)

	return nil
}

func (s *Server) processor(ctx context.Context) {
	span, _ := tracing.NewSpan(ctx, "server.Server.processor")
	defer span.Close()

	for {
		select {
		case <-s.stopChan:
		default:
			conn, err := s.listner.Accept()
			if err != nil {
				s.logger.Println("connection processing",
					"error", err,
				)
			}

			if s.connCnt.Load() > int32(s.config.maxConns) {
				if _, err := conn.Write([]byte("max conns exceeded")); err != nil {
					s.logger.Println("max conn response sending",
						"error", err,
					)
				}
				conn.Close()
				continue
			}

			s.connCnt.Add(1)
			go func() {
				defer conn.Close()
				defer s.connCnt.Add(-1)
				if err := s.serv(ctx, conn); err != nil {
					s.logger.Println("connection serving",
						"error", err,
					)
				}
			}()
		}
	}
}

func (s *Server) Stop(ctx context.Context) {
	span, _ := tracing.NewSpan(ctx, "server.Server.Stop")
	defer span.Close()

	close(s.stopChan)
}

func (s *Server) serv(ctx context.Context, conn net.Conn) error {
	span, _ := tracing.NewSpan(ctx, "server.Server.serv")
	defer span.Close()

	clientWelcome := make([]byte, 1024)
	n, err := conn.Read(clientWelcome)
	if err != nil {
		return errors.Wrap(err, "reading client welcome")
	}
	clientWelcome = clientWelcome[:n]
	clientWelcomeRequest := &protocol.ClientWelcomeRequest{}
	if err := json.Unmarshal(clientWelcome, clientWelcomeRequest); err != nil {
		return errors.Wrap(err, "unmarshalling client welcome request")
	}
	if clientWelcomeRequest.Type != protocol.MessageTypeClientWelcome {
		return errors.Errorf("client welcome request: received wrong message (%s)", clientWelcomeRequest)
	}

	serverQuestionResponse, err := json.Marshal(protocol.ServerQuestionRequest{
		Message: protocol.Message{
			Type: protocol.MessageTypeServerQuestion,
		},
		Prefix:     rand.Int63(),
		Difficulty: s.config.difficulty,
	})
	if err != nil {
		return errors.Wrap(err, "marshalling server question response")
	}
	if _, err := conn.Write(serverQuestionResponse); err != nil {
		return errors.Wrap(err, "sending server question response")
	}

	clientAnswer := make([]byte, 1024)
	n, err = conn.Read(clientAnswer)
	if err != nil {
		return errors.Wrap(err, "reading client welcome")
	}
	clientAnswer = clientAnswer[:n]
	clientAnswerResponse := &protocol.ClientAnswerResponse{}
	if err := json.Unmarshal(clientAnswer, clientAnswerResponse); err != nil {
		return errors.Wrap(err, "unmarshalling client welcome response")
	}
	if clientAnswerResponse.Type != protocol.MessageTypeClientAnswer {
		return errors.Errorf("client welcome response: received wrong message (%s)", clientAnswerResponse)
	}

	ok := sha256.Check(
		ctx,
		clientAnswerResponse.Nonce,
		clientAnswerResponse.Prefix,
		clientAnswerResponse.Difficulty,
	)

	resp := protocol.ServerResultResponse{
		Message: protocol.Message{
			Type: protocol.MessageTypeServerResult,
		},
	}
	switch {
	case ok:
		resp.Success = true
		resp.Payload = s.quoteList[rand.Intn(2)]
	default:
		resp.Payload = "invalid result"
	}

	serverResultResponse, err := json.Marshal(resp)
	if err != nil {
		return errors.Wrap(err, "marshalling server result response")
	}
	if _, err := conn.Write(serverResultResponse); err != nil {
		return errors.Wrap(err, "sending server result response")
	}

	return nil
}
