# Advanced Monitoring Guide

## 🎯 **Overview**
This guide covers the advanced monitoring setup for the LVTN project, including GraphQL metrics, business analytics, alerting, and performance monitoring.

## 📊 **Monitoring Architecture**

### Core Components
- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboards  
- **GraphQL Metrics**: Request/response tracking
- **Business Metrics**: Domain-specific analytics
- **Alert Manager**: Alert handling (optional)

## 🔧 **Setup Instructions**

### 1. Start Monitoring Stack
```bash
cd monitoring
docker-compose -f docker-compose.monitoring.yml up -d
```

### 2. Access Monitoring Services
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **GraphQL Gateway**: http://localhost:8080/metrics

## 📈 **Available Dashboards**

### 1. Services Monitoring Dashboard
- Service health status
- Request rates and latencies
- Error rates by service
- Database performance

### 2. Advanced Monitoring Dashboard
- GraphQL operation metrics
- Business logic analytics
- Field resolution performance
- Custom business metrics

## 🚨 **Alert Rules**

### Service Health Alerts
- **ServiceDown**: Service unavailable for >30s
- **ServiceUnhealthy**: Health check failing for >1m

### Performance Alerts
- **HighGraphQLErrorRate**: Error rate >10% for 2m
- **SlowGraphQLRequests**: 95th percentile >2s for 5m
- **SlowDatabaseQueries**: DB queries >1s for 5m

### Resource Alerts
- **HighCPUUsage**: CPU >80% for 5m
- **HighMemoryUsage**: Memory >80% for 5m
- **LowDiskSpace**: Disk >85% full for 5m

### Business Alerts
- **NoThesisSubmissions**: No submissions for 2h
- **HighFileUploadFailures**: Upload failures >10%

## 📊 **Business Metrics**

### Thesis Management
```go
// Record thesis submission
metrics.RecordThesisSubmission("2024", "CS", "submitted")

// Track approval time
metrics.RecordThesisApprovalTime("initial_review", "CS", 7.5)

// Update active theses count
metrics.SetActiveThesesByStage("under_review", "CS", 25)
```

### User Activity
```go
// Track user login
metrics.RecordUserLogin("student", "CS", "morning")

// Record session duration
metrics.RecordUserSessionDuration("lecturer", "grading", time.Hour*2)
```

### File Operations
```go
// Track file uploads
metrics.RecordFileUpload("pdf", "thesis", "CS")

// Record processing time
metrics.RecordFileProcessingTime("upload", "pdf", time.Second*5)
```

## 🎯 **GraphQL Metrics**

### Request Tracking
- Total requests by operation and status
- Request duration percentiles
- Active connection count
- Field resolution performance

### Custom GraphQL Middleware
```go
// Add to GraphQL server
srv.Use(middleware.GraphQLMetrics())
srv.Use(middleware.BusinessMetrics())
```

## 🔍 **Key Metrics to Monitor**

### Performance Metrics
- `graphql_request_duration_seconds`
- `graphql_requests_total`
- `database_query_duration_seconds`
- `grpc_request_duration_seconds`

### Business Metrics
- `thesis_submissions_total`
- `user_activities_total`
- `file_operations_total`
- `council_meetings_total`

### System Metrics
- `service_health`
- `graphql_active_connections`
- `node_cpu_seconds_total`
- `node_memory_MemAvailable_bytes`

## 📱 **Grafana Dashboard Features**

### System Overview Panel
- Services up/down status
- Overall system health
- Active connections

### Performance Panels
- Request rate graphs
- Response time percentiles
- Error rate tracking
- Database performance

### Business Analytics
- Thesis workflow metrics
- User activity patterns
- File upload statistics
- Council operation tracking

## ⚡ **Performance Optimization**

### Metrics Collection
- Use appropriate bucket sizes for histograms
- Limit high-cardinality labels
- Sample expensive metrics if needed

### Alert Tuning
- Adjust thresholds based on baseline performance
- Use different severity levels (info/warning/critical)
- Add runbook URLs to alerts

## 🛠 **Troubleshooting**

### Common Issues

#### Metrics Not Appearing
1. Check service is exposing metrics endpoint
2. Verify Prometheus scrape configuration
3. Check network connectivity between services

#### Grafana Dashboard Empty
1. Verify data source configuration
2. Check query syntax
3. Confirm metric names match

#### High Alert Volume
1. Review alert thresholds
2. Add alert grouping/suppression
3. Check for metric spikes

### Debugging Commands
```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Test metric endpoint
curl http://localhost:8080/metrics

# View container logs
docker logs prometheus
docker logs grafana
```

## 🚀 **Next Steps**

### 1. Distributed Tracing
- Add Jaeger for request tracing
- Implement OpenTelemetry
- Track cross-service requests

### 2. Load Testing
- Create performance baselines
- Monitor under load
- Identify bottlenecks

### 3. Advanced Analytics
- Add machine learning for anomaly detection
- Create predictive alerts
- Implement capacity planning

## 🔗 **Resources**

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [GraphQL Metrics Best Practices](https://prometheus.io/docs/practices/naming/)
- [Alert Manager Configuration](https://prometheus.io/docs/alerting/latest/configuration/)

## 📝 **Maintenance**

### Regular Tasks
- Review and adjust alert thresholds
- Update dashboards based on new requirements
- Clean up old metrics data
- Monitor storage usage

### Metric Retention
- Default: 15 days for Prometheus
- Adjust based on storage requirements
- Consider long-term storage solutions