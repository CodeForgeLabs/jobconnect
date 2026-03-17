module jobconnect/recommendation

go 1.26.1

require (
	google.golang.org/grpc v1.79.2
	google.golang.org/protobuf v1.36.11
	jobconnect/job v0.0.0-00010101000000-000000000000
	jobconnect/user v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)

replace jobconnect/job => ../job

replace jobconnect/user => ../user
