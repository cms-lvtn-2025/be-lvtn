# 🔧 Environment Configuration Guide

## 📋 Overview

Dự án LVTN sử dụng environment files để configure các services. Mỗi service có file .env riêng biệt để dễ quản lý và deploy.

## 📁 Environment Files Structure

```
LVTN/
├── .env.production.template     # Production template
├── src/server/.server.env       # GraphQL Gateway
├── src/service/
│   ├── academic/academic.env    # Academic Service
│   ├── council/council.env      # Council Service  
│   ├── file/file.env           # File Service
│   ├── role/role.env           # Role Service
│   ├── thesis/thesis.env       # Thesis Service
│   └── user/user.env           # User Service
└── plagiarism-checker-service/
    └── .env                    # Plagiarism Checker Service
```

## 🚀 Quick Setup

### Development Environment:
```bash
# Run setup script
chmod +x setup-env.sh
./setup-env.sh

# Or manually create logs and certs directories
mkdir -p logs
mkdir -p certs/{ca,clients,academic,council,file,role,thesis,user}
```

### Production Environment:
```bash
# Copy production template
cp .env.production.template .env.production

# Update with production values
# Then copy to each service directory with appropriate names
```

## 🔑 Key Configuration Categories

### 1. Service Configuration
- `SERVICE_NAME`: Service identifier
- `SERVICE_PORT`: gRPC service port  
- `METRICS_PORT`: Prometheus metrics port

### 2. Database Configuration
- `DB_HOST`, `DB_PORT`: Database connection
- `DB_USER`, `DB_PASSWORD`: Database credentials
- `DB_NAME`: Database name

### 3. TLS/Security Configuration
- `TLS_ENABLED`: Enable/disable TLS
- `CERT_DIR`: Certificate directory path
- `CERT_FILE`, `KEY_FILE`: Certificate files

### 4. External Services
- **Redis**: Cache and queues
- **MongoDB**: Document storage
- **MinIO**: File storage
- **Google OAuth**: Authentication

## 📊 Port Allocation

### gRPC Services:
- Academic: `50051` (metrics: `9091`)
- Council: `50052` (metrics: `9092`)
- File: `50053` (metrics: `9093`)
- Role: `50054` (metrics: `9094`)
- Thesis: `50055` (metrics: `9095`)
- User: `50056` (metrics: `9096`)

### HTTP Services:
- GraphQL Gateway: `8080` (metrics: `8081`)
- Plagiarism Service: `3000` (metrics: `3001`)

### Infrastructure:
- MySQL: `3306`
- Redis: `6379`
- MongoDB: `10000`
- MinIO: `9000`
- Prometheus: `9090`
- Grafana: `3000`

## 🔒 Security Best Practices

### Development:
- Use simple passwords for local development
- Keep TLS enabled even in development
- Use localhost for all connections

### Production:
- **Change all default passwords**
- Use strong, unique secrets for JWT
- Use proper TLS certificates
- Restrict database access
- Use environment-specific credentials

## 🛠️ Customization

### Database Configuration:
```bash
# Update in all service .env files:
DB_HOST=your-database-host
DB_USER=your-database-user
DB_PASSWORD=your-secure-password
DB_NAME=your-database-name
```

### Adding New Environment Variables:
1. Add to appropriate service .env file
2. Update service code to read the variable
3. Document the variable in this guide
4. Add to production template if needed

## 🧪 Testing Configuration

### Check Environment Loading:
```bash
# Test academic service
cd src/service/academic
go run main.go

# Should not show "env file not found" errors
```

### Verify Database Connection:
```bash
# Test MySQL connection
mysql -h localhost -P 3306 -u root -ppassword -e "SELECT 1;"

# Test Redis connection  
redis-cli -h localhost -p 6379 ping
```

### Check Metrics Endpoints:
```bash
# Test service metrics (after starting service)
curl http://localhost:9091/metrics   # Academic service
curl http://localhost:9096/metrics   # User service
# etc.
```

## 🚨 Common Issues

### "env file not found":
- Ensure .env file exists in service directory
- Check file name matches what service expects
- Verify file permissions

### Database connection errors:
- Check database is running
- Verify credentials in .env file
- Test connection manually

### TLS certificate errors:
- Ensure certificate directories exist
- Generate certificates if missing
- Check certificate paths in .env

### Port conflicts:
- Check if ports are already in use
- Update port numbers in .env files
- Ensure Prometheus config matches

## 📝 Environment Variables Reference

### Required for All Services:
- `SERVICE_NAME`: Service identifier
- `SERVICE_PORT`: gRPC port
- `METRICS_PORT`: Metrics port
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: Database config
- `TLS_ENABLED`: TLS configuration
- `LOG_LEVEL`: Logging level

### Service-Specific:
- **File Service**: MinIO configuration
- **GraphQL Gateway**: JWT, OAuth, Redis, MongoDB config
- **Plagiarism Service**: Worker config, Bull Board config

### Optional:
- `ENV`: Environment (development/production)
- `LOG_DIR`: Custom log directory
- `HEALTH_CHECK_*`: Health check settings

## 🔄 Environment Updates

When updating environment variables:
1. Update .env files
2. Restart affected services
3. Check metrics for any errors
4. Update documentation if needed
5. Update production template if applicable

---

**🚀 Ready to run! All environment files are configured for development use.**