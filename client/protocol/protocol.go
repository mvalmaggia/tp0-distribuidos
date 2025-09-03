package protocol

import (
	"bufio"
	"fmt"
	"net"
)

// sends bytes over the connection
func SendMessage(conn net.Conn, message string) error {
	writer := bufio.NewWriter(conn)
	// log.infof("Sending message: %s", message)
	if _, err := writer.WriteString(message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush message: %w", err)
	}
	return nil
}

// a full message until \n
func ReceiveMessage(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read message: %w", err)
	}
	return line, nil
}
