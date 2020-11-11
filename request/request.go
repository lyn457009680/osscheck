package request

type Request struct {
	Url        string
	DeviceType string
	ParserFunc func([]byte, string) ParseResult
}

type ParseResult struct {
	Requests []Request
	Items    []interface{}
}
