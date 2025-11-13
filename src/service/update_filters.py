#!/usr/bin/env python3

import os
import re
import shutil
from datetime import datetime

# Files to update
FILES = [
    "academic/handler/major.go",
    "academic/handler/semester.go",
    "council/handler/council.go",
    "council/handler/defence.go",
    "council/handler/gradedefencecriterion.go",
    "council/handler/gradedefence.go",
    "file/handler/file.go",
    "role/handler/rolesystem.go",
    "thesis/handler/enrollment.go",
    "thesis/handler/final.go",
    "thesis/handler/gradereview.go",
    "thesis/handler/midterm.go",
    "thesis/handler/topiccouncil.go",
    "thesis/handler/topiccouncilsupervisor.go",
    "user/handler/student.go",
    "user/handler/teacher.go",
]

# Old pattern (regex)
OLD_PATTERN = r'''(\t)if req\.Search != nil && len\(req\.Search\.Filters\) > 0 \{\n\t\twhereConditions := \[\]string\{\}\n\t\tfor _, filter := range req\.Search\.Filters \{\n\t\t\tif filter\.GetCondition\(\) != nil \{\n\t\t\t\tcondition := filter\.GetCondition\(\)\n\t\t\t\tif _, ok := whiteMap\[condition\.Field\]; !ok \{\n\t\t\t\t\tcontinue\n\t\t\t\t\}\n\t\t\t\twhereConditions = append\(whereConditions, helper\.BuildFilterCondition\(condition, &args\)\)\n\t\t\t\}\n\t\t\}\n\t\tif len\(whereConditions\) > 0 \{\n\t\t\twhereClause = "WHERE " \+ strings\.Join\(whereConditions, " AND "\)\n\t\t\}\n\t\}'''

# New pattern
NEW_PATTERN = r'''\1if req.Search != nil && len(req.Search.Filters) > 0 {
\t\twhereClause = helper.BuildWhereClause(req.Search.Filters, &args, whiteMap)
\t}'''

def main():
    print("=" * 60)
    print("Auto-Update Filter Pattern to BuildWhereClause")
    print("=" * 60)
    print()

    # Create backup directory
    backup_dir = f"backup_filters_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
    os.makedirs(backup_dir, exist_ok=True)
    print(f"Backup directory: {backup_dir}/")
    print()

    updated_count = 0
    skipped_count = 0
    error_count = 0

    for filepath in FILES:
        if not os.path.exists(filepath):
            print(f"⚠️  File not found: {filepath}")
            error_count += 1
            continue

        print(f"Processing: {filepath}")

        # Read file
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()

        # Check if already updated
        if 'BuildWhereClause' in content:
            print(f"  ✅ Already updated (skipped)")
            skipped_count += 1
            continue

        # Backup
        backup_path = os.path.join(backup_dir, os.path.basename(filepath))
        shutil.copy2(filepath, backup_path)

        # Replace pattern using simpler string replacement
        old_block = '''	if req.Search != nil && len(req.Search.Filters) > 0 {
		whereConditions := []string{}
		for _, filter := range req.Search.Filters {
			if filter.GetCondition() != nil {
				condition := filter.GetCondition()
				if _, ok := whiteMap[condition.Field]; !ok {
					continue
				}
				whereConditions = append(whereConditions, helper.BuildFilterCondition(condition, &args))
			}
		}
		if len(whereConditions) > 0 {
			whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
		}
	}'''

        new_block = '''	if req.Search != nil && len(req.Search.Filters) > 0 {
		whereClause = helper.BuildWhereClause(req.Search.Filters, &args, whiteMap)
	}'''

        if old_block in content:
            new_content = content.replace(old_block, new_block)

            # Write updated content
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(new_content)

            print(f"  ✅ Updated")
            updated_count += 1
        else:
            print(f"  ⚠️  Pattern not found (needs manual review)")
            error_count += 1

    print()
    print("=" * 60)
    print("Summary")
    print("=" * 60)
    print(f"Updated:       {updated_count}")
    print(f"Already OK:    {skipped_count}")
    print(f"Needs Review:  {error_count}")
    print()
    print(f"Backups saved to: {backup_dir}/")
    print()

    if updated_count > 0:
        print("✅ Update complete!")
        print()
        print("Next steps:")
        print("  1. Verify changes: ./find_filter_usage.sh")
        print("  2. Run tests: go test ./...")
        print("  3. Rebuild all services")
        print()
    else:
        print("⚠️  No files were updated")

if __name__ == "__main__":
    main()
