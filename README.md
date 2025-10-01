# ImgAPI PoC

A minimal image storage and processing microservice in Go. It supports upload and retrieval with optional format conversion (PNG/JPEG).

## Project Layout

- `cmd/imgapi`: service entrypoint
- `internal/config`: configuration loading (env)
- `internal/logging`: logger wrapper
- `internal/storage`: filesystem storage backend
- `internal/processing`: format detection and transcoding
- `internal/service`: app service wiring storage + processing
- `internal/httpapi`: HTTP server, routes, handlers
- `pkg/api`: public API types (JSON envelopes)

## Run

```bash
IMGAPI_ADDR=:8080 IMGAPI_DATA_DIR=./data go run ./cmd/imgapi
```

- Health: `GET /healthz`
- Upload raw (octet-stream):

## Quick Test
1. Store a file in repo
```bash
curl -sS -F "file=@parrot.png" http://localhost:8080/images
```

2. Retrieve using ID with some special effects (gray scale)
```bash
curl -v "http://localhost:8080/images/<ID>.png?gray=1" -o gray.png
```

## Other commands
```bash
curl -sS -X POST \
  -H 'Content-Type: application/octet-stream' \
  -H 'X-Filename: sample.jpg' \
  --data-binary @sample.jpg \
  http://localhost:8080/images
```

- Upload multipart:

```bash
curl -sS -F "file=@parrot.jpg".jpg" http://localhost:8080/images
```

Response:

```json
{"id":"<image-id>"}
```

- Get original (auto Content-Type):

```bash
curl -v http://localhost:8080/images/<image-id> -o out
```

- Get with extension (transcode):

```bash
curl -v http://localhost:8080/images/<image-id>.jpg -o out.jpg
```

- Get with Accept negotiation:

```bash
curl -v -H 'Accept: image/jpeg' http://localhost:8080/images/<image-id> -o out.jpg
```

## Processing options

The GET endpoint supports basic processing via query parameters. You can combine these with extension-based output or Accept negotiation.

- `w`, `h`: resize width/height in pixels. If one is 0 or omitted, it will be used as-is.
- `thumb` (or `thumbnail`): if truthy and both `w` and `h` are provided, performs a center-crop thumbnail at the target size.
- `gray` (or `grayscale`): converts image to grayscale.
- `quality`: JPEG quality 1-100 (applies when output is JPEG).

Examples (assume you already have `ID` from upload):

```bash
# 1) Grayscale PNG, keep original format (no extension)
curl -v "http://localhost:8080/images/$ID?gray=1" -o gray.png

# 2) JPEG conversion with quality 60
curl -v "http://localhost:8080/images/$ID.jpg?quality=60" -o q60.jpg

# 3) Resize to width 400, auto height, keep format
curl -v "http://localhost:8080/images/$ID?w=400" -o w400

# 4) Thumbnail 160x160 JPEG, grayscale
curl -v "http://localhost:8080/images/$ID.jpg?w=160&h=160&thumb=1&gray=1" -o thumb160.jpg

# 5) Accept negotiation to JPEG with resize
curl -v -H 'Accept: image/jpeg' "http://localhost:8080/images/$ID?w=800" -o 800.jpg
```

Notes:
- If no target format is specified (no extension and no Accept), the original format is preserved when possible.
- When producing JPEG, `quality` defaults to 85 if not provided.

## Test

```bash
go test ./...
```

## Notes

- PoC supports PNG and JPEG. Add more encoders by implementing in `internal/processing`.
- Local filesystem storage for simplicity; swap out by implementing `storage.Store`.
