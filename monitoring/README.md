# 🚀 LVTN Monitoring Setup

## Quick Start

### 1. Start Monitoring Stack
```bash
cd monitoring
docker-compose -f docker-compose.monitoring.yml up -d
```

### 2. Access URLs
- **Grafana**: http://localhost:3000 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **Node Exporter**: http://localhost:9100/metrics

### 3. First Time Setup

#### Grafana Login
1. Go to http://localhost:3000
2. Login: `admin` / `admin123`
3. Dashboard sẽ tự động load

#### Check Prometheus Targets
1. Go to http://localhost:9090/targets
2. Verify all targets are UP

## 📊 Learning Path

### Day 1: Basics
- [ ] Explore Grafana UI
- [ ] Understand Prometheus metrics
- [ ] Create your first panel
- [ ] Learn PromQL basics

### Day 2: Custom Metrics
- [ ] Add metrics to Go services
- [ ] Create service dashboards
- [ ] Monitor request rates & latency

### Day 3: Alerting
- [ ] Setup alert rules
- [ ] Configure notifications
- [ ] Test alerting scenarios

## 🔧 Next Steps

After monitoring is running:
1. **Add metrics to Go services** (we'll do this next)
2. **Create service-specific dashboards**
3. **Setup alerting for critical metrics**
4. **Monitor business metrics**

## 📚 Grafana Learning Resources

### Key Concepts to Learn:
- **Panels**: Basic building blocks
- **Queries**: PromQL language
- **Variables**: Dynamic dashboards
- **Alerts**: Proactive monitoring
- **Datasources**: Data connections

### Useful PromQL Examples:
```promql
# CPU Usage
100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Memory Usage
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100

# Request Rate (per second)
rate(http_requests_total[5m])

# Request Duration 95th percentile
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error Rate
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])
```

## 🎯 Goals

By end of Phase 1, you'll understand:
- How Prometheus collects metrics
- How to create Grafana dashboards
- How to write PromQL queries
- How to setup alerts
- How to monitor application performance

## ⚠️ Troubleshooting

### Common Issues:
1. **Targets down**: Check service URLs in prometheus.yml
2. **No data**: Verify metrics endpoints exist
3. **Dashboard empty**: Check datasource connection
4. **Alerts not firing**: Verify alert rules syntax

### Useful Commands:
```bash
# Check container logs
docker-compose logs grafana
docker-compose logs prometheus

# Restart services
docker-compose restart grafana
docker-compose restart prometheus

# Update configuration
docker-compose down && docker-compose up -d
```