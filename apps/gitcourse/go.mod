module gitcourse

go 1.25.5

require (
	github.com/a-h/templ v0.3.1001
	github.com/fastygo/hubcore v0.0.0
	github.com/fastygo/hubrelay-sdk v0.0.0
	github.com/fastygo/ui8kit v0.1.1
)

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/grpc v1.80.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/fastygo/ui8kit => ../dashboard/ui8kit

replace github.com/fastygo/hubcore => ../../hubcore

replace github.com/fastygo/hubrelay-sdk => ../../sdk/hubrelay
