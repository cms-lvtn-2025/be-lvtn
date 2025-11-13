#!/bin/bash

# Script to update all handlers to use new BuildWhereClause function
# This ensures all services can use the new nested filter functionality

set -e

echo "========================================"
echo "Filter Update Script"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Find all handler files
HANDLER_FILES=$(find . -path "*/handler/*.go" -type f | grep -E "(thesis|academic|council|user|file|role)")

echo "Found handler files:"
echo "$HANDLER_FILES" | wc -l
echo ""

UPDATED_COUNT=0
SKIPPED_COUNT=0
ERROR_COUNT=0

# Pattern to find and replace
OLD_PATTERN='if req.Search != nil && len(req.Search.Filters) > 0 {
		whereConditions := \[\]string{}
		for _, filter := range req.Search.Filters {
			if filter.GetCondition\(\) != nil {
				condition := filter.GetCondition\(\)
				if _, ok := whiteMap\[condition.Field\]; !ok {
					continue
				}
				whereConditions = append\(whereConditions, helper.BuildFilterCondition\(condition, &args\)\)
			}
		}
		if len\(whereConditions\) > 0 {
			whereClause = "WHERE " \+ strings.Join\(whereConditions, " AND "\)
		}
	}'

for file in $HANDLER_FILES; do
    echo -n "Checking $file... "

    # Check if file contains old pattern
    if grep -q "BuildFilterCondition" "$file" 2>/dev/null; then
        if grep -q "BuildWhereClause" "$file" 2>/dev/null; then
            echo -e "${GREEN}[ALREADY UPDATED]${NC}"
            ((SKIPPED_COUNT++))
        else
            echo -e "${YELLOW}[NEEDS UPDATE]${NC}"
            echo "  File uses old BuildFilterCondition pattern"
            echo "  Please manually review and update: $file"
            ((ERROR_COUNT++))
        fi
    else
        echo -e "${GREEN}[OK - Not using filters or already updated]${NC}"
        ((SKIPPED_COUNT++))
    fi
done

echo ""
echo "========================================"
echo "Summary"
echo "========================================"
echo -e "Updated:       ${GREEN}${UPDATED_COUNT}${NC}"
echo -e "Already OK:    ${GREEN}${SKIPPED_COUNT}${NC}"
echo -e "Needs Review:  ${YELLOW}${ERROR_COUNT}${NC}"
echo ""

if [ $ERROR_COUNT -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Some files need manual review${NC}"
    echo ""
    echo "To update manually, replace this pattern:"
    echo ""
    echo "OLD CODE:"
    echo '```go'
    echo 'if req.Search != nil && len(req.Search.Filters) > 0 {'
    echo '    whereConditions := []string{}'
    echo '    for _, filter := range req.Search.Filters {'
    echo '        if filter.GetCondition() != nil {'
    echo '            condition := filter.GetCondition()'
    echo '            if _, ok := whiteMap[condition.Field]; !ok {'
    echo '                continue'
    echo '            }'
    echo '            whereConditions = append(whereConditions, helper.BuildFilterCondition(condition, &args))'
    echo '        }'
    echo '    }'
    echo '    if len(whereConditions) > 0 {'
    echo '        whereClause = "WHERE " + strings.Join(whereConditions, " AND ")'
    echo '    }'
    echo '}'
    echo '```'
    echo ""
    echo "NEW CODE:"
    echo '```go'
    echo 'if req.Search != nil && len(req.Search.Filters) > 0 {'
    echo '    whereClause = helper.BuildWhereClause(req.Search.Filters, &args, whiteMap)'
    echo '}'
    echo '```'
    echo ""
else
    echo -e "${GREEN}✅ All files are ready!${NC}"
fi

echo ""
echo "For detailed migration guide, see:"
echo "  pkg/helper/FILTER_MIGRATION_GUIDE.md"
echo ""
