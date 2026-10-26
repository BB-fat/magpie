package gateway

type tuiRuntimeRequest struct {
	Request
	compactHistory []Message
	compactSystem  string
	seedHistory    []Message
	replayKey      string
	resumedCopy    bool
	tuiSearchTool  string
}

func tuiRuntimeParseRequest(body []byte) (*tuiRuntimeRequest, error) {
	request, err := parseAnthropic(body)
	if err != nil {
		return nil, err
	}
	return &tuiRuntimeRequest{Request: *request}, nil
}
