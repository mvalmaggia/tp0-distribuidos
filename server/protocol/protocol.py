import socket

def send_message(conn: socket.socket, message: str) -> None:
    """
    Send a message over the socket, ensuring it ends with '\n'.
    """
    if not message.endswith("\n"):
        message += "\n"
    conn.sendall(message.encode("utf-8"))

def receive_message(conn: socket.socket) -> str:
    """
    Receive a message terminated by '\n'.
    Reads until newline.
    """
    buffer = b""
    while True:
        chunk = conn.recv(1)
        if not chunk:
            break  # connection closed
        buffer += chunk
        if chunk == b"\n":
            break
    return buffer.decode("utf-8").strip()

def send_ack(conn: socket.socket) -> None:
    """
    Send a simple ACK back to the client.
    """
    send_message(conn, "ACK")
