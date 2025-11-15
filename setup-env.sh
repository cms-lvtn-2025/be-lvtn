#!/bin/bash

# Environment Setup Script for LVTN Project

echo "🚀 Setting up LVTN environment files..."

# Check if running in correct directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# Create logs directory
echo "📁 Creating logs directory..."
mkdir -p logs

# Create certs directories
echo "🔐 Creating certificate directories..."
mkdir -p certs/{ca,clients,academic,council,file,role,thesis,user}

# Set permissions for development
echo "🔧 Setting development permissions..."
chmod 755 logs
chmod -R 755 certs

# Check if Docker is running
echo "🐳 Checking Docker status..."
if ! docker info > /dev/null 2>&1; then
    echo "⚠️  Warning: Docker is not running. Please start Docker for full functionality."
else
    echo "✅ Docker is running"
fi

# Check if MySQL is accessible
echo "🗄️  Checking MySQL connection..."
if command -v mysql > /dev/null 2>&1; then
    if mysql -h localhost -P 3306 -u root -ppassword -e "SELECT 1;" > /dev/null 2>&1; then
        echo "✅ MySQL connection successful"
    else
        echo "⚠️  Warning: Cannot connect to MySQL. Please check database configuration."
    fi
else
    echo "⚠️  Warning: MySQL client not found. Please install MySQL or update connection settings."
fi

# Check if Redis is accessible
echo "📦 Checking Redis connection..."
if command -v redis-cli > /dev/null 2>&1; then
    if redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then
        echo "✅ Redis connection successful"
    else
        echo "⚠️  Warning: Cannot connect to Redis. Please check Redis configuration."
    fi
else
    echo "⚠️  Warning: Redis client not found. Please install Redis or update connection settings."
fi

echo ""
echo "✅ Environment setup completed!"
echo ""
echo "📋 Next steps:"
echo "1. Update database credentials in .env files if needed"
echo "2. Generate TLS certificates for services"
echo "3. Start services using: make run-all"
echo "4. Check Grafana monitoring at: http://localhost:3000"
echo ""
echo "📁 Environment files created:"
echo "  - src/service/academic/academic.env"
echo "  - src/service/user/user.env"
echo "  - src/service/file/file.env"
echo "  - src/service/thesis/thesis.env"
echo "  - src/service/council/council.env"
echo "  - src/service/role/role.env"
echo "  - src/server/.server.env"
echo "  - plagiarism-checker-service/.env"
echo ""
echo "🔒 Security note: Update passwords and secrets before deploying to production!"