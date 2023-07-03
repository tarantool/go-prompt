import subprocess
import uuid


class Tmux:
    def __init__(self):
        self.uuid = uuid.uuid4().hex

    def setup(self, executable, params):
        x = "100"
        y = "30"
        args = None
        if "x" in params:
            x = params["x"]
        if "y" in params:
            y = params["y"]
        if "args" in params:
            args = params["args"]
        cmd = ["tmux", "-L", self.uuid, "new-session", "-d", "-x", x,
               "-y", y, executable]
        if args is not None:
            cmd += args

        subprocess.run(cmd)

    def dump_screen(self):
        subprocess.run(["tmux", "-L", self.uuid, "capture-pane"])
        sub = subprocess.run(["tmux", "-L", self.uuid, "show-buffer"],
                             capture_output=True)
        subprocess.run(["tmux", "-L", self.uuid, "delete-buffer"])

        return sub.stdout.decode("utf-8")

    def dump_workspace(self):
        return self.dump_screen().rstrip("\n")

    def get_cursor(self):
        sub = subprocess.run(["tmux", "-L", self.uuid, "display-message", "-p",
                              "#{cursor_x} #{cursor_y}"], capture_output=True)
        output = sub.stdout.decode("utf-8")
        x, y = list(map(int, output.split(" ")))
        return x, y

    def send_key(self, key):
        sub = subprocess.run(["tmux", "-L", self.uuid, "send-keys", key])
        return sub.returncode

    def close(self):
        sub = subprocess.run(["tmux", "-L", self.uuid, "kill-server"])
        assert sub.returncode == 0
