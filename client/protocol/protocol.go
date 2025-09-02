package protocol

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"your_project/model"
)

// Serialize ClientBet with key/value pairs separated by |
func serializeBet(bet model.ClientBet) string {
	// Format date to YYYY-MM_DD
	date := bet.Birthdate.Format("2006-01-02")

	return fmt.Sprintf("dni:%d|numero:%d|nombre:%s|apellido:%s|fecha:%s\n",
		bet.ID, bet.Number, bet.Name, bet.Lastname, date)
}

func SendBet(conn net.Conn, bet model.ClientBet) error {
	writer := bufio.NewWriter(conn)
	message := serializeBet(bet)

	if _, err := writer.WriteString(message); err != nil {
		return fmt.Errorf("failed to write bet: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush bet: %w", err)
	}

	return nil
}

func ReceiveBet(conn net.Conn) (model.ClientBet, error) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n') // read until newline
	if err != nil {
		return model.ClientBet{}, fmt.Errorf("failed to read bet: %w", err)
	}
	return deserializeBet(line)
}