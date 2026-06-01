package openai

// chat API types (OpenAI-compatible /v1/chat/completions).

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Error   *apiError    `json:"error,omitempty"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}
