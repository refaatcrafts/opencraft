package tools

import "google.golang.org/genai"

// Registry holds all registered tools.
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) Tool {
	return r.tools[name]
}

// ToGeminiTools converts the registry to the Gemini SDK tool slice.
func (r *Registry) ToGeminiTools() []*genai.Tool {
	decls := make([]*genai.FunctionDeclaration, 0, len(r.tools))
	for _, t := range r.tools {
		decls = append(decls, &genai.FunctionDeclaration{
			Name:                 t.Name(),
			Description:          t.Description(),
			ParametersJsonSchema: t.InputSchema(),
		})
	}
	return []*genai.Tool{{FunctionDeclarations: decls}}
}
