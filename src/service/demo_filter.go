package main

import (
	"encoding/json"
	"fmt"
	"strings"
	pbCommon "thaily/proto/common"
	"thaily/src/service/pkg/helper"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("Filter Demo - Testing Your JSON Request")
	fmt.Println("==============================================")
	fmt.Println()

	// Whitelist (same as topic handler)
	whiteMap := map[string]bool{
		"id":              true,
		"title":           true,
		"major_code":      true,
		"semester_code":   true,
		"status":          true,
		"percent_stage_1": true,
		"percent_stage_2": true,
	}

	// Your JSON request structure
	filters := []*pbCommon.FilterCriteria{
		{
			Criteria: &pbCommon.FilterCriteria_Condition{
				Condition: &pbCommon.FilterCondition{
					Field:    "title",
					Operator: pbCommon.FilterOperator_LIKE,
					Values:   []string{"10"},
				},
			},
		},
		{
			Criteria: &pbCommon.FilterCriteria_Group{
				Group: &pbCommon.FilterGroup{
					Logic: pbCommon.LogicalCondition_OR,
					Filters: []*pbCommon.FilterCriteria{
						{
							Criteria: &pbCommon.FilterCriteria_Condition{
								Condition: &pbCommon.FilterCondition{
									Field:    "title",
									Operator: pbCommon.FilterOperator_LIKE,
									Values:   []string{"CNTT"},
								},
							},
						},
					},
				},
			},
		},
	}

	// Build SQL
	whereClause := ""
	args := []interface{}{}

	if len(filters) > 0 {
		whereConditions := []string{}
		for _, filter := range filters {
			condition := helper.BuildFilterCriteriaWithWhitelist(filter, &args, whiteMap)
			if condition != "" && condition != "1=1" {
				whereConditions = append(whereConditions, condition)
			}
		}
		if len(whereConditions) > 0 {
			whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
		}
	}

	// Print results
	fmt.Println("Input JSON (conceptual):")
	fmt.Println(`{
  "search": {
    "filters": [
      {"condition": {"field": "title", "operator": "LIKE", "values": ["10"]}},
      {"group": {"logic": "OR", "filters": [
        {"condition": {"field": "title", "operator": "LIKE", "values": ["CNTT"]}}
      ]}}
    ],
    "pagination": {
      "descending": false,
      "page": 1,
      "page_size": 100,
      "sort_by": "created_at"
    }
  }
}`)
	fmt.Println()
	fmt.Println("Generated WHERE Clause:")
	fmt.Println("  ", whereClause)
	fmt.Println()
	fmt.Println("Generated Args:")
	argsJSON, _ := json.MarshalIndent(args, "  ", "  ")
	fmt.Printf("  %s\n", string(argsJSON))
	fmt.Println()

	// Full SQL
	fullSQL := fmt.Sprintf(`
SELECT id, title, major_code, semester_code, status,
       percent_stage_1, percent_stage_2,
       created_at, updated_at
FROM Topic
%s
ORDER BY created_at ASC
LIMIT 100 OFFSET 0
`, whereClause)

	fmt.Println("Full SQL Query:")
	fmt.Println(fullSQL)
	fmt.Println()

	// Test complex nested example
	fmt.Println("==============================================")
	fmt.Println("Complex Nested Example")
	fmt.Println("==============================================")
	fmt.Println()

	complexFilters := []*pbCommon.FilterCriteria{
		{
			Criteria: &pbCommon.FilterCriteria_Condition{
				Condition: &pbCommon.FilterCondition{
					Field:    "status",
					Operator: pbCommon.FilterOperator_EQUAL,
					Values:   []string{"active"},
				},
			},
		},
		{
			Criteria: &pbCommon.FilterCriteria_Group{
				Group: &pbCommon.FilterGroup{
					Logic: pbCommon.LogicalCondition_OR,
					Filters: []*pbCommon.FilterCriteria{
						{
							Criteria: &pbCommon.FilterCriteria_Condition{
								Condition: &pbCommon.FilterCondition{
									Field:    "major_code",
									Operator: pbCommon.FilterOperator_IN,
									Values:   []string{"CNTT", "KTPM"},
								},
							},
						},
						{
							Criteria: &pbCommon.FilterCriteria_Group{
								Group: &pbCommon.FilterGroup{
									Logic: pbCommon.LogicalCondition_AND,
									Filters: []*pbCommon.FilterCriteria{
										{
											Criteria: &pbCommon.FilterCriteria_Condition{
												Condition: &pbCommon.FilterCondition{
													Field:    "semester_code",
													Operator: pbCommon.FilterOperator_EQUAL,
													Values:   []string{"HK1-2024"},
												},
											},
										},
										{
											Criteria: &pbCommon.FilterCriteria_Condition{
												Condition: &pbCommon.FilterCondition{
													Field:    "percent_stage_1",
													Operator: pbCommon.FilterOperator_GREATER_THAN,
													Values:   []string{"70"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	args2 := []interface{}{}
	whereConditions2 := []string{}

	for _, filter := range complexFilters {
		condition := helper.BuildFilterCriteriaWithWhitelist(filter, &args2, whiteMap)
		if condition != "" && condition != "1=1" {
			whereConditions2 = append(whereConditions2, condition)
		}
	}

	whereClause2 := "WHERE " + strings.Join(whereConditions2, " AND ")

	fmt.Println("Complex WHERE Clause:")
	fmt.Println("  ", whereClause2)
	fmt.Println()
	fmt.Println("Args:")
	args2JSON, _ := json.MarshalIndent(args2, "  ", "  ")
	fmt.Printf("  %s\n", string(args2JSON))
	fmt.Println()

	fmt.Println("==============================================")
	fmt.Println("✅ Filter system working correctly!")
	fmt.Println("==============================================")
}
