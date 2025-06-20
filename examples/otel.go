package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/raywall/cloud-service-pack/go/metrics/otel"
)

var otelClient otel.OtelClient

func main() {
	// Inicializar cliente OpenTelemetry
	var err error
	otelClient, err = otel.New("my-go-service", "1.0.0", "http://localhost:4317")
	if err != nil {
		log.Fatal("Erro ao criar cliente OpenTelemetry:", err)
	}
	defer otelClient.Close()

	// Configurar rotas
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("Servidor iniciado na porta 8080")
	log.Println("Métricas sendo enviadas para OpenTelemetry Collector")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Erro ao iniciar servidor:", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	// Criar span para tracing
	ctx, span := otelClient.StartSpan(ctx, "home_handler", otel.OtelTags{
		otelClient.NewOtelTag("http.method", r.Method),
		otelClient.NewOtelTag("http.path", r.URL.Path),
	})
	defer span.End()

	// Simular processamento
	processingTime := time.Duration(rand.Intn(100)) * time.Millisecond
	time.Sleep(processingTime)

	// Criar tags para métricas
	tags := otel.OtelTags{
		otelClient.NewOtelTag("method", r.Method),
		otelClient.NewOtelTag("path", r.URL.Path),
		otelClient.NewOtelTag("status", "200"),
	}

	// Registrar métricas
	duration := time.Since(start).Seconds()

	// Incrementar contador de requests
	if err := otelClient.Increment(ctx, "http_requests_total", 1, tags); err != nil {
		log.Printf("Erro ao registrar contador: %v", err)
	}

	// Registrar tempo de resposta
	if err := otelClient.Histogram(ctx, "http_request_duration_seconds", duration, tags); err != nil {
		log.Printf("Erro ao registrar histograma: %v", err)
	}

	// Registrar evento no span
	otelClient.RecordEvent(ctx, "processing_completed", otel.OtelTags{
		otelClient.NewOtelTag("duration_ms", processingTime.Milliseconds()),
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, OpenTelemetry!"))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Criar span
	ctx, span := otelClient.StartSpan(ctx, "users_handler", otel.OtelTags{
		otelClient.NewOtelTag("http.method", r.Method),
		otelClient.NewOtelTag("http.path", r.URL.Path),
	})
	defer span.End()

	// Simular consulta ao banco
	ctx, dbSpan := otelClient.StartSpan(ctx, "database_query", otel.OtelTags{
		otelClient.NewOtelTag("db.operation", "SELECT"),
		otelClient.NewOtelTag("db.table", "users"),
	})

	time.Sleep(50 * time.Millisecond) // Simular latência do banco
	dbSpan.End()

	// Registrar métrica de usuários ativos (exemplo usando gauge)
	activeUsers := float64(rand.Intn(1000) + 100)
	if err := otelClient.Gauge(ctx, "active_users", activeUsers, otel.OtelTags{
		otelClient.NewOtelTag("region", "us-east-1"),
	}); err != nil {
		log.Printf("Erro ao registrar gauge: %v", err)
	}

	// Incrementar contador específico para users endpoint
	tags := otel.OtelTags{
		otelClient.NewOtelTag("method", r.Method),
		otelClient.NewOtelTag("endpoint", "users"),
	}

	if err := otelClient.Increment(ctx, "users_endpoint_calls", 1, tags); err != nil {
		log.Printf("Erro ao registrar contador: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"users": ["alice", "bob", "charlie"]}`))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Exemplo de UpDownCounter para monitorar conexões ativas
	if err := otelClient.UpDownCounter(ctx, "active_connections", 1, otel.OtelTags{
		otelClient.NewOtelTag("type", "health_check"),
	}); err != nil {
		log.Printf("Erro ao registrar up-down counter: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
