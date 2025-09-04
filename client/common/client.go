package common

import (
    "net"
    "os"
    "os/signal"
    "syscall"
    "strings"
    "fmt"
    "time"

    "github.com/op/go-logging"
    "github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
    "github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
    "github.com/7574-sistemas-distribuidos/docker-compose-init/client/codec"
)

var log = logging.MustGetLogger("log")

const MAX_AMOUNT_POLLS = 5
const BETS_DATASET_PATH = "./dataset.csv"

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

func setupSignalHandler(c *Client) {
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
}

func HandleEndOfBatch(c *Client) {
    // if err := c.createClientSocket(); err != nil {
    //     log.Errorf("action: handle_end_of_batch | result: fail | error: %v", err)
    //     return
    // }

    if err := protocol.SendMessage(c.conn, fmt.Sprintf("BATCH_END:%s", c.config.ID)); err != nil {
        log.Errorf("action: send_end_of_batch | result: fail | error: %v", err)
        c.conn.Close()
        return
    }

    c.conn.Close()

    polls := 0
    for polls < MAX_AMOUNT_POLLS {
        if err := c.createClientSocket(); err != nil {
            log.Errorf("action: handle_end_of_batch | result: fail | error: %v", err)
            return
        }

        if err := protocol.SendMessage(c.conn, fmt.Sprintf("GET_WINNERS:%s", c.config.ID)); err != nil {
            log.Errorf("action: get_winners | result: fail | error: %v", err)
            c.conn.Close()
            return
        }

        msg, err := protocol.ReceiveMessage(c.conn)
        if err != nil {
            log.Errorf("action: get_winners | result: fail | error: %v", err)
            polls++
            time.Sleep(3 * time.Second)
            continue
        }

        // Check if the message is an error message
        if strings.HasPrefix(msg, "ERROR:") {
            errorMsg := strings.TrimPrefix(msg, "ERROR:")
            if errorMsg == "NOT_ALL_BATCHES_RECEIVED" {
                log.Infof("action: get_winners | result: waiting | reason: not_all_batches_received")
                c.conn.Close()
                polls++
                time.Sleep(3 * time.Second)
                continue
            }
        } else {
            // Process successful response
            winners, _ := codec.DecodeWinners(msg)
            log.Infof("action: get_winners | result: success | cant_ganadores: %d", len(winners))
            c.conn.Close()  
            return
        }
        log.Warningf("action: get_winners | result: fail | reason: max_polls_reached")
    }
}    

func HandleReceiveAck(c *Client) bool {
    msg, err := protocol.ReceiveMessage(c.conn)
    c.conn.Close()
    log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)

    if err != nil {
        log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
            c.config.ID, err)
        return false
    }

    if strings.TrimSpace(msg) == "ACK" {
        return true
    } else {
        return false
    }
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClient(bet model.ClientBet) {
    setupSignalHandler(c)

    betsParser, err := NewBetParser(c.config.ID, BETS_DATASET_PATH)
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

        if err := c.createClientSocket(); err != nil {
            continue
        }

        if len(bets) == 0 {
            log.Infof("action: no_more_bets | result: success | client_id: %v", c.config.ID)
            HandleEndOfBatch(c)
            break
        }

        encodedBets := codec.EncodeBetBatch(bets)
        // log.Infof("action: send_bet_batch | result: in_progress | bets: %v | batch_size: %v",
            // encodedBets, len(bets))
        if err := protocol.SendMessage(c.conn, encodedBets); err != nil {
            log.Errorf("action: send_bet_batch | result: fail | client_id: %v | error: %v",
                c.config.ID, err)
            c.conn.Close()
            continue
        }

        receivedAck := HandleReceiveAck(c)

        if receivedAck {
            log.Infof("action: apuesta_batch_enviada | result: success | client_id: %v | batch_size: %v",
                c.config.ID, len(bets))
        } else {
            log.Warningf("action: receive_ack | result: fail | client_id: %v",
                c.config.ID)
        }
    }
    os.Exit(0)
}