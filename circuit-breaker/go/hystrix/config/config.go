package config

var (
	// DefaultTimeoutMillis is how long to wait for command to complete
	DefaultTimeoutMillis = 1000
	// DefaultMinRequestNum is the minimum number of requests needed before a circuit can be tripped due to health
	DefaultMinRequestNum = 20
	// DefaultBackoffMillis is how long to wait after a circuit opens before transiting to half-open state.
	DefaultBackoffMillis = 5000
	// DefaultErrorPercentThreshold suggests open the circuit if the error percent surpasses it.
	DefaultErrorPercentThreshold = 50
)
