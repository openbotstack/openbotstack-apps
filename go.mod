module github.com/openbotstack/openbotstack-apps

go 1.26.1

replace github.com/openbotstack/openbotstack-core => ../openbotstack-core

require (
	github.com/openbotstack/openbotstack-core v0.0.0-00010101000000-000000000000
	github.com/openbotstack/openbotstack-runtime v1.0.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.8.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/openbotstack/openbotstack-runtime => ../openbotstack-runtime
