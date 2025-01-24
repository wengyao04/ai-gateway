// Package openai contains the following is the OpenAI API schema definitions.
// Note that we intentionally do not use the code generation tools like OpenAPI Generator not only to keep the code simple
// but also because the OpenAI's OpenAPI definition is not compliant with the spec and the existing tools do not work well.
package openai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Chat message role defined by the OpenAI API.
const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleDeveloper = "developer"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleFunction  = "function"
	ChatMessageRoleTool      = "tool"
)

// ChatCompletionContentPartRefusalType The type of the content part.
type ChatCompletionContentPartRefusalType string

// ChatCompletionContentPartInputAudioType The type of the content part. Always `input_audio`.
type ChatCompletionContentPartInputAudioType string

// ChatCompletionContentPartTextType The type of the content part.
type ChatCompletionContentPartTextType string

// ChatCompletionContentPartImageType The type of the content part.
type ChatCompletionContentPartImageType string

const (
	ChatCompletionContentPartTextTypeText             ChatCompletionContentPartTextType       = "text"
	ChatCompletionContentPartRefusalTypeRefusal       ChatCompletionContentPartRefusalType    = "refusal"
	ChatCompletionContentPartInputAudioTypeInputAudio ChatCompletionContentPartInputAudioType = "input_audio"
	ChatCompletionContentPartImageTypeImageURL        ChatCompletionContentPartImageType      = "image_url"
)

// ChatCompletionContentPartTextParam Learn about
// [text inputs](https://platform.openai.com/docs/guides/text-generation).
type ChatCompletionContentPartTextParam struct {
	// The text content.
	Text string `json:"text"`
	// The type of the content part.
	Type string `json:"type"`
}

type ChatCompletionContentPartRefusalParam struct {
	// The refusal message generated by the model.
	Refusal string `json:"refusal"`
	// The type of the content part.
	Type ChatCompletionContentPartRefusalType `json:"type"`
}

// ChatCompletionContentPartInputAudioParam Learn about [audio inputs](https://platform.openai.com/docs/guides/audio).
type ChatCompletionContentPartInputAudioParam struct {
	InputAudio ChatCompletionContentPartInputAudioInputAudioParam `json:"input_audio"`
	// The type of the content part. Always `input_audio`.
	Type ChatCompletionContentPartInputAudioType `json:"type"`
}

// ChatCompletionContentPartInputAudioInputAudioFormat The format of the encoded audio data. Currently supports "wav" and "mp3".
type ChatCompletionContentPartInputAudioInputAudioFormat string

const (
	ChatCompletionContentPartInputAudioInputAudioFormatWAV ChatCompletionContentPartInputAudioInputAudioFormat = "wav"
	ChatCompletionContentPartInputAudioInputAudioFormatMP3 ChatCompletionContentPartInputAudioInputAudioFormat = "mp3"
)

type ChatCompletionContentPartInputAudioInputAudioParam struct {
	// Base64 encoded audio data.
	Data string `json:"data"`
	// The format of the encoded audio data. Currently supports "wav" and "mp3".
	Format ChatCompletionContentPartInputAudioInputAudioFormat `json:"format"`
}

type ChatCompletionContentPartImageImageURLDetail string

const (
	ChatCompletionContentPartImageImageURLDetailAuto ChatCompletionContentPartImageImageURLDetail = "auto"
	ChatCompletionContentPartImageImageURLDetailLow  ChatCompletionContentPartImageImageURLDetail = "low"
	ChatCompletionContentPartImageImageURLDetailHigh ChatCompletionContentPartImageImageURLDetail = "high"
)

type ChatCompletionContentPartImageImageURLParam struct {
	// Either a URL of the image or the base64 encoded image data.
	URL string `json:"url"`
	// Specifies the detail level of the image. Learn more in the
	// [Vision guide](https://platform.openai.com/docs/guides/vision#low-or-high-fidelity-image-understanding).
	Detail ChatCompletionContentPartImageImageURLDetail `json:"detail,omitempty"`
}

