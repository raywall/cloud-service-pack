#!/bin/bash

echo "🚀 Configurando ambiente OpenTelemetry local..."

# Criar diretórios necessários
mkdir -p grafana/provisioning/datasources
mkdir -p grafana/provisioning/dashboards

# Criar arquivo de datasource para Grafana
cat > grafana/provisioning/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF

# Criar arquivo de dashboard provider para Grafana
cat > grafana/provisioning/dashboards/dashboard.yml << 'EOF'
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

echo "📁 Estrutura de arquivos criada:"
echo "├── docker-compose.yml"
echo "├── otel-collector-config.yaml"
echo "├── prometheus.yml"
echo "├── grafana/"
echo "│   └── provisioning/"
echo "│       ├── datasources/"
echo "│       │   └── prometheus.yml"
echo "│       └── dashboards/"
echo "│           └── dashboard.yml"
echo "├── main.go (exemplo de aplicação Go)"
echo "├── go.mod"
echo "└── setup.sh"

echo ""
echo "🐳 Para iniciar o ambiente, execute:"
echo "docker-compose up -d"
echo ""
echo "📊 Acesse:"
echo "- Grafana: http://localhost:3000 (admin/admin)"
echo "- Prometheus: http://localhost:9090"
echo "- OpenTelemetry Collector: http://localhost:8888/metrics"
echo ""
echo "🔧 Para testar sua aplicação Go:"
echo "1. go mod tidy"
echo "2. go run main.go"
echo "3. curl http://localhost:8080"
echo ""
echo "✅ Setup concluído!"