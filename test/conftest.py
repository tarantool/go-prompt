import os
import socket
import subprocess
import tempfile
import uuid

import py
import pytest
from prompt_app import PromptApp


# ######## #
# Fixtures #
# ######## #
def get_tmpdir(request):
    tmpdir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: tmpdir.remove(rec=1))
    return str(tmpdir)


@pytest.fixture(scope="session")
def session_tmpdir(request):
    return get_tmpdir(request)


@pytest.fixture(scope="session")
def go_prompt_app(session_tmpdir):
    go_prompt_path = os.path.relpath(
        os.path.join(os.path.dirname(__file__), ".."))
    prompt_app_path = os.path.join(session_tmpdir, "prompt_app")

    build_env = os.environ.copy()
    build_env["GO_RPOMPT_BUILD_PATH"] = prompt_app_path

    process = subprocess.run(["make", "build_app"], cwd=go_prompt_path,
                             env=build_env)
    assert process.returncode == 0, "Failed to build go prompt app"

    return prompt_app_path


@pytest.fixture(scope="session")
def notification_server(request):
    socket_uri = os.path.join(tempfile.gettempdir(), uuid.uuid4().hex)
    server_socket = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    server_socket.bind(socket_uri)
    server_socket.listen()
    request.addfinalizer(server_socket.close)
    return socket_uri, server_socket


@pytest.fixture(scope="function")
def prompt(request, go_prompt_app, notification_server):
    params = {}
    if hasattr(request, "param"):
        params = request.param
    if "args" not in params:
        params["args"] = [""]

    socket_uri, server_socket = notification_server
    params["server_socket"] = server_socket
    params["server_socket_uri"] = socket_uri
    params["args"].append(socket_uri)

    prompt_app = PromptApp()
    prompt_app.setup(go_prompt_app, params)
    request.addfinalizer(prompt_app.close)

    return prompt_app