// ChatCompletionContentPartImageParam Learn about [image inputs](https://platform.openai.com/docs/guides/vision).
type ChatCompletionContentPartImageParam struct {
	ImageURL ChatCompletionContentPartImageImageURLParam `json:"image_url"`
	// The type of the content part.
	Type ChatCompletionContentPartImageType `json:"type"`
}

// ChatCompletionContentPartUserUnionParam Learn about
// [text inputs](https://platform.openai.com/docs/guides/text-generation).
type ChatCompletionContentPartUserUnionParam struct {
	TextContent       *ChatCompletionContentPartTextParam
	InputAudioContent *ChatCompletionContentPartInputAudioParam
	ImageContent      *ChatCompletionContentPartImageParam
}

func (c *ChatCompletionContentPartUserUnionParam) UnmarshalJSON(data []byte) error {
	var chatContentPart map[string]interface{}
	if err := json.Unmarshal(data, &chatContentPart); err != nil {
		return err
	}
	var contentType string
	var ok bool
	if contentType, ok = chatContentPart["type"].(string); !ok {
		return fmt.Errorf("chat content does not have type")
	}
	switch contentType {
	case string(ChatCompletionContentPartTextTypeText):
		var textContent ChatCompletionContentPartTextParam
		if err := json.Unmarshal(data, &textContent); err != nil {
			return err
		}
		c.TextContent = &textContent
	case string(ChatCompletionContentPartInputAudioTypeInputAudio):
		var audioContent ChatCompletionContentPartInputAudioParam
		if err := json.Unmarshal(data, &audioContent); err != nil {
			return err
		}
		c.InputAudioContent = &audioContent
	case string(ChatCompletionContentPartImageTypeImageURL):
		var imageContent ChatCompletionContentPartImageParam
		if err := json.Unmarshal(data, &imageContent); err != nil {
			return err
		}
		c.ImageContent = &imageContent
	default:
		return fmt.Errorf("unknown ChatCompletionContentPartUnionParam type: %v", contentType)
	}
	return nil
}

type StringOrArray struct {
	Value interface{}
}

func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err == nil {
		s.Value = str
		return nil
	}

	var arr []ChatCompletionContentPartTextParam
	err = json.Unmarshal(data, &arr)
	if err == nil {
		s.Value = arr
		return nil
	}

	return fmt.Errorf("cannot unmarshal JSON data as string or array of string")
}

type StringOrUserRoleContentUnion struct {
	Value interface{}
}

func (s *StringOrUserRoleContentUnion) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err == nil {
		s.Value = str
		return nil
	}

	var arr []ChatCompletionContentPartUserUnionParam
	err = json.Unmarshal(data, &arr)
	if err == nil {
		s.Value = arr
		return nil
	}

	return fmt.Errorf("cannot unmarshal JSON data as string or array of content parts")
}

type ChatCompletionMessageParamUnion struct {
	Value interface{}
	Type  string
}

func (c *ChatCompletionMessageParamUnion) UnmarshalJSON(data []byte) error {
	var chatMessage map[string]interface{}
	if err := json.Unmarshal(data, &chatMessage); err != nil {
		return err
	}
	if _, ok := chatMessage["role"]; !ok {
		return fmt.Errorf("chat message does not have role")
	}
	var role string
	var ok bool
	if role, ok = chatMessage["role"].(string); !ok {
		return fmt.Errorf("chat message role is not string: %s", role)
	}
	switch role {
	case ChatMessageRoleUser:
		var userMessage ChatCompletionUserMessageParam
		if err := json.Unmarshal(data, &userMessage); err != nil {
			return err
		}
		c.Value = userMessage
		c.Type = ChatMessageRoleUser
	case ChatMessageRoleAssistant:
		var assistantMessage ChatCompletionAssistantMessageParam
		if err := json.Unmarshal(data, &assistantMessage); err != nil {
			return err
		}
		c.Value = assistantMessage
		c.Type = ChatMessageRoleAssistant
	case ChatMessageRoleSystem:
		var systemMessage ChatCompletionSystemMessageParam
		if err := json.Unmarshal(data, &systemMessage); err != nil {
			return err
		}
		c.Value = systemMessage
		c.Type = ChatMessageRoleSystem
	case ChatMessageRoleDeveloper:
		var developerMessage ChatCompletionDeveloperMessageParam
		if err := json.Unmarshal(data, &developerMessage); err != nil {
			return err
		}
		c.Value = developerMessage
		c.Type = ChatMessageRoleDeveloper
	case ChatMessageRoleTool:
		var toolMessage ChatCompletionToolMessageParam
		if err := json.Unmarshal(data, &toolMessage); err != nil {
			return err
		}
		c.Value = toolMessage
		c.Type = ChatMessageRoleTool
	default:
		return fmt.Errorf("unknown ChatCompletionMessageParam type: %v", role)
	}
	return nil
}

