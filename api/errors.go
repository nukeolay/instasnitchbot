package api

// 	---------EndpointErrorHttp---------
type EndpointErrorHttp struct {
	s string
}

func (err EndpointErrorHttp) Error() string {
	return err.s
}

func NewEndpointErrorHttp(text string) error {
	return EndpointErrorHttp{text}
}

// 	---------EndpointErrorParsing--------- возникает когда от сервера приходит ответ не в json (скорее всего заблокировали)
type EndpointErrorParsing struct {
	s string
}

func (err EndpointErrorParsing) Error() string {
	return err.s
}
func NewEndpointErrorParsing(text string) error {
	return EndpointErrorParsing{text}
}

// 	---------EndpointErrorAccountNotFound---------
type EndpointErrorAccountNotFound struct {
	s string
}

func (err EndpointErrorAccountNotFound) Error() string {
	return err.s
}
func NewEndpointErrorAccountNotFound(text string) error {
	return EndpointErrorAccountNotFound{text}
}
