package gateway

type tuiRuntimeRequest struct {
	Request
	compactHistory []Message
	compactSystem  string
	seedHistory    []Message
	replayKey      string
	resumedCopy    bool                  // resumes a session copied for this turn alone
	pointResume    bool                  // resumes a cut point's session in place
	workKey        string                // state key of the cwd of the session it resumes
	heldTools      []byte                // tool definitions kept for the cache, not offered (heldTools)
	toolRepair     *tuiRuntimeToolRepair // resumes its corrected tool round with --reply-on-resume, no prompt
	tuiSearchTool  string
}

func tuiRuntimeParseRequest(body []byte) (*tuiRuntimeRequest, error) {
	request, err := parseAnthropic(body)
	if err != nil {
		return nil, err
	}
	return &tuiRuntimeRequest{Request: *request}, nil
}