// ChatCompletionUserMessageParam Messages sent by an end user, containing prompts or additional context
// information.
type ChatCompletionUserMessageParam struct {
	// The contents of the user message.
	Content StringOrUserRoleContentUnion `json:"content"`
	// The role of the messages author, in this case `user`.
	Role string `json:"role"`
	// An optional name for the participant. Provides the model information to
	// differentiate between participants of the same role.
	Name string `json:"name,omitempty"`
}

// ChatCompletionSystemMessageParam Developer-provided instructions that the model should follow, regardless of
// messages sent by the user. With o1 models and newer, use `developer` messages
// for this purpose instead.
type ChatCompletionSystemMessageParam struct {
	// The contents of the system message.
	Content StringOrArray `json:"content"`
	// The role of the messages author, in this case `system`.
	Role string `json:"role"`
	// An optional name for the participant. Provides the model information to
	// differentiate between participants of the same role.
	Name string `json:"name,omitempty"`
}

// ChatCompletionDeveloperMessageParam Developer-provided instructions that the model should follow, regardless of
// messages sent by the user. With o1 models and newer, use `developer` messages
// for this purpose instead.
type ChatCompletionDeveloperMessageParam struct {
	// The contents of the developer message.
	Content StringOrArray `json:"content"`
	// The role of the messages author, in this case `developer`.
	Role string `json:"role"`
	// An optional name for the participant. Provides the model information to
	// differentiate between participants of the same role.
	Name string `json:"name,omitempty"`
}

type ChatCompletionToolMessageParam struct {
	// The contents of the tool message.
	Content StringOrArray `json:"content"`
	// The role of the messages author, in this case `tool`.
	Role string `json:"role"`
	// Tool call that this message is responding to.
	ToolCallID string `json:"tool_call_id"`
}

// ChatCompletionAssistantMessageParamAudio Data about a previous audio response from the model.
// [Learn more](https://platform.openai.com/docs/guides/audio).
type ChatCompletionAssistantMessageParamAudio struct {
	// Unique identifier for a previous audio response from the model.
	ID string `json:"id"`
}

// ChatCompletionAssistantMessageParamContentType The type of the content part.
type ChatCompletionAssistantMessageParamContentType string

const (
	ChatCompletionAssistantMessageParamContentTypeText    ChatCompletionAssistantMessageParamContentType = "text"
	ChatCompletionAssistantMessageParamContentTypeRefusal ChatCompletionAssistantMessageParamContentType = "refusal"
)

// ChatCompletionAssistantMessageParamContent Learn about
// [text inputs](https://platform.openai.com/docs/guides/text-generation).
type ChatCompletionAssistantMessageParamContent struct {
	// The type of the content part.
	Type ChatCompletionAssistantMessageParamContentType `json:"type"`
	// The refusal message generated by the model.
	Refusal *string `json:"refusal,omitempty"`
	// The text content.
	Text *string `json:"text,omitempty"`
}

// ChatCompletionAssistantMessageParam Messages sent by the model in response to user messages.
type ChatCompletionAssistantMessageParam struct {
	// The role of the messages author, in this case `assistant`.
	Role string `json:"role"`
	// Data about a previous audio response from the model.
	// [Learn more](https://platform.openai.com/docs/guides/audio).
	Audio ChatCompletionAssistantMessageParamAudio `json:"audio,omitempty"`
	// The contents of the assistant message. Required unless `tool_calls` or
	// `function_call` is specified.
	Content ChatCompletionAssistantMessageParamContent `json:"content"`
	// An optional name for the participant. Provides the model information to
	// differentiate between participants of the same role.
	Name string `json:"name,omitempty"`
	// The refusal message by the assistant.
	Refusal string `json:"refusal,omitempty"`
	// The tool calls generated by the model, such as function calls.
	ToolCalls []ChatCompletionMessageToolCallParam `json:"tool_calls,omitempty"`
}

