from tmux import Tmux


class PromptApp(Tmux):
    def __init__(self):
        super().__init__()
        self.socket_uri = None
        self.socket = None

    def setup(self, prompt_app, params):
        server_socket = params["server_socket"]
        super().setup(prompt_app, params)

        self.socket_uri = params["server_socket_uri"]

        # Wait for the prompt_app to connect.
        self.socket, _ = server_socket.accept()

    def send_keys(self, keys):
        if not isinstance(keys, list):
            keys = [keys]

        for key in keys:
            super().send_key(key)
            # Poll rendering socket.
            self.socket.recv(1)
