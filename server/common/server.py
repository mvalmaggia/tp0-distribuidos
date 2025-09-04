import socket
import logging
import signal
import threading
import time

from codec.codec import decode_bet_batch, encode_winners
from protocol.protocol import receive_message, send_ack, send_message
from common.utils import store_bets, load_bets, has_won

class Server:
    def __init__(self, port, listen_backlog, expected_agencies=5):
        
        self._client_threads = []
        self._lock = threading.Lock()

        self._finished_agencies = set()
        self._expected_agencies = expected_agencies

        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._winners_by_agency = {}

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
                t = threading.Thread(
                    target=self.__handle_client_connection, 
                    args=(self._client_sock,)
                )
                t.start()
                self._client_threads.append(t)
                self._client_sock = None
            
            alive_threads = []
            for t in self._client_threads:
                if t.is_alive():
                    alive_threads.append(t)
                else:
                    t.join()
            self._client_threads = alive_threads

        for t in self._client_threads:
            t.join()
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

    def _get_winners_by_agency(self):
        """
        Retrieves dictiornary with agency ids as keys and list of document ids as values
        for all the winning bets stored in the STORAGE_FILEPATH file.
        """
        with self._lock:
            self._winners_by_agency = {}
    
            bets = load_bets()
            for bet in bets:
                if has_won(bet):
                    agency = bet.agency
                    if agency not in self._winners_by_agency:
                        self._winners_by_agency[agency] = []
                    self._winners_by_agency[agency].append(bet.document)

    def _get_winners_for_agency(self, agency_id: int):
        """
        Retrieves a list of document ids for all the winning bets
        for a specific agency.
        """
        logging.info(f"action: get_winners_for_agency | result: in_progress | agency: {agency_id}")
        return self._winners_by_agency.get(agency_id, [])

    def _handle_get_winners(self, client_sock, agency_id):
        if len(self._finished_agencies) != self._expected_agencies:
            send_message(client_sock, "ERROR:NOT_ALL_BATCHES_RECEIVED")
            return

        winners = self._get_winners_for_agency(agency_id)
        send_message(client_sock, encode_winners(winners))
        logging.info(f"action: send_winners | result: success | agency: {agency_id}")

    def _handle_batch_bet(self, encoded_msg):
        bets = decode_bet_batch(encoded_msg)

        with self._lock:
            store_bets(bets)

        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

    def _handle_batch_end(self, agency_id):
        with self._lock:
            self._finished_agencies.add(agency_id)
            logging.info(f"action: batch_end_received | result: success | agency: {agency_id}")

        if len(self._finished_agencies) == self._expected_agencies:
            logging.info("action: sorteo | result: success")
            self._get_winners_by_agency()


    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            encoded_msg = receive_message(client_sock)
            if encoded_msg.startswith("GET_WINNERS"):
                logging.info("action: get_winners_received | result: success")
                agency_id = int(encoded_msg.split(":", 1)[1].strip())
                self._handle_get_winners(client_sock, agency_id)
                return

            if encoded_msg.startswith("BET_BATCH"):
               self._handle_batch_bet(encoded_msg)

            if encoded_msg.startswith("BATCH_END"):
                agency_id = encoded_msg.split(":", 1)[1].strip()
                self._handle_batch_end(agency_id)
                return
            
            addr = client_sock.getpeername()
            send_ack(client_sock)
            logging.info(f'action: send_ack | result: success | ip: {addr[0]}')
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
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
                time.sleep(0.1) 
                return c
            except socket.timeout:
                # Silent timeout, just continue the loop
                continue
            
        return None