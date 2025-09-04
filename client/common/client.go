package common

import (
    "net"
    "os"
    "os/signal"
    "syscall"
    "strings"

    "github.com/op/go-logging"
    "github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
    "github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
    "github.com/7574-sistemas-distribuidos/docker-compose-init/client/codec"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
    ID            string
    ServerAddress string
    MaxBatchAmount int
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
func (c *Client) StartClient(bet model.ClientBet) {
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

    betsParser, err := NewBetParser(c.config.ID, "./dataset.csv")
    if err != nil {
        log.Errorf("action: create bet parser | result: fail | error: %v", err)
        return
    }
    defer betsParser.Close()

    for c.running {
        bets, err := betsParser.NextBatch(c.config.MaxBatchAmount)
        if err != nil {
            log.Errorf("action: read_bets | result: fail | client_id: %v | error: %v",
                c.config.ID, err)
            continue
        }

        if len(bets) == 0 {
            log.Infof("action: no_more_bets | result: success | client_id: %v", c.config.ID)
            break
        }

        encodedBets := codec.EncodeBetBatch(bets)

        if err := c.createClientSocket(); err != nil {
            continue
        }

        if err := protocol.SendMessage(c.conn, encodedBets); err != nil {
            log.Errorf("action: send_bet_batch | result: fail | client_id: %v | error: %v",
                c.config.ID, err)
            c.conn.Close()
            continue
        }

        msg, err := protocol.ReceiveMessage(c.conn)
        c.conn.Close()
        log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)

        if err != nil {
            log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
                c.config.ID, err)
            continue
        }

        if strings.TrimSpace(msg) == "ACK" {
            log.Infof("action: apuesta_batch_enviada | result: success | client_id: %v | batch_size: %v",
                c.config.ID, len(bets))
            os.Exit(0)
        } else {
            log.Warningf("action: receive_ack | result: fail | client_id: %v | unexpected_msg: %v",
                c.config.ID, msg)
            os.Exit(1)
        }
    }
}