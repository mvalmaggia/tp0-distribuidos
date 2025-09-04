import socket

HEADER_SIZE = 8  # 8 characters for message length

def send_message(conn: socket.socket, message: str) -> None:
    """
    Send a message with a length header.
    """
    data = message.encode("utf-8")
    length_str = f"{len(data):0{HEADER_SIZE}}".encode("utf-8")
    conn.sendall(length_str + data)

def receive_message(conn: socket.socket) -> str:
    """
    Receive a message with length header.
    """
    # Read the header
    header = read_bytes(conn, HEADER_SIZE)
    length = int(header.decode("utf-8"))

    # Read the payload
    payload = read_bytes(conn, length)
    return payload.decode("utf-8")

def read_bytes(conn: socket.socket, length):
    buffer = b""
    while len(buffer) < length:
        chunk = conn.recv(length - len(buffer))
        if not chunk:
            raise ConnectionError("Connection closed while reading payload")
        buffer += chunk

    return buffer


def send_ack(conn: socket.socket) -> None:
    """
    Send ACK back to connection
    """
    send_message(conn, "ACK")
