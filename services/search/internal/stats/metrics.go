package stats

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerMetricsOnce sync.Once
)

type Metrics struct {
	conversations      *prometheus.CounterVec
	turns              *prometheus.CounterVec
	ragTurns           *prometheus.CounterVec
	retrievalSuccess   *prometheus.CounterVec
	retrievalEmpty     *prometheus.CounterVec
	answerSuccess      *prometheus.CounterVec
	answerFailure      *prometheus.CounterVec
	retrievalDuration  *prometheus.HistogramVec
	generationDuration *prometheus.HistogramVec
	turnDuration       *prometheus.HistogramVec
}

func NewMetrics(registerer prometheus.Registerer) *Metrics {
	m := &Metrics{
		conversations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_conversation_total",
			Help: "Total number of unique conversations observed by the RAG flow.",
		}, metricLabelNames()),
		turns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_turn_total",
			Help: "Total number of conversation turns observed.",
		}, metricLabelNames()),
		ragTurns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_turn_rag_total",
			Help: "Total number of turns that entered RAG retrieval.",
		}, metricLabelNames()),
		retrievalSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_retrieval_success_total",
			Help: "Total number of RAG turns that retrieved at least one document.",
		}, metricLabelNames()),
		retrievalEmpty: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_retrieval_empty_total",
			Help: "Total number of RAG turns that retrieved zero documents.",
		}, metricLabelNames()),
		answerSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_answer_success_total",
			Help: "Total number of RAG turns that returned a successful answer.",
		}, metricLabelNames()),
		answerFailure: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_answer_failure_total",
			Help: "Total number of RAG turns that failed to return an answer.",
		}, metricLabelNames()),
		retrievalDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "rag_retrieval_duration_seconds",
			Help:    "Duration of the retrieval phase for RAG turns.",
			Buckets: prometheus.DefBuckets,
		}, metricLabelNames()),
		generationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "rag_generation_duration_seconds",
			Help:    "Duration of the answer generation phase for RAG turns.",
			Buckets: prometheus.DefBuckets,
		}, metricLabelNames()),
		turnDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "rag_turn_duration_seconds",
			Help:    "End-to-end duration of a RAG turn.",
			Buckets: prometheus.DefBuckets,
		}, metricLabelNames()),
	}

	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	registerMetricsOnce.Do(func() {
		registerer.MustRegister(
			m.conversations,
			m.turns,
			m.ragTurns,
			m.retrievalSuccess,
			m.retrievalEmpty,
			m.answerSuccess,
			m.answerFailure,
			m.retrievalDuration,
			m.generationDuration,
			m.turnDuration,
		)
	})

	return m
}

func NewMetricsForRegistry(registerer prometheus.Registerer) *Metrics {
	m := &Metrics{
		conversations: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "rag_conversation_total", Help: "Total number of unique conversations observed by the RAG flow."}, metricLabelNames()),
		turns:         prometheus.NewCounterVec(prometheus.CounterOpts{Name: "rag_turn_total", Help: "Total number of conversation turns observed."}, metricLabelNames()),
		ragTurns:      prometheus.NewCounterVec(prometheus.CounterOpts{Name: "rag_turn_rag_total", Help: "Total number of turns that entered RAG retrieval."}, metricLabelNames()),
		retrievalSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_retrieval_success_total", Help: "Total number of RAG turns that retrieved at least one document.",
		}, metricLabelNames()),
		retrievalEmpty: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_retrieval_empty_total", Help: "Total number of RAG turns that retrieved zero documents.",
		}, metricLabelNames()),
		answerSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_answer_success_total", Help: "Total number of RAG turns that returned a successful answer.",
		}, metricLabelNames()),
		answerFailure: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rag_answer_failure_total", Help: "Total number of RAG turns that failed to return an answer.",
		}, metricLabelNames()),
		retrievalDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "rag_retrieval_duration_seconds", Help: "Duration of the retrieval phase for RAG turns.", Buckets: prometheus.DefBuckets,
		}, metricLabelNames()),
		generationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "rag_generation_duration_seconds", Help: "Duration of the answer generation phase for RAG turns.", Buckets: prometheus.DefBuckets,
		}, metricLabelNames()),
		turnDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "rag_turn_duration_seconds", Help: "End-to-end duration of a RAG turn.", Buckets: prometheus.DefBuckets,
		}, metricLabelNames()),
	}

	registerer.MustRegister(
		m.conversations,
		m.turns,
		m.ragTurns,
		m.retrievalSuccess,
		m.retrievalEmpty,
		m.answerSuccess,
		m.answerFailure,
		m.retrievalDuration,
		m.generationDuration,
		m.turnDuration,
	)
	return m
}

func (m *Metrics) IncConversation(labels MetricLabels) {
	m.conversations.WithLabelValues(labelValues(labels)...).Inc()
}

func (m *Metrics) IncTurn(labels MetricLabels) {
	m.turns.WithLabelValues(labelValues(labels)...).Inc()
}

func (m *Metrics) IncRAGTurn(labels MetricLabels) {
	m.ragTurns.WithLabelValues(labelValues(labels)...).Inc()
}

func (m *Metrics) IncRetrievalSuccess(labels MetricLabels) {
	m.retrievalSuccess.WithLabelValues(labelValues(labels)...).Inc()
}

func (m *Metrics) IncRetrievalEmpty(labels MetricLabels) {
	m.retrievalEmpty.WithLabelValues(labelValues(labels)...).Inc()
}

func (m *Metrics) IncAnswerSuccess(labels MetricLabels) {
	m.answerSuccess.WithLabelValues(labelValues(labels)...).Inc()
}

func (m *Metrics) IncAnswerFailure(labels MetricLabels) {
	m.answerFailure.WithLabelValues(labelValues(labels)...).Inc()
}

func (m *Metrics) ObserveRetrievalDuration(labels MetricLabels, duration time.Duration) {
	m.retrievalDuration.WithLabelValues(labelValues(labels)...).Observe(duration.Seconds())
}

func (m *Metrics) ObserveGenerationDuration(labels MetricLabels, duration time.Duration) {
	m.generationDuration.WithLabelValues(labelValues(labels)...).Observe(duration.Seconds())
}

func (m *Metrics) ObserveTurnDuration(labels MetricLabels, duration time.Duration) {
	m.turnDuration.WithLabelValues(labelValues(labels)...).Observe(duration.Seconds())
}

func metricLabelNames() []string {
	return []string{"app_id", "knowledge_base_id", "channel", "rag_strategy", "status"}
}

func labelValues(labels MetricLabels) []string {
	return []string{
		emptyToUnknown(labels.AppID),
		emptyToUnknown(labels.KnowledgeBaseID),
		emptyToUnknown(labels.Channel),
		emptyToUnknown(labels.RAGStrategy),
		emptyToUnknown(labels.Status),
	}
}
