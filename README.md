# Vizualizer

## Install

1. Install Golang
2. `go get github.com/malyutinegor/vizualizer`

## Run

1. `cd $GOPATH/src/github.com/malyutinegor/vizualizer`
2. `go run main.go`

# Perfomance tips

## PostgreSQL

For better perfomance with PostrgeSQL, you should set timestamptz as primary key (or create index on it). Example of PostgreSQL database setup:

```sql
create table records (
	time timestamptz primary key,
	voltage numeric not null,
	amperage numeric not null,
	power numeric not null,
	energy_supplied numeric not null,
	energy_received numeric not null
)
```

/* TODO: add sslmode=disable */
/* TODO: add cache tables tutorial */
