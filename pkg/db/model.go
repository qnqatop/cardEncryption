package db

type EncryptCard struct {
	EcnCard string
	Token   string
	Dec     string
}

type Dek struct {
	Key   string
	Count int
}
