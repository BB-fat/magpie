"""Remove local environment identifiers from public TLC evidence."""
import re


def redact_tlc_output(output: str) -> str:
    lines = []
    for line in output.splitlines(keepends=True):
        ending = "\n" if line.endswith("\n") else ""
        text = line.rstrip("\r\n")
        if text.startswith("Running "):
            mode = re.match(r"(.*?\bwith \d+ workers?)\b", text)
            if mode:
                text = mode.group(1) + ". [machine details redacted]"
        elif text.startswith("Parsing file "):
            path = text.removeprefix("Parsing file ").replace("\\", "/")
            if path.startswith("/") or re.match(r"^[A-Za-z]:/", path):
                marker = "/formal/claude-tui/"
                public = "formal/claude-tui/" + path.split(marker, 1)[1] if marker in path else path.rsplit("/", 1)[-1]
                text = "Parsing file " + public
        lines.append(text + ending)
    return "".join(lines)
