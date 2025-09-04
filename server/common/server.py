import socket
import logging
import signal

from server.codec.codec import decode_bet_batch
from protocol.protocol import receive_message, send_ack
from common.utils import store_bets

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        self._server_socket.settimeout(1.0)

        self._running = True
        self._client_sock = None

        # Register SIGTERM handler to this instance
        signal.signal(signal.SIGTERM, self._handle_sigterm)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._running:
            self._client_sock = self.__accept_new_connection()
            if self._client_sock:
                self.__handle_client_connection(self._client_sock)
                self._client_sock = None
        self._shutdown()

    def _handle_sigterm(self, signum, frame):
        logging.info(f'action: signal_received | result: in_progress | signal: {signum}')
        self._running = False

    def _shutdown(self):
        logging.info('in shutdown')
        if self._client_sock:
            self._client_sock.shutdown(socket.SHUT_RDWR)
            self._client_sock.close()
            logging.info('action: shutdown_client_socket | result: success')

        if self._server_socket:
            self._server_socket.close()
            logging.info('action: shutdown_server_socket | result: success')

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            encoded_msg = receive_message(client_sock)
            bets = decode_bet_batch(encoded_msg)

            store_bets(bets)
            logging.info(f'action: batch_recibido | result: in_progress | cantidad: {len(bets)}')

            addr = client_sock.getpeername()
            send_ack(client_sock)
            logging.info(f'action: send_ack | result: success | ip: {addr[0]}')
            
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
    
        while self._running:
            try:
                c, addr = self._server_socket.accept()
                logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
                return c
            except socket.timeout:
                # Silent timeout, just continue the loop
                continue
            
        return None