package apiclient

type HttpClientServiceVerb string

const (
	HttpClientServiceVerbGet     HttpClientServiceVerb = "GET"
	HttpClientServiceVerbPost    HttpClientServiceVerb = "POST"
	HttpClientServiceVerbPut     HttpClientServiceVerb = "PUT"
	HttpClientServiceVerbDelete  HttpClientServiceVerb = "DELETE"
	HttpClientServiceVerbPatch   HttpClientServiceVerb = "PATCH"
	HttpClientServiceVerbHead    HttpClientServiceVerb = "HEAD"
	HttpClientServiceVerbOptions HttpClientServiceVerb = "OPTIONS"
)

func (v HttpClientServiceVerb) String() string {
	return string(v)
}
