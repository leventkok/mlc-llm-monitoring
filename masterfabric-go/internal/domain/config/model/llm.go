package model

// LLMConfig holds runtime LLM parameters editable from the admin panel.
type LLMConfig struct {
	SystemPrompt  string  `json:"system_prompt"`
	Temperature   float64 `json:"temperature"`
	MaxTokens     int     `json:"max_tokens"`
	TopP          float64 `json:"top_p"`
	ActiveModel   string  `json:"active_model"`
	ActiveAdapter string  `json:"active_adapter"`
}

// DefaultSystemPrompt is the production classifier prompt (matches original mlc client).
const DefaultSystemPrompt = `You are a strict classifier for app store reviews (any language).
Classify into exactly one category and one sentiment.
Categories: bug, feature, praise, spam, other.
Sentiments: positive, negative, neutral.

Rules:
- Match the reviewer's tone: complaints and dissatisfaction → negative; compliments → positive; factual/neutral → neutral.
- bug: crashes, errors, broken or slow functionality.
- feature: requests for new capability.
- praise: explicit compliments.
- spam: promotional junk or fake reviews.
- other: general feedback that does not fit above (still use the correct sentiment).

Examples:
{"category":"bug","sentiment":"negative"} — "App keeps crashing"
{"category":"other","sentiment":"negative"} — "This app is terrible" / "Kötü bir uygulama"
{"category":"praise","sentiment":"positive"} — "Love this app!"

Respond with ONLY a JSON object and nothing else.`

func DefaultLLMConfig(activeModel, activeAdapter string) LLMConfig {
	return LLMConfig{
		SystemPrompt:  DefaultSystemPrompt,
		Temperature:   0,
		MaxTokens:     60,
		TopP:          1,
		ActiveModel:   activeModel,
		ActiveAdapter: activeAdapter,
	}
}
