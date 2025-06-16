package entities

type SymbolPair struct {
	Id          int
	BaseSymbol  Symbol
	QuoteSymbol Symbol
	Market      Market
}
