#!/bin/bash

# Quick script to add metrics to remaining services
# Services to update: thesis, council, role

SERVICES=("thesis" "council" "role")
PORTS=("50055:9095" "50052:9092" "50054:9094")

for i in "${!SERVICES[@]}"; do
    SERVICE=${SERVICES[$i]}
    IFS=':' read -r SERVICE_PORT METRICS_PORT <<< "${PORTS[$i]}"
    
    echo "Updating $SERVICE service..."
    
    # Update main.go imports
    sed -i 's|"os"|"net/http"\n\t"os"|g' "src/service/$SERVICE/main.go"
    sed -i 's|"thaily/src/service/pkg/logger"|"thaily/src/service/pkg/logger"\n\t"thaily/src/service/pkg/metrics"|g' "src/service/$SERVICE/main.go"
    
    # Add metrics to main function (this is simplified - manual update recommended)
    echo "Manual update required for $SERVICE service main.go"
    
    # Update handler.go
    sed -i 's|"database/sql"|"database/sql"\n\t"thaily/src/service/pkg/metrics"|g' "src/service/$SERVICE/handler/handler.go"
    
    echo "$SERVICE service updated (partial - manual completion needed)"
done

echo "Script completed. Manual updates required for full metrics integration."