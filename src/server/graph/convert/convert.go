package convert

import (
	pb "thaily/proto/common"

	"thaily/src/server/graph/model"
)

func ConvertSearchRequestToPB(input model.SearchRequestInput) *pb.SearchRequest {
	if input.Pagination == nil && (input.Filters == nil || len(input.Filters) == 0) {
		return nil
	}

	req := &pb.SearchRequest{}

	if input.Pagination != nil {
		req.Pagination = convertPaginationToPB(input.Pagination)
	}

	if input.Filters != nil && len(input.Filters) > 0 {
		req.Filters = make([]*pb.FilterCriteria, 0, len(input.Filters))
		for _, filter := range input.Filters {
			if filter != nil {
				req.Filters = append(req.Filters, convertFilterCriteriaToPB(filter))
			}
		}
	}

	return req
}

func convertPaginationToPB(input *model.PaginationInput) *pb.Pagination {
	if input == nil {
		return nil
	}

	pagination := &pb.Pagination{}

	if input.Page != nil {
		pagination.Page = *input.Page
	}
	if input.PageSize != nil {
		pagination.PageSize = *input.PageSize
	}
	if input.SortBy != nil {
		pagination.SortBy = *input.SortBy
	}
	if input.Descending != nil {
		pagination.Descending = *input.Descending
	}

	return pagination
}

func convertFilterCriteriaToPB(input *model.FilterCriteriaInput) *pb.FilterCriteria {
	if input == nil {
		return nil
	}

	criteria := &pb.FilterCriteria{}

	if input.Condition != nil {
		criteria.Criteria = &pb.FilterCriteria_Condition{
			Condition: convertFilterConditionToPB(input.Condition),
		}
	} else if input.Group != nil {
		criteria.Criteria = &pb.FilterCriteria_Group{
			Group: convertFilterGroupToPB(input.Group),
		}
	}

	return criteria
}

func convertFilterConditionToPB(input *model.FilterConditionInput) *pb.FilterCondition {
	if input == nil {
		return nil
	}

	return &pb.FilterCondition{
		Field:    input.Field,
		Operator: convertFilterOperatorToPB(input.Operator),
		Values:   input.Values,
	}
}

func convertFilterGroupToPB(input *model.FilterGroupInput) *pb.FilterGroup {
	if input == nil {
		return nil
	}

	group := &pb.FilterGroup{}

	if input.Logic != nil {
		group.Logic = convertLogicalConditionToPB(*input.Logic)
	}

	if input.Filters != nil && len(input.Filters) > 0 {
		group.Filters = make([]*pb.FilterCriteria, 0, len(input.Filters))
		for _, filter := range input.Filters {
			if filter != nil {
				group.Filters = append(group.Filters, convertFilterCriteriaToPB(filter))
			}
		}
	}

	return group
}

func convertFilterOperatorToPB(op model.FilterOperator) pb.FilterOperator {
	switch op {
	case model.FilterOperatorEqual:
		return pb.FilterOperator_EQUAL
	case model.FilterOperatorNotEqual:
		return pb.FilterOperator_NOT_EQUAL
	case model.FilterOperatorGreaterThan:
		return pb.FilterOperator_GREATER_THAN
	case model.FilterOperatorGreaterThanEqual:
		return pb.FilterOperator_GREATER_THAN_EQUAL
	case model.FilterOperatorLessThan:
		return pb.FilterOperator_LESS_THAN
	case model.FilterOperatorLessThanEqual:
		return pb.FilterOperator_LESS_THAN_EQUAL
	case model.FilterOperatorLike:
		return pb.FilterOperator_LIKE
	case model.FilterOperatorIn:
		return pb.FilterOperator_IN
	case model.FilterOperatorNotIn:
		return pb.FilterOperator_NOT_IN
	case model.FilterOperatorIsNull:
		return pb.FilterOperator_IS_NULL
	case model.FilterOperatorIsNotNull:
		return pb.FilterOperator_IS_NOT_NULL
	case model.FilterOperatorBetween:
		return pb.FilterOperator_BETWEEN
	default:
		return pb.FilterOperator_EQUAL
	}
}

func convertLogicalConditionToPB(cond model.LogicalCondition) pb.LogicalCondition {
	switch cond {
	case model.LogicalConditionAnd:
		return pb.LogicalCondition_AND
	case model.LogicalConditionOr:
		return pb.LogicalCondition_OR
	default:
		return pb.LogicalCondition_AND
	}
}
