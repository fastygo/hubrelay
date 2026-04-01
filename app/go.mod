module hubrelay-dashboard

go 1.25.5

require (
	github.com/a-h/templ v0.3.1001
	sshbot v0.0.0
)

require github.com/fastygo/ui8kit v0.1.1

replace github.com/fastygo/ui8kit => ./ui8kit

replace sshbot => ../
