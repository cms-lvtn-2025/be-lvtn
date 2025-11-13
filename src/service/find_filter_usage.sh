#!/bin/bash

echo "========================================"
echo "Finding Filter Usage in Handlers"
echo "========================================"
echo ""

# Find all files using the old pattern
echo "Files using OLD pattern (GetCondition check):"
echo "--------------------------------------"
grep -r "filter.GetCondition()" --include="*.go" \
    thesis/handler/ academic/handler/ council/handler/ \
    user/handler/ file/handler/ role/handler/ 2>/dev/null | \
    cut -d: -f1 | sort -u | while read file; do
    echo "  - $file"
done

echo ""
echo "Files using NEW pattern (BuildWhereClause):"
echo "--------------------------------------"
grep -r "BuildWhereClause" --include="*.go" \
    thesis/handler/ academic/handler/ council/handler/ \
    user/handler/ file/handler/ role/handler/ 2>/dev/null | \
    cut -d: -f1 | sort -u | while read file; do
    echo "  ✅ $file"
done

echo ""
echo "========================================"
echo "Summary"
echo "========================================"

OLD_COUNT=$(grep -r "filter.GetCondition()" --include="*.go" \
    thesis/handler/ academic/handler/ council/handler/ \
    user/handler/ file/handler/ role/handler/ 2>/dev/null | \
    cut -d: -f1 | sort -u | wc -l)

NEW_COUNT=$(grep -r "BuildWhereClause" --include="*.go" \
    thesis/handler/ academic/handler/ council/handler/ \
    user/handler/ file/handler/ role/handler/ 2>/dev/null | \
    cut -d: -f1 | sort -u | wc -l)

echo "Files using old pattern: $OLD_COUNT"
echo "Files using new pattern: $NEW_COUNT"
echo ""

if [ "$OLD_COUNT" -gt 0 ]; then
    echo "⚠️  Migration needed for $OLD_COUNT files"
    echo ""
    echo "To update a file, replace:"
    echo ""
    echo "  OLD:"
    echo "    if filter.GetCondition() != nil {"
    echo "        condition := filter.GetCondition()"
    echo "        // ..."
    echo "        helper.BuildFilterCondition(condition, &args)"
    echo "    }"
    echo ""
    echo "  NEW:"
    echo "    whereClause = helper.BuildWhereClause(req.Search.Filters, &args, whiteMap)"
    echo ""
else
    echo "✅ All files updated!"
fi
