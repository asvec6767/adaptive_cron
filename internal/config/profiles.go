package config

type Profile struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Features    []string `yaml:"features"`
}

var AvailableProfiles = []Profile{
	{
		Name:        "lite",
		Description: "Lightweight version with SQLite and in-memory queue",
		Features: []string{
			"db.sqlite",
			"queue.memory",
			"log.file",
			"transport.inmemory",
			"metrics.basic",
		},
	},
	{
		Name:        "standard",
		Description: "Standard version with PostgreSQL and Redis",
		Features: []string{
			"db.postgres",
			"queue.redis",
			"log.file",
			"log.json",
			"transport.rest",
			"transport.grpc",
			"metrics.prometheus",
		},
	},
	{
		Name:        "enterprise",
		Description: "Enterprise version with ClickHouse and NATS",
		Features: []string{
			"db.postgres",
			"db.replication",
			"queue.redis",
			"queue.persistent",
			"log.clickhouse",
			"log.structured",
			"transport.grpc",
			"transport.nats",
			"metrics.prometheus",
			"service.discovery",
			"scheduler.distributed",
		},
	},
}

func GetProfile(name string) *Profile {
	for _, p := range AvailableProfiles {
		if p.Name == name {
			return &p
		}
	}
	return nil
}
