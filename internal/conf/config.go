package conf

import "time"

type Config struct {
	NbrOfRobots            int           `env:"NBR_OF_ROBOTS,required=true"`
	Secret                 string        `env:"SECRET,required=true"`
	OutputFile             string        `env:"OUTPUT_FILE,required=true"`
	BufferSize             int           `env:"BUFFER_SIZE,required=true"`
	EndOfSecret            string        `env:"END_OF_SECRET,required=true"`
	PercentageOfLost       int           `env:"PERCENTAGE_OF_LOST,required=true"`
	PercentageOfDuplicated int           `env:"PERCENTAGE_OF_DUPLICATED,required=true"`
	DuplicatedNumber       int           `env:"DUPLICATED_NUMBER,required=true"`
	MaxAttempts            int           `env:"MAX_ATTEMPTS,required=true"`
	Timeout                time.Duration `env:"TIMEOUT,required=true"`
	QuietPeriod            time.Duration `env:"QUIET_PERIOD,required=true"`
	GossipTime             time.Duration `env:"GOSSIP_TIME,required=true"`
	MetricInterval         time.Duration `env:"METRIC_INTERVAL,required=true"`
	ObservabilityInterval  time.Duration `env:"OBSERVABILITY_INTERVAL,required=true"`
	LowCapacityThreshold   int           `env:"LOW_CAPACITY_THRESHOLD,required=true"`
	LogLevel               string        `env:"LOG_LEVEL,default=INFO"`
}