// ChatCompletionMessageToolCallType The type of the tool. Currently, only `function` is supported.
type ChatCompletionMessageToolCallType string

const (
	ChatCompletionMessageToolCallTypeFunction ChatCompletionMessageToolCallType = "function"
)

// ChatCompletionMessageToolCallFunctionParam The function that the model called.
type ChatCompletionMessageToolCallFunctionParam struct {
	// The arguments to call the function with, as generated by the model in JSON
	// format. Note that the model does not always generate valid JSON, and may
	// hallucinate parameters not defined by your function schema. Validate the
	// arguments in your code before calling your function.
	Arguments string `json:"arguments"`
	// The name of the function to call.
	Name string `json:"name"`
}

type ChatCompletionMessageToolCallParam struct {
	// The ID of the tool call.
	ID string `json:"id"`
	// The function that the model called.
	Function ChatCompletionMessageToolCallFunctionParam `json:"function"`
	// The type of the tool. Currently, only `function` is supported.
	Type ChatCompletionMessageToolCallType `json:"type"`
}

type ChatCompletionResponseFormatType string

const (
	ChatCompletionResponseFormatTypeJSONObject ChatCompletionResponseFormatType = "json_object"
	ChatCompletionResponseFormatTypeJSONSchema ChatCompletionResponseFormatType = "json_schema"
	ChatCompletionResponseFormatTypeText       ChatCompletionResponseFormatType = "text"
)

type ChatCompletionResponseFormat struct {
	Type       ChatCompletionResponseFormatType        `json:"type,omitempty"`
	JSONSchema *ChatCompletionResponseFormatJSONSchema `json:"json_schema,omitempty"` //nolint:tagliatelle //follow openai api
}

type ChatCompletionResponseFormatJSONSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      json.Marshaler `json:"schema"`
	Strict      bool           `json:"strict"`
}

