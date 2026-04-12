package pipeline

import "context"

// IntentProcessor is the interface that all pipeline processors must implement
type IntentProcessor interface {
	// Process processes the pipeline context
	Process(ctx context.Context, pCtx *PipelineContext) error
	// Name returns the processor name for logging
	Name() string
}

// ProcessorRegistry holds registered processors by name
type ProcessorRegistry struct {
	processors map[string]IntentProcessor
}

// NewProcessorRegistry creates a new ProcessorRegistry
func NewProcessorRegistry() *ProcessorRegistry {
	return &ProcessorRegistry{
		processors: make(map[string]IntentProcessor),
	}
}

// Register registers a processor
func (r *ProcessorRegistry) Register(processor IntentProcessor) {
	if processor == nil {
		return
	}
	r.processors[processor.Name()] = processor
}

// Get returns a processor by name
func (r *ProcessorRegistry) Get(name string) IntentProcessor {
	return r.processors[name]
}

// List returns all registered processors
func (r *ProcessorRegistry) List() []IntentProcessor {
	result := make([]IntentProcessor, 0, len(r.processors))
	for _, p := range r.processors {
		result = append(result, p)
	}
	return result
}
