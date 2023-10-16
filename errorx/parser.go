package errorx

type Parser interface {
	Parse(error) ErrorX
	Support(error) bool
}

var Parsers = &parserHolder{parsers: make(map[string]Parser)}

type parserHolder struct {
	parsers map[string]Parser
}

func (h *parserHolder) Add(name string, parser Parser) {
	h.parsers[name] = parser
}

func (h *parserHolder) Remove(name string) {
	delete(h.parsers, name)
}

func (h *parserHolder) Iter() map[string]Parser {
	return h.parsers
}