// ChatCompletionRequest represents a request structure for chat completion API.
type ChatCompletionRequest struct {
	// Messages: A list of messages comprising the conversation so far.
	// Depending on the model you use, different message types (modalities) are supported,
	// like text, images, and audio.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-messages
	Messages []ChatCompletionMessageParamUnion `json:"messages"`

	// Model: ID of the model to use
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-model
	Model string `json:"model"`

	// FrequencyPenalty: Number between -2.0 and 2.0
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-frequency_penalty
	FrequencyPenalty *float32 `json:"frequency_penalty,omitempty"` //nolint:tagliatelle //follow openai api

	// LogitBias Modify the likelihood of specified tokens appearing in the completion.
	// It must be a token id string (specified by their token ID in the tokenizer), not a word string.
	// incorrect: `"logit_bias":{"You": 6}`, correct: `"logit_bias":{"1639": 6}`
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-logit_bias
	LogitBias map[string]int `json:"logit_bias,omitempty"` //nolint:tagliatelle //follow openai api

	// LogProbs indicates whether to return log probabilities of the output tokens or not.
	// If true, returns the log probabilities of each output token returned in the content of message.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-logprobs
	LogProbs *bool `json:"logprobs,omitempty"`

	// TopLogProbs is an integer between 0 and 5 specifying the number of most likely tokens to return at each
	// token position, each with an associated log probability.
	// logprobs must be set to true if this parameter is used.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-top_logprobs
	TopLogProbs *int `json:"top_logprobs,omitempty"` //nolint:tagliatelle //follow openai api

	// MaxTokens The maximum number of tokens that can be generated in the chat completion.
	// This value can be used to control costs for text generated via API.
	// This value is now deprecated in favor of max_completion_tokens, and is not compatible with o1 series models.
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-max_tokens
	MaxTokens *int64 `json:"max_tokens,omitempty"` //nolint:tagliatelle //follow openai api

	// N: LLM Gateway does not support multiple completions.
	// The only accepted value is 1.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-n
	N *int `json:"n,omitempty"`

	// PresencePenalty Positive values penalize new tokens based on whether they appear in the text so far,
	// increasing the model's likelihood to talk about new topics.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-presence_penalty
	PresencePenalty *float32 `json:"presence_penalty,omitempty"` //nolint:tagliatelle //follow openai api

	// ResponseFormat is only for GPT models.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-response_format
	ResponseFormat *ChatCompletionResponseFormat `json:"response_format,omitempty"` //nolint:tagliatelle //follow openai api

	// Seed: This feature is in Beta. If specified, our system will make a best effort to
	// sample deterministically, such that repeated requests with the same `seed` and
	// parameters should return the same result. Determinism is not guaranteed, and you
	// should refer to the `system_fingerprint` response parameter to monitor changes
	// in the backend.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-seed
	Seed *int `json:"seed,omitempty"`

	// Stop string / array / null Defaults to null
	// Up to 4 sequences where the API will stop generating further tokens.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-stop
	Stop []*string `json:"stop,omitempty"`

	// Stream: If set, partial message deltas will be sent, like in ChatGPT.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-stream
	Stream bool `json:"stream,omitempty"`

	// StreamOptions for streaming response. Only set this when you set stream: true.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-stream_options
	StreamOptions *StreamOptions `json:"stream_options,omitempty"` //nolint:tagliatelle //follow openai api

	// Temperature What sampling temperature to use, between 0 and 2.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-temperature
	Temperature *float64 `json:"temperature,omitempty"`

	// TopP An alternative to sampling with temperature, called nucleus sampling,
	// where the model considers the results of the tokens with top_p probability mass.
	// So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-top_p
	TopP *float64 `json:"top_p,omitempty"` //nolint:tagliatelle //follow openai api

	// Tools provide a list of tool definitions to be used by the LLM.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-response_format
	Tools []Tool `json:"tools,omitempty"`

	// ToolChoice specifies a specific tool to be used by name (given in the tool definition),
	// or use "auto" to auto select the most appropriate.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-tool_choice
	ToolChoice any `json:"tool_choice,omitempty"` //nolint:tagliatelle //follow openai api

	// ParallelToolCalls enables multiple tools to be returned by the model.
	// Docs: https://platform.openai.com/docs/guides/function-calling/parallel-function-calling
	ParallelToolCalls bool `json:"parallel_tool_calls,omitempty"` //nolint:tagliatelle //follow openai api

	// User: A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.
	// Docs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-user
	User string `json:"user,omitempty"`
}

type StreamOptions struct {
	// If set, an additional chunk will be streamed before the data: [DONE] message.
	// The usage field on this chunk shows the token usage statistics for the entire request,
	// and the choices field will always be an empty array.
	// All other chunks will also include a usage field, but with a null value.
	IncludeUsage bool `json:"include_usage,omitempty"` //nolint:tagliatelle //follow openai api
}

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

type Tool struct {
	Type     ToolType            `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`
}

type ToolChoice struct {
	Type     ToolType     `json:"type"`
	Function ToolFunction `json:"function,omitempty"`
}

type ToolFunction struct {
	Name string `json:"name"`
}

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Strict      bool   `json:"strict,omitempty"`
	// Parameters is an object describing the function.
	// You can pass json.RawMessage to describe the schema,
	// or you can pass in a struct which serializes to the proper JSON schema.
	// The jsonschema package is provided for convenience, but you should
	// consider another specialized library if you require more complex schemas.
	Parameters any `json:"parameters"`
}

// Deprecated: use FunctionDefinition instead.
type FunctionDefine = FunctionDefinition

type TopLogProbs struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"`
}

// LogProb represents the probability information for a token.
type LogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"` // Omitting the field if it is null
	// TopLogProbs is a list of the most likely tokens and their log probability, at this token position.
	// In rare cases, there may be fewer than the number of requested top_logprobs returned.
	TopLogProbs []TopLogProbs `json:"top_logprobs"` //nolint:tagliatelle //follow openai api
}

