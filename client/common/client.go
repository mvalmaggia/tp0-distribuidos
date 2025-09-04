package common

import (
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"strings"

	"github.com/op/go-logging"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/codec"
)

const MAX_RETRIES = 8
const ACK_TIMEOUT = 2 * time.Second

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config  ClientConfig
	conn    net.Conn
	running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:  config,
		running: true,
	}
	return client
}

// createClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and error is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(bet model.ClientBet) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigs
		log.Infof("action: signal_received | result: in_progress | signal: %v", sig)
		c.running = false
		if c.conn != nil {
			c.conn.Close()
			log.Infof("action: shutdown_client_socket | result: success")
		}
		os.Exit(0)
	}()

	// Create the connection once
	if err := c.createClientSocket(); err != nil {
		return
	}
	defer c.conn.Close()

	success := false
	for attempt := 1; attempt <= MAX_RETRIES && c.running; attempt++ {
		// Send bet
		if err := protocol.SendMessage(c.conn, codec.EncodeBet(bet)); err != nil {
			log.Errorf("action: send_bet | result: fail | client_id: %v | attempt: %v | error: %v",
				c.config.ID, attempt, err)
			continue
		}

		// Wait for server response with timeout
		c.conn.SetReadDeadline(time.Now().Add(ACK_TIMEOUT))
		msg, err := protocol.ReceiveMessage(c.conn)
		if err != nil {
			log.Warningf("action: receive_message | result: fail | client_id: %v | attempt: %v | error: %v",
				c.config.ID, attempt, err)
			continue
		}

		// Validate ACK
		if strings.TrimSpace(msg) == "ACK" {
			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v | attempt: %v",
				bet.ID, bet.Number, attempt)
			success = true
			break
		} else {
			log.Warningf("action: receive_ack | result: fail | client_id: %v | attempt: %v | unexpected_msg: %v",
				c.config.ID, attempt, msg)
		}
	}

	if !success {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | reason: no_ack_received",
			c.config.ID)
	} else {
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}
}

