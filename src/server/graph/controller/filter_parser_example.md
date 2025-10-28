# Filter Parser - Hướng dẫn sử dụng

## ParseFilterQuery Function

Function `ParseFilterQuery` cho phép parse từ JSON string sang `pb.SearchRequest`.

## Format cơ bản

```json
{
  "page": 1,
  "pageSize": 10,
  "sortBy": "created_at",
  "descending": true,
  "filters": {
    "and": [
      {"field": "status", "op": "eq", "values": ["approved"]}
    ]
  }
}
```

## Các toán tử (operators) hỗ trợ

| Operator | Aliases | Ý nghĩa |
|----------|---------|---------|
| `eq` | `=`, `equal` | Bằng |
| `ne` | `!=`, `not_equal` | Khác |
| `gt` | `>`, `greater_than` | Lớn hơn |
| `gte` | `>=`, `greater_than_equal` | Lớn hơn hoặc bằng |
| `lt` | `<`, `less_than` | Nhỏ hơn |
| `lte` | `<=`, `less_than_equal` | Nhỏ hơn hoặc bằng |
| `like` | | LIKE (tìm kiếm text) |
| `in` | | IN (trong danh sách) |
| `not_in` | `nin` | NOT IN |
| `is_null` | `null` | IS NULL |
| `is_not_null` | `not_null` | IS NOT NULL |
| `between` | | BETWEEN |

## Ví dụ sử dụng

### 1. Filter đơn giản với AND

```json
{
  "page": 1,
  "pageSize": 20,
  "sortBy": "created_at",
  "descending": true,
  "filters": {
    "and": [
      {"field": "status", "op": "eq", "values": ["approved"]},
      {"field": "major_code", "op": "eq", "values": ["MAJ_DTVT_VT"]}
    ]
  }
}
```

SQL tương đương:
```sql
WHERE status = 'approved' AND major_code = 'MAJ_DTVT_VT'
ORDER BY created_at DESC
LIMIT 20 OFFSET 0
```

### 2. Filter với OR

```json
{
  "filters": {
    "or": [
      {"field": "status", "op": "eq", "values": ["approved"]},
      {"field": "status", "op": "eq", "values": ["in_progress"]}
    ]
  }
}
```

SQL tương đương:
```sql
WHERE status = 'approved' OR status = 'in_progress'
```

### 3. Filter với IN operator

```json
{
  "filters": {
    "and": [
      {"field": "major_code", "op": "in", "values": ["MAJ1", "MAJ2", "MAJ3"]}
    ]
  }
}
```

SQL tương đương:
```sql
WHERE major_code IN ('MAJ1', 'MAJ2', 'MAJ3')
```

### 4. Filter với LIKE

```json
{
  "filters": {
    "and": [
      {"field": "title", "op": "like", "values": ["test"]}
    ]
  }
}
```

SQL tương đương:
```sql
WHERE title LIKE '%test%'
```

### 5. Filter với BETWEEN

```json
{
  "filters": {
    "and": [
      {"field": "created_at", "op": "between", "values": ["2024-01-01", "2024-12-31"]}
    ]
  }
}
```

SQL tương đương:
```sql
WHERE created_at BETWEEN '2024-01-01' AND '2024-12-31'
```

### 6. Nested filter (AND và OR kết hợp)

```json
{
  "filters": {
    "and": [
      {"field": "semester_code", "op": "eq", "values": ["SEM_2025_1"]},
      {
        "or": [
          {"field": "status", "op": "eq", "values": ["approved"]},
          {"field": "status", "op": "eq", "values": ["in_progress"]}
        ]
      }
    ]
  }
}
```

SQL tương đương:
```sql
WHERE semester_code = 'SEM_2025_1'
  AND (status = 'approved' OR status = 'in_progress')
```

### 7. Nested filter phức tạp

```json
{
  "filters": {
    "and": [
      {"field": "semester_code", "op": "eq", "values": ["SEM_2025_1"]},
      {
        "or": [
          {
            "and": [
              {"field": "status", "op": "eq", "values": ["approved"]},
              {"field": "major_code", "op": "in", "values": ["MAJ1", "MAJ2"]}
            ]
          },
          {"field": "teacher_supervisor_code", "op": "eq", "values": ["TCH_00006"]}
        ]
      }
    ]
  }
}
```

SQL tương đương:
```sql
WHERE semester_code = 'SEM_2025_1'
  AND (
    (status = 'approved' AND major_code IN ('MAJ1', 'MAJ2'))
    OR teacher_supervisor_code = 'TCH_00006'
  )
```

### 8. Filter với NULL check

```json
{
  "filters": {
    "and": [
      {"field": "updated_by", "op": "is_not_null", "values": []}
    ]
  }
}
```

SQL tương đương:
```sql
WHERE updated_by IS NOT NULL
```

## Cách sử dụng trong code

```go
import (
	"fmt"
	"thaily/src/graph/controller"
)

func main() {
	queryStr := `{
		"page": 1,
		"pageSize": 10,
		"sortBy": "created_at",
		"descending": true,
		"filters": {
			"and": [
				{"field": "status", "op": "eq", "values": ["approved"]},
				{"field": "semester_code", "op": "eq", "values": ["SEM_2025_1"]}
			]
		}
	}`

	searchRequest, err := controller.ParseFilterQuery(queryStr)
	if err != nil {
		fmt.Printf("Error parsing filter: %v\n", err)
		return
	}

	// Use searchRequest in your gRPC call
	// topics, err := client.ListTopics(ctx, &pb.ListTopicsRequest{
	// 	Search: searchRequest,
	// })
}
```

## Lưu ý

1. **Pagination mặc định**: Nếu không cung cấp, sẽ sử dụng:
   - `page`: 1
   - `pageSize`: 10
   - `sortBy`: "created_at"
   - `descending`: true

2. **Filters phải bắt đầu với "and" hoặc "or"**:
   ```json
   "filters": {
     "and": [...]  // ✅ Đúng
   }
   ```

   ```json
   "filters": [...]  // ❌ Sai
   ```

3. **Values luôn là array**: Ngay cả khi chỉ có 1 giá trị
   ```json
   {"field": "status", "op": "eq", "values": ["approved"]}  // ✅ Đúng
   {"field": "status", "op": "eq", "values": "approved"}    // ❌ Sai
   ```

4. **Nested filter không giới hạn độ sâu**: Có thể nest nhiều tầng tùy ý

5. **Empty query**: Nếu truyền empty string `""`, function sẽ trả về empty SearchRequest với default pagination
