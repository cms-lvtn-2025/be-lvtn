#!/bin/bash

# Auto-update all handler files to use BuildWhereClause
# Backup files before updating

set -e

echo "========================================"
echo "Auto-Update Filter Pattern"
echo "========================================"
echo ""

# Files to update
FILES=(
    "academic/handler/faculty.go"
    "academic/handler/major.go"
    "academic/handler/semester.go"
    "council/handler/council.go"
    "council/handler/defence.go"
    "council/handler/gradedefencecriterion.go"
    "council/handler/gradedefence.go"
    "file/handler/file.go"
    "role/handler/rolesystem.go"
    "thesis/handler/enrollment.go"
    "thesis/handler/final.go"
    "thesis/handler/gradereview.go"
    "thesis/handler/midterm.go"
    "thesis/handler/topiccouncil.go"
    "thesis/handler/topiccouncilsupervisor.go"
    "user/handler/student.go"
    "user/handler/teacher.go"
)

BACKUP_DIR="backup_filters_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

echo "Creating backups in: $BACKUP_DIR/"
echo ""

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "Processing: $file"

        # Backup
        cp "$file" "$BACKUP_DIR/$(basename $file)"

        # Create temp file for the update
        temp_file="${file}.tmp"

        # Use awk to replace the multi-line pattern
        awk '
        /if req\.Search != nil && len\(req\.Search\.Filters\) > 0 {/ {
            # Check if this is the old pattern
            getline next_line
            if (next_line ~ /whereConditions := \[\]string{}/) {
                # Found old pattern, skip until closing brace and replace
                in_old_block = 1
                brace_count = 1
                print "\tif req.Search != nil && len(req.Search.Filters) > 0 {"
                print "\t\twhereClause = helper.BuildWhereClause(req.Search.Filters, &args, whiteMap)"
                next
            } else {
                # Not old pattern, keep as is
                print
                print next_line
                next
            }
        }

        in_old_block && /{/ { brace_count++ }
        in_old_block && /}/ {
            brace_count--
            if (brace_count == 0) {
                print "\t}"
                in_old_block = 0
                next
            }
        }

        !in_old_block { print }
        ' "$file" > "$temp_file"

        # Replace original file
        mv "$temp_file" "$file"

        echo "  ✅ Updated"
    else
        echo "  ⚠️  File not found: $file"
    fi
done

echo ""
echo "========================================"
echo "Update Complete!"
echo "========================================"
echo ""
echo "Backups saved to: $BACKUP_DIR/"
echo ""
echo "To verify changes:"
echo "  ./find_filter_usage.sh"
echo ""
echo "To restore from backup (if needed):"
echo "  cp $BACKUP_DIR/*.go <original_location>/"
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff"
echo "  2. Run tests: go test ./..."
echo "  3. Rebuild services"
echo ""
