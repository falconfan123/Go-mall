package pipeline

import (
	"context"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

// ProcessorOrder defines the execution order and parallelism
type ProcessorOrder int

const (
	// OrderPreprocessor must run first (produces tokens)
	OrderPreprocessor ProcessorOrder = iota
	// OrderEntityExtractor runs after Preprocessor (consumes tokens)
	OrderEntityExtractor
	// OrderCategoryPredictor can run in parallel with EntityExtractor
	OrderCategoryPredictor
	// OrderRewriter must run last (consumes all previous results)
	OrderRewriter
)

// PipelineExecutor executes processors in order with automatic parallel execution
type PipelineExecutor struct {
	processors []IntentProcessor
	logger     logx.Logger
}

// NewPipelineExecutor creates a new PipelineExecutor
func NewPipelineExecutor(processors ...IntentProcessor) *PipelineExecutor {
	return &PipelineExecutor{
		processors: processors,
		logger:     logx.WithContext(context.Background()),
	}
}

// Execute runs all processors in the appropriate order
// It uses automatic parallel execution for independent processors
func (e *PipelineExecutor) Execute(ctx context.Context, pCtx *PipelineContext) error {
	if len(e.processors) == 0 {
		return nil
	}

	// Phase 1: Preprocessor (must run first)
	e.logger.Debug("Executing processor: %s", e.processors[OrderPreprocessor].Name())
	if err := e.processors[OrderPreprocessor].Process(ctx, pCtx); err != nil {
		pCtx.SetError(err)
		return err
	}

	// Phase 2: EntityExtractor and CategoryPredictor can run in parallel
	// Both depend on Preprocessor but not on each other
	var wg sync.WaitGroup
	var entityErr, categoryErr error

	wg.Add(2)

	// EntityExtractor
	go func() {
		defer wg.Done()
		if pCtx.IsCancelled() {
			return
		}
		e.logger.Debug("Executing processor: %s", e.processors[OrderEntityExtractor].Name())
		if err := e.processors[OrderEntityExtractor].Process(ctx, pCtx); err != nil {
			entityErr = err
			pCtx.SetError(err)
		}
	}()

	// CategoryPredictor (can run in parallel with EntityExtractor)
	go func() {
		defer wg.Done()
		if pCtx.IsCancelled() {
			return
		}
		e.logger.Debug("Executing processor: %s", e.processors[OrderCategoryPredictor].Name())
		if err := e.processors[OrderCategoryPredictor].Process(ctx, pCtx); err != nil {
			categoryErr = err
			pCtx.SetError(err)
		}
	}()

	wg.Wait()

	// Return error if any occurred
	if entityErr != nil {
		return entityErr
	}
	if categoryErr != nil {
		return categoryErr
	}

	// Phase 3: Rewriter (must run last, depends on all previous results)
	if len(e.processors) > int(OrderRewriter) && e.processors[int(OrderRewriter)] != nil {
		e.logger.Debug("Executing processor: %s", e.processors[int(OrderRewriter)].Name())
		if err := e.processors[int(OrderRewriter)].Process(ctx, pCtx); err != nil {
			pCtx.SetError(err)
			return err
		}
	}

	return nil
}

// AddProcessor adds a processor to the executor
func (e *PipelineExecutor) AddProcessor(processor IntentProcessor) {
	if processor == nil {
		return
	}
	e.processors = append(e.processors, processor)
}
