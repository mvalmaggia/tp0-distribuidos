package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const HEADER_SIZE = 8

func writeAll(conn net.Conn, data []byte) error {
	total := 0
	for total < len(data) {
		n, err := conn.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

// SendMessage sends a length-prefixed message over the connection.
func SendMessage(conn net.Conn, message string) error {
    data := []byte(message)
    header := []byte(fmt.Sprintf("%08d", len(data))) // 8-byte ASCII length
    fullMessage := append(header, data...)

    return writeAll(conn, fullMessage)
}

// ReceiveMessage reads a length-prefixed message from the connection and returns it as a string.
func ReceiveMessage(conn net.Conn) (string, error) {
	header := make([]byte, HEADER_SIZE)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", fmt.Errorf("failed to read length header: %w", err)
	}

	length := binary.BigEndian.Uint32(header)
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return "", fmt.Errorf("failed to read payload: %w", err)
	}

	return string(payload), nil
}