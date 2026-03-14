# ASCII Art Web

## Description

A web application that converts text into ASCII art using different banner styles. Built with Go's standard library — no external dependencies.

- Convert any printable ASCII text into ASCII art
- Three banner styles: **Standard**, **Shadow**, and **Thinkertoy**
- Multiline text support (newlines are preserved)
- Proper HTTP error handling (400 Bad Request, 404 Not Found, 405 Method Not Allowed, 500 Internal Server Error)

## Authors

- nikolaos-makridis

## Project Structure

```
ascii-art-web/
├── main.go              # HTTP server, route handlers, ASCII art logic
├── main_test.go         # Automated tests
├── go.mod               # Go module definition
├── banners/
│   ├── standard.txt     # Standard banner font
│   ├── shadow.txt       # Shadow banner font
│   └── thinkertoy.txt   # Thinkertoy banner font
└── templates/           # HTML templates (must be in the project root)
    ├── index.html       # Main page template
    └── error.html       # Error page template
```

> **Note:** The `templates/` directory must exist in the project root alongside `main.go`. The server loads templates from this path at startup.

## Usage: how to run

### Requirements

- Go 1.18+

### Start the server

```bash
git clone <repo-url>
cd ascii-art-web
go run main.go
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

### Using the web interface

1. Enter any printable ASCII text in the textarea
2. Select a banner style using the radio buttons: `standard`, `shadow`, or `thinkertoy`
3. Click **Generate** — the page sends a JSON request to the API and displays the result without reloading

### Using curl

```bash
# Successful API request
curl -X POST http://localhost:8080/api/ascii-art \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello","banner":"standard"}'

# 400 Bad Request — invalid banner
curl -X POST http://localhost:8080/api/ascii-art \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello","banner":"invalid"}'

# 400 Bad Request — non-ASCII character
curl -X POST http://localhost:8080/api/ascii-art \
  -H "Content-Type: application/json" \
  -d '{"text":"Héllo","banner":"standard"}'

# 405 Method Not Allowed
curl -X GET http://localhost:8080/api/ascii-art

# 404 Not Found
curl http://localhost:8080/unknown
```

### Running the tests

```bash
# Run all tests
go test ./...

# Run with verbose output (shows each test name and result)
go test -v ./...
```

## Implementation details: algorithm

Each banner file (`banners/standard.txt`, `banners/shadow.txt`, `banners/thinkertoy.txt`) stores the ASCII art for all printable characters (codes 32–126). Each character is represented as **8 rows** of text, followed by a blank separator line, giving a block size of 9 lines per character.

**Loading a banner:**
1. Read all lines from the banner file into a slice.
2. For each character code `c` in range 32–126, compute `start = (c - 32) * 9`.
3. Map the rune to `lines[start : start+8]` (the 8 art rows).

**Generating ASCII art:**
1. Split the input text on `\n` to get lines.
2. For each non-empty line, iterate over 8 row indices (0–7).
3. For each row, concatenate the corresponding art row from every character in the line, then append a newline.
4. Empty input lines produce a single newline in the output, preserving blank lines.

**HTTP endpoints:**

| Method | Path              | Description                                      |
|--------|-------------------|--------------------------------------------------|
| GET    | `/`               | Renders the main page (200 OK)                   |
| POST   | `/ascii-art`      | Generates ASCII art from form input (200 OK)     |
| POST   | `/api/ascii-art`  | JSON API: accepts `{"text","banner"}`, returns `{"result"}` (200 OK) |

**Input validation:**
- Only printable ASCII characters (32–126) and `\n` are accepted; anything else returns `400 Bad Request`.
- Banner must be one of `standard`, `shadow`, `thinkertoy`; otherwise returns `400 Bad Request`.
- Unknown routes return `404 Not Found`.
- Wrong HTTP method returns `405 Method Not Allowed`.
