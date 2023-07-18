import pytest

DEFAULT_KEYS = {
    "Left": "Left",
    "Right": "Right",
    "Erase": "BSpace",
    "Enter": "Enter",
    "Up": "Up",
    "Down": "Down",
    "C-r": "C-r",
    "Home": "Home",
    "End": "End",
    "M-b": "M-b",
    "M-f": "M-f",
}

EMACS_KEYS = {
    "Left": "C-b",
    "Right": "C-f",
    "Erase": "C-h",
    "Enter": "C-m",
    "Up": "C-p",
    "Down": "C-n",
    "C-r": "C-r",
    "Home": "c-a",
    "End": "c-e",
    "M-b": "M-b",
    "M-f": "M-f",
}


def test_launch(prompt):
    assert prompt.get_cursor() == (12, 0)
    assert prompt.dump_workspace() == "prompt_app>"


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
def test_input_text(prompt, keys):
    prompt.send_keys("cmd")
    assert (15, 0) == prompt.get_cursor()

    prompt.send_keys(keys["Enter"])
    assert (12, 2) == prompt.get_cursor()

    prompt.send_keys("русский язык")
    assert (24, 2) == prompt.get_cursor()

    expected = \
        """prompt_app> cmd
cmd: cmd
prompt_app> русский язык"""
    assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
def test_remove_text(prompt, keys):
    prompt.send_keys("##### e хай hello")
    prompt.send_keys([keys["Left"]] * 9)
    prompt.send_keys([keys["Erase"]] * 7)
    assert prompt.get_cursor() == (13, 0)

    expected = "prompt_app> #хай hello"
    assert prompt.dump_workspace() == expected

    prompt.send_keys([keys["Erase"]] * 2)
    assert prompt.get_cursor() == (12, 0)
    expected2 = "prompt_app> хай hello"
    assert prompt.dump_workspace() == expected2


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
def test_move_over_input(prompt, keys):
    prompt.send_keys("hello!")
    prompt.send_keys([keys["Left"]] * 2)
    assert prompt.get_cursor() == (16, 0)

    prompt.send_keys([keys["Left"]] * 10)
    assert prompt.get_cursor() == (12, 0)

    prompt.send_keys([keys["Right"]] * 7)
    assert prompt.get_cursor() == (18, 0)

    prompt.send_keys("слово")
    prompt.send_keys([keys["Left"]] * 3)
    assert prompt.get_cursor() == (20, 0)

    expected = "prompt_app> hello!слово"
    assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
@pytest.mark.parametrize("prompt", [{"x": "100"}], indirect=True)
def test_multiline_commands(prompt, keys):
    prompt.send_keys("строка1\nline2\nline3a")
    assert prompt.get_cursor() == (6, 2)

    prompt.send_keys([keys["Left"]] * 7)
    assert prompt.get_cursor() == (5, 1)

    prompt.send_keys([keys["Left"]] * 4)
    prompt.send_keys([keys["Erase"]] * 2)
    assert prompt.get_cursor() == (19, 0)

    expected = """prompt_app> строка1ine2
line3a"""
    assert prompt.dump_workspace() == expected

    prompt.send_keys("абвгдеёжзийклмнопрст")
    assert prompt.get_cursor() == (39, 0)

    expected = """prompt_app> строка1абвгдеёжзийклмнопрстine2
line3a"""
    assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
def test_enter(prompt, keys):
    pipeline = [
        "one line cmd",
        keys["Enter"],
        "строка1\nстрока2aa",
        keys["Enter"],
        "бессовестно\nмного\nстрокЖ",
        keys["Enter"]
    ]
    for cmd in pipeline:
        prompt.send_keys(cmd)
    assert prompt.get_cursor() == (12, 12)

    expected = """prompt_app> one line cmd
cmd: one line cmd
prompt_app> строка1
строка2aa
cmd: строка1
строка2aa
prompt_app> бессовестно
много
строкЖ
cmd: бессовестно
много
строкЖ
prompt_app>"""
    assert prompt.dump_workspace() == expected


# Don't match with completion, using in app.
HISTORY = [
    "print(C)",
    "print(D)",
    "yuk#",
    "command",
    "команда1\nкоманда2\nкоманда3",
    "a\n\nb",
    "if a then\n print(x)\nelse\nprint(y)",
    "interpeter\nдстрокаслово\nfdlqfdsl_fsldgg\nrpewr",
]

HISTORY_ARG = ";".join(HISTORY)


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
@pytest.mark.parametrize("prompt", [{"args": [HISTORY_ARG]}], indirect=True)
def test_move_history(prompt, keys):
    history_rev = HISTORY.copy()
    history_rev.reverse()

    # Up.
    for entry in history_rev:
        prompt.send_keys(keys["Up"])
        expected = "prompt_app> " + entry
        assert prompt.dump_workspace() == expected

    # Down.
    for i in range(1, len(HISTORY)):
        prompt.send_keys(keys["Down"])
        expected = "prompt_app> " + HISTORY[i]
        assert prompt.dump_workspace() == expected

    prompt.send_keys("Down")
    prompt.send_keys("cmd to\nshift cursor")

    # Up.
    for entry in history_rev:
        prompt.send_keys(keys["Up"])
        expected = "prompt_app> " + entry
        assert prompt.dump_workspace() == expected

    # Down.
    for i in range(1, len(HISTORY)):
        prompt.send_keys(keys["Down"])
        expected = "prompt_app> " + HISTORY[i]
        assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
