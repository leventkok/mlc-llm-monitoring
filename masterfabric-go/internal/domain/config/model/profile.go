package model

// ModelProfile is a pre-registered model/adapter the admin can hot-swap to.
type ModelProfile struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	RequestModel string `json:"request_model"` // OpenAI model id sent to MLC (Render MLC_LLM_MODEL)
	EngineModel  string `json:"engine_model"`  // MLC_MODEL env for local mlc-engine (HF://…)
	LocalAdapter string `json:"local_adapter"` // MLC_LORA_ADAPTER (merged dir name), empty = HF engine model
	Description  string `json:"description,omitempty"`
}

// DefaultModelProfiles are built-in profiles for the hybrid stack.
func DefaultModelProfiles() []ModelProfile {
	return []ModelProfile{
		{
			ID:           "inferreview-v2",
			Label:        "InferReview LoRA v2",
			RequestModel: "HF://levonov/inferreview-gemma-lora-v2-q4f16_1-MLC",
			EngineModel:  "HF://levonov/inferreview-gemma-lora-v2-q4f16_1-MLC",
			LocalAdapter: "",
			Description:  "Production fine-tuned classifier (Colab v2, 240 examples)",
		},
		{
			ID:           "base-gemma",
			Label:        "Base Gemma 2B",
			RequestModel: "HF://mlc-ai/gemma-2-2b-it-q4f16_1-MLC",
			EngineModel:  "HF://mlc-ai/gemma-2-2b-it-q4f16_1-MLC",
			LocalAdapter: "",
			Description:  "Unfine-tuned baseline for A/B comparison",
		},
	}
}

func FindProfile(profiles []ModelProfile, id string) (ModelProfile, bool) {
	for _, p := range profiles {
		if p.ID == id {
			return p, true
		}
	}
	return ModelProfile{}, false
}
