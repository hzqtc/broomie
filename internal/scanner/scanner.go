package scanner

type Reason string

const (
	StaleCache Reason = "Stale Cache"
)

type ScannerResult struct {
	path         string
	sizeBytes    int64
	creationDate int64
	modifiedDate int64
	reason       Reason
}

type Scanner interface {
	Scan() []ScannerResult
}
