import io
import os
import signal
import socket
import errno
import sys


class WSGIServer:
    REQUEST_QUEUE_SIZE = 5

    def __init__(self, server_address):
        # For references:
        # https://www.ibm.com/docs/en/i/7.4?topic=family-af-inet-address
        # https://stackoverflow.com/questions/5815675/what-is-sock-dgram-and-sock-stream
        self.listen_socket = listen_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        listen_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        listen_socket.bind(server_address)
        listen_socket.listen(self.REQUEST_QUEUE_SIZE)
        signal.signal(signal.SIGCHLD, grim_reaper)

        host, port = listen_socket.getsockname()[:2]
        self.server_name = socket.getfqdn(host)
        self.server_port = port
        # Return headers set by Web framework/Web application
        self.headers = []

    def set_app(self, application):
        self.application = application

    def serve_forever(self):
        print(f'Serving HTTP on {self.server_name}:{self.server_port} with PID {os.getpid()}')
        sock = self.listen_socket
        while True:
            try:
                client_connection, client_address = sock.accept()
            except IOError as e:
                code, msg = e.args
                if code == errno.EINTR:
                    continue
                else:
                    raise

            pid = os.fork()
            if pid == 0:  # child
                sock.close()  # close child copy
                self.handle_one_request(client_connection)
                client_connection.close()
                os._exit(0)  # child exits here
            else:  # parent
                client_connection.close()  # close parent copy and loop over

    def handle_one_request(self, client_connection: socket.socket):
        request_data = client_connection.recv(1024)
        request_text = request_data.decode('utf-8')

        request_line = request_text.splitlines()[0]
        method, visit_path, version = request_line.rstrip('\r\n').split()

        env = {
            # Required WSGI variables
            'wsgi.version': (1, 0),
            'wsgi.url_scheme': 'http',
            'wsgi.input': io.BytesIO(request_data),
            'wsgi.errors': sys.stderr,
            'wsgi.multithread': False,
            'wsgi.multiprocess': True,
            'wsgi.run_once': False,
            # Required CGI variables
            'REQUEST_METHOD': method,
            'PATH_INFO': visit_path,
            'SERVER_NAME': self.server_name,
            'SERVER_PORT': str(self.server_port)
        }

        result = self.application(env, self.start_response)

        try:
            status, response_headers = self.headers
            response = f'HTTP/1.1 {status}\r\n'
            for header in response_headers:
                response += '{0}: {1}\r\n'.format(*header)
            response += '\r\n'
            for data in result:
                response += data.decode('utf-8')
            response_bytes = response.encode()
            client_connection.sendall(response_bytes)
        finally:
            client_connection.close()

    def start_response(self, status, response_headers, exc_info=None):
        # Add necessary server headers
        server_headers = [
            ('Server', 'simple WSGIServer 0.1'),
        ]
        self.headers = [status, server_headers + response_headers]
        # To adhere to WSGI specification the start_response must return
        # a 'write' callable. For simplicity's sake we'll ignore that detail
        # for now.
        # return self.finish_response


def grim_reaper(signum, frame):
    while True:
        try:
            pid, status = os.waitpid(
                -1,  # Wait for any child process
                os.WNOHANG  # Do not block and return EWOULDBLOCK error
            )
        except OSError:
            return

        if pid == 0:  # no more zombies
            return


def make_server(server_address, application):
    server = WSGIServer(server_address)
    server.set_app(application)
    return server


SERVER_ADDRESS = (HOST, PORT) = '', 8888

if __name__ == '__main__':
    if len(sys.argv) < 2:
        sys.exit('Provide a WSGI application object as module:callable')
    app_path = sys.argv[1]
    module_name, application_name = app_path.split(':')
    module = __import__(module_name)
    app = getattr(module, application_name)
    httpd = make_server(SERVER_ADDRESS, app)
    httpd.serve_forever()