@pytest.mark.parametrize("prompt", [{"args": [HISTORY_ARG]}], indirect=True)
def test_enter_history(prompt, keys):
    prompt.send_keys([keys["Up"], keys["Up"], keys["Enter"]])

    expected = """prompt_app> if a then
 print(x)
else
print(y)
cmd: if a then
 print(x)
else
print(y)
prompt_app>"""
    assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
@pytest.mark.parametrize("prompt", [{"args": [HISTORY_ARG]}], indirect=True)
def test_reverse_search(prompt, keys):
    prompt.send_keys(["Some-текст", keys["C-r"]])
    expected = """(reverse-i-search)`':Some-текст"""
    assert prompt.dump_workspace() == expected

    prompt.send_keys(["print("])
    expected = """(reverse-i-search)`print(':if a then
 print(x)
else
print(y)"""
    assert prompt.dump_workspace() == expected

    prompt.send_keys(keys["C-r"])
    expected = """(reverse-i-search)`print(':print(D)"""
    assert prompt.dump_workspace() == expected

    prompt.send_keys(keys["C-r"])
    expected = """(reverse-i-search)`print(':print(C)"""
    assert prompt.dump_workspace() == expected

    prompt.send_keys(keys["Left"])
    expected = """prompt_app> print(C)"""
    assert prompt.dump_workspace() == expected

    # Check that we are at the past.
    for i in range(1, len(HISTORY)):
        prompt.send_keys(keys["Down"])
        expected = "prompt_app> " + HISTORY[i]
        assert prompt.dump_workspace() == expected

    prompt.send_keys([keys["C-r"], "not matched with any"])
    prompt.send_keys([keys["C-r"], keys["C-r"], keys["C-r"]])
    expected = """(failed reverse-i-search)`not matched with any':"""
    assert prompt.dump_workspace() == expected

    prompt.send_keys(keys["Up"])
    expected = """prompt_app>"""
    assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
@pytest.mark.parametrize("prompt", [{"args": [HISTORY_ARG]}], indirect=True)
def test_enter_reverse_search(prompt, keys):
    prompt.send_keys(["C-r", "print(", keys["Enter"]])
    expected = """prompt_app> if a then
 print(x)
else
print(y)
cmd: if a then
 print(x)
else
print(y)
prompt_app>"""
    assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
@pytest.mark.parametrize(
    "prompt",
    [{"args": ["if 1\tprint(2)else\tprint(3)"]}],
    indirect=True
)
def test_tabs(prompt, keys):
    prompt.send_keys("Hello,\tпривет,\ttabs")
    assert prompt.get_cursor() == (37, 0)
    assert prompt.dump_workspace() == "prompt_app> Hello,    привет,    tabs"

    prompt.send_keys(keys["Up"])
    expected = "prompt_app> if 1    print(2)else    print(3)"
    assert prompt.dump_workspace() == expected


def test_console_not_broken(prompt):
    prompt.send_keys(["exit", "text"])
    expected = """prompt_app> exittext"""
    assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
def test_home_end_keys(prompt, keys):
    cmd = """здравствуй,
nebo
в облаках"""
    prompt.send_keys(cmd)
    assert prompt.get_cursor() == (9, 2)

    prompt.send_keys(keys["Home"])
    assert prompt.get_cursor() == (12, 0)

    prompt.send_keys(["Right"] * 7 + [keys["Home"]])
    assert prompt.get_cursor() == (12, 0)

    prompt.send_keys(keys["End"])
    assert prompt.get_cursor() == (9, 2)

    prompt.send_keys(["Left"] * 9 + [keys["End"]])
    assert prompt.get_cursor() == (9, 2)

    prompt.send_keys(["\n-текст"])
    expected = """prompt_app> здравствуй,
nebo
в облаках
-текст"""
    assert prompt.dump_workspace() == expected


@pytest.mark.parametrize("keys", [DEFAULT_KEYS, EMACS_KEYS])
def test_go_left_right_word(prompt, keys):
    cmd = """a b c d
слово1 слово2 слово3   слово4
d


б"""
    # Go left from the end.
    cmds = [
        [cmd, keys["M-b"]],
        keys["M-b"],
        keys["M-b"],
        ["Left", "Left", keys["M-b"]],
        [keys["M-b"]] * 6
    ]
    cursors = [
        (0, 5),
        (0, 2),
        (23, 1),
        (14, 1),
        (12, 0),
    ]
    for cmd, cursor in zip(cmds, cursors):
        prompt.send_keys(cmd)
        assert prompt.get_cursor() == cursor

    # Go right from the beginning.
    cmds = [
        keys["M-f"],
        keys["M-f"],
        [keys["M-f"]] * 5,
        ["Right", "Right", keys["M-f"]],
        [keys["M-f"]] * 2
    ]
    cursors = [
        (13, 0),
        (15, 0),
        (20, 1),
        (29, 1),
        (1, 5),
    ]

    for cmd, cursor in zip(cmds, cursors):
        prompt.send_keys(cmd)
        assert prompt.get_cursor() == cursor