// LogProbs is the top-level structure containing the log probability information.
type LogProbs struct {
	// Content is a list of message content tokens with log probability information.
	Content []LogProb `json:"content"`
}

// ChatCompletionResponse represents a response from /v1/chat/completions.
// https://platform.openai.com/docs/api-reference/chat/object
type ChatCompletionResponse struct {
	// Choices are described in the OpenAI API documentation:
	// https://platform.openai.com/docs/api-reference/chat/object#chat/object-choices
	Choices []ChatCompletionResponseChoice `json:"choices,omitempty"`

	// Object is always "chat.completion" for completions.
	// https://platform.openai.com/docs/api-reference/chat/object#chat/object-object
	Object string `json:"object,omitempty"`

	// Usage is described in the OpenAI API documentation:
	// https://platform.openai.com/docs/api-reference/chat/object#chat/object-usage
	Usage ChatCompletionResponseUsage `json:"usage,omitempty"`
}

// ChatCompletionChoicesFinishReason The reason the model stopped generating tokens. This will be `stop` if the model
// hit a natural stop point or a provided stop sequence, `length` if the maximum
// number of tokens specified in the request was reached, `content_filter` if
// content was omitted due to a flag from our content filters, `tool_calls` if the
// model called a tool, or `function_call` (deprecated) if the model called a
// function.
type ChatCompletionChoicesFinishReason string

const (
	ChatCompletionChoicesFinishReasonStop          ChatCompletionChoicesFinishReason = "stop"
	ChatCompletionChoicesFinishReasonLength        ChatCompletionChoicesFinishReason = "length"
	ChatCompletionChoicesFinishReasonToolCalls     ChatCompletionChoicesFinishReason = "tool_calls"
	ChatCompletionChoicesFinishReasonContentFilter ChatCompletionChoicesFinishReason = "content_filter"
)

type ChatCompletionTokenLogprobTopLogprob struct {
	// The token.
	Token string `json:"token"`
	// A list of integers representing the UTF-8 bytes representation of the token.
	// Useful in instances where characters are represented by multiple tokens and
	// their byte representations must be combined to generate the correct text
	// representation. Can be `null` if there is no bytes representation for the token.
	Bytes []int64 `json:"bytes,omitempty"`
	// The log probability of this token, if it is within the top 20 most likely
	// tokens. Otherwise, the value `-9999.0` is used to signify that the token is very
	// unlikely.
	Logprob float64 `json:"logprob"`
}

type ChatCompletionTokenLogprob struct {
	// The token.
	Token string `json:"token"`
	// A list of integers representing the UTF-8 bytes representation of the token.
	// Useful in instances where characters are represented by multiple tokens and
	// their byte representations must be combined to generate the correct text
	// representation. Can be `null` if there is no bytes representation for the token.
	Bytes []int64 `json:"bytes,omitempty"`
	// The log probability of this token, if it is within the top 20 most likely
	// tokens. Otherwise, the value `-9999.0` is used to signify that the token is very
	// unlikely.
	Logprob float64 `json:"logprob"`
	// List of the most likely tokens and their log probability, at this token
	// position. In rare cases, there may be fewer than the number of requested
	// `top_logprobs` returned.
	TopLogprobs []ChatCompletionTokenLogprobTopLogprob `json:"top_logprobs"`
}

// ChatCompletionChoicesLogprobs Log probability information for the choice.
type ChatCompletionChoicesLogprobs struct {
	// A list of message content tokens with log probability information.
	Content []ChatCompletionTokenLogprob `json:"content,omitempty"`
	// A list of message refusal tokens with log probability information.
	Refusal []ChatCompletionTokenLogprob `json:"refusal,omitempty"`
}

