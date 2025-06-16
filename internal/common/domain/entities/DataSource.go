package entities

type DataSource struct {
	Id               int
	Name             string
	ConnectionString string
	Login            string
	Password         string
	RateLimit        int
	SymbolPairs      []SymbolPair
}
