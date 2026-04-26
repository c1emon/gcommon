module github.com/c1emon/gcommon/httpx

go 1.25.0

require (
	github.com/c1emon/gcommon v0.0.0
	github.com/imroc/req/v3 v3.57.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/icholy/digest v1.1.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.57.1 // indirect
	github.com/refraction-networking/utls v1.8.1 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/time v0.15.0
)

replace (
	github.com/c1emon/gcommon => ../
	// req/v3 pins quic-go in its go.mod; MVS can select a newer quic-go that breaks req's http3 build.
	github.com/quic-go/quic-go => github.com/quic-go/quic-go v0.57.1
)