// ChatCompletionResponseChoice is described in the OpenAI API documentation:
// https://platform.openai.com/docs/api-reference/chat/object#chat/object-choices
type ChatCompletionResponseChoice struct {
	// The reason the model stopped generating tokens. This will be `stop` if the model
	// hit a natural stop point or a provided stop sequence, `length` if the maximum
	// number of tokens specified in the request was reached, `content_filter` if
	// content was omitted due to a flag from our content filters, `tool_calls` if the
	// model called a tool, or `function_call` (deprecated) if the model called a
	// function.
	FinishReason ChatCompletionChoicesFinishReason `json:"finish_reason"`
	// The index of the choice in the list of choices.
	Index int64 `json:"index"`
	// Log probability information for the choice.
	Logprobs ChatCompletionChoicesLogprobs `json:"logprobs,omitempty"`
	// Message is described in the OpenAI API documentation:
	// https://platform.openai.com/docs/api-reference/chat/object#chat/object-choices
	Message ChatCompletionResponseChoiceMessage `json:"message,omitempty"`
}

// ChatCompletionResponseChoiceMessage is described in the OpenAI API documentation:
// https://platform.openai.com/docs/api-reference/chat/object#chat/object-choices
type ChatCompletionResponseChoiceMessage struct {
	// The contents of the message.
	Content *string `json:"content,omitempty"`

	// The role of the author of this message.
	Role string `json:"role,omitempty"`

	// The tool calls generated by the model, such as function calls.
	ToolCalls []ChatCompletionMessageToolCallParam `json:"tool_calls,omitempty"`
}

// ChatCompletionResponseUsage is described in the OpenAI API documentation:
// https://platform.openai.com/docs/api-reference/chat/object#chat/object-usage
type ChatCompletionResponseUsage struct {
	CompletionTokens int `json:"completion_tokens,omitempty"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// ChatCompletionResponseChunk is described in the OpenAI API documentation:
// https://platform.openai.com/docs/api-reference/chat/streaming#chat-create-messages
type ChatCompletionResponseChunk struct {
	// Choices are described in the OpenAI API documentation:
	// https://platform.openai.com/docs/api-reference/chat/streaming#chat/streaming-choices
	Choices []ChatCompletionResponseChunkChoice `json:"choices,omitempty"`

	// Object is always "chat.completion.chunk" for completions.
	// https://platform.openai.com/docs/api-reference/chat/streaming#chat/streaming-object
	Object string `json:"object,omitempty"`

	// Usage is described in the OpenAI API documentation:
	// https://platform.openai.com/docs/api-reference/chat/streaming#chat/streaming-usage
	Usage *ChatCompletionResponseUsage `json:"usage,omitempty"`
}

// String implements fmt.Stringer.
func (c *ChatCompletionResponseChunk) String() string {
	buf, _ := json.Marshal(c)
	return strings.ReplaceAll(string(buf), ",", ", ")
}

// ChatCompletionResponseChunkChoice is described in the OpenAI API documentation:
// https://platform.openai.com/docs/api-reference/chat/streaming#chat/streaming-choices
type ChatCompletionResponseChunkChoice struct {
	Delta        *ChatCompletionResponseChunkChoiceDelta `json:"delta,omitempty"`
	FinishReason ChatCompletionChoicesFinishReason       `json:"finish_reason,omitempty"`
}

// ChatCompletionResponseChunkChoiceDelta is described in the OpenAI API documentation:
// https://platform.openai.com/docs/api-reference/chat/streaming#chat/streaming-choices
type ChatCompletionResponseChunkChoiceDelta struct {
	Content   *string                              `json:"content,omitempty"`
	Role      string                               `json:"role"`
	ToolCalls []ChatCompletionMessageToolCallParam `json:"tool_calls,omitempty"`
}

// Error is described in the OpenAI API documentation
// https://platform.openai.com/docs/api-reference/realtime-server-events/error
type Error struct {
	// The unique ID of the server event.
	EventID *string `json:"event_id,omitempty"`
	// The event type, must be error.
	Type string `json:"type"`
	// Details of the error.
	Error ErrorType `json:"error"`
}

type ErrorType struct {
	// The type of error (e.g., "invalid_request_error", "server_error").
	Type string `json:"type"`
	// Error code, if any.
	Code *string `json:"code,omitempty"`
	// A human-readable error message.
	Message string `json:"message,omitempty"`
	// Parameter related to the error, if any.
	Param *string `json:"param,omitempty"`
	// The event_id of the client event that caused the error, if applicable.
	EventID *string `json:"event_id,omitempty"`
}
