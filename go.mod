module github.com/joncaudill/gator

go 1.23.4

require (
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	internal/config v0.0.0-20220103123456-123456789012
)

replace internal/config => ./internal/config
