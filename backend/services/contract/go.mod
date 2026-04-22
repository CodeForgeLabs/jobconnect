module jobconnect/contract

go 1.25.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
	jobconnect/job v0.0.0
	jobconnect/proposal v0.0.0
	jobconnect/user v0.0.0
)

replace jobconnect/job => ../job
replace jobconnect/proposal => ../proposal
replace jobconnect/user => ../user

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)
