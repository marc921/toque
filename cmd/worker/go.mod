module marcbrun.io/kubernetes-api/worker

go 1.22.0

require (
	marcbrun.io/kubernetes-api/db/sqlcgen v0.0.0-00010101000000-000000000000
	marcbrun.io/kubernetes-api/pkg v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.3 // indirect
	github.com/rabbitmq/amqp091-go v1.9.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	marcbrun.io/kubernetes-api/db/sqlcgen => ../../db/sqlcgen
	marcbrun.io/kubernetes-api/pkg => ../../pkg
)
