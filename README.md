# Vizualizer
## Install
1. Install Golang
2. `go get github.com/malyutinegor/vizualizer`

## Run
1. `cd $GOPATH/src/github.com/malyutinegor/vizualizer`
2. `go run main.go`

# Perfomance tips
## PostgreSQL
For better perfomance with PostrgeSQL, you should set timestamptz as primary key (or create index on it). Example of PostgreSQL database setup (you can change these columns for your needs):

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

### Cache tables
For the best perfomance, you should create the cache tables. They can minimize calculations when user is requesting average rows.

#### Creating cache tables
These commands will create cache tables, which copies the entire structure of main table and store average data for larger time intervals (day e.g.).

```sql
-- Create cache table, which structure is exactly same as main table structure
create table records_minutes (
	time timestamptz primary key,
	voltage numeric not null,
	amperage numeric not null,
	power numeric not null,
	energy_supplied numeric not null,
	energy_received numeric not null
)

-- Fill it with data from main table
insert into records_minutes (
	select
	date_trunc('minute', time) as time,
	avg(voltage) as voltage,
	avg(amperage) as amperage,
	avg(power) as power,
	avg(energy_supplied) as energy_supplied,
	avg(energy_received) as energy_received
	from records
	group by date_trunc('minute', time)
);

create table records_hours (
	time timestamptz primary key,
	voltage numeric not null,
	amperage numeric not null,
	power numeric not null,
	energy_supplied numeric not null,
	energy_received numeric not null
)

insert into records_hours (
	select
	date_trunc('hour', time) as time,
	avg(voltage) as voltage,
	avg(amperage) as amperage,
	avg(power) as power,
	avg(energy_supplied) as energy_supplied,
	avg(energy_received) as energy_received
	from records_minutes
	group by date_trunc('hour', time)
);

create table records_days (
	time timestamptz primary key,
	voltage numeric not null,
	amperage numeric not null,
	power numeric not null,
	energy_supplied numeric not null,
	energy_received numeric not null
)

insert into records_days (
	select
	date_trunc('day', time) as time,
	avg(voltage) as voltage,
	avg(amperage) as amperage,
	avg(power) as power,
	avg(energy_supplied) as energy_supplied,
	avg(energy_received) as energy_received
	from records_hours
	group by date_trunc('day', time)
);
```

#### Setting auto cache rules for cache tables
The cache tables also must be updated, when new data appears in the main table. There's an PostgreSQL rule, which automatically updates all your cache tables when the new data is added in new table (note: don't forget to change this query structure for your database setup):

```sql
create or replace rule cache_records as on insert to records
do also (
	insert into records_minutes (time, voltage, amperage, power, energy_supplied, energy_received)
			(select date_trunc('minute', NEW.time) as time, * from
				(select
			 	 avg(voltage) as voltage,
				 avg(amperage) as amperage,
				 avg(power) as power,
				 avg(energy_supplied) as energy_supplied,
				 avg(energy_received) as energy_received
				 from records
				 where time >= date_trunc('minute', NEW.time)
	  			   and time <= date_trunc('minute', interval '1 min' + NEW.time)
			    ) tbl1
			)
		on conflict (time)
		do update set
			voltage = EXCLUDED.voltage,
			amperage = EXCLUDED.amperage,
			power = EXCLUDED.power,
			energy_supplied = EXCLUDED.energy_supplied,
			energy_received = EXCLUDED.energy_received;
	
	insert into records_hours (time, voltage, amperage, power, energy_supplied, energy_received)
			(select date_trunc('hour', NEW.time) as time, * from
				(select
			 	 avg(voltage) as voltage,
				 avg(amperage) as amperage,
				 avg(power) as power,
				 avg(energy_supplied) as energy_supplied,
				 avg(energy_received) as energy_received
				 from records_minutes
				 where time >= date_trunc('hour', NEW.time)
	  			   and time <= date_trunc('hour', interval '1 hour' + NEW.time)
			    ) tbl1
			)
		on conflict (time)
		do update set
			voltage = EXCLUDED.voltage,
			amperage = EXCLUDED.amperage,
			power = EXCLUDED.power,
			energy_supplied = EXCLUDED.energy_supplied,
			energy_received = EXCLUDED.energy_received;
	
	insert into records_days (time, voltage, amperage, power, energy_supplied, energy_received)
			(select date_trunc('day', NEW.time) as time, * from
				(select
			 	 avg(voltage) as voltage,
				 avg(amperage) as amperage,
				 avg(power) as power,
				 avg(energy_supplied) as energy_supplied,
				 avg(energy_received) as energy_received
				 from records_hours
				 where time >= date_trunc('day', NEW.time)
	  			   and time <= date_trunc('day', interval '1 day' + NEW.time)
			    ) tbl1
			)
		on conflict (time)
		do update set
			voltage = EXCLUDED.voltage,
			amperage = EXCLUDED.amperage,
			power = EXCLUDED.power,
			energy_supplied = EXCLUDED.energy_supplied,
			energy_received = EXCLUDED.energy_received;
);
```

## Configuration
This vizualizer uses configuration file named `config.toml` in your current directory. Here's all available parameters:

```toml
# Name of database adapter, which is used to get data from database. At this moment, the only available adapter is "postgresql".
db = "postgresql"

# Settings for PostgreSQL driver.
[postgresql]
# URL of PostgreSQL database with passwords and other things in it.
url = "..."
# Name of main tables (example).
table = "records"
# Names of cache tables (example).
days-table = "records_days"
hours-table = "records_hours"
minutes-table = "records_minutes"
# Sequence of columns in PostgreSQL database (example).
format = ["date", "number", "number", "number", "number", "number"]
```

<!-- TODO: auto create cache tables -->
