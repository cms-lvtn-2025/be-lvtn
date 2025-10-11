package controller

import (
	pb "thaily/proto/common"
	"thaily/src/graph/model"
	"thaily/src/server/client"
)

type Controller struct {
	academic *client.GRPCAcadamicClient
	council  *client.GRPCCouncil
	file     *client.GRPCfile
	role     *client.GRPCRole
	thesis   *client.GRPCthesis
	user     *client.GRPCUser
}

// Constructor function
func NewController(academic *client.GRPCAcadamicClient, council *client.GRPCCouncil, file *client.GRPCfile, role *client.GRPCRole, thesis *client.GRPCthesis, user *client.GRPCUser) *Controller {
	return &Controller{
		academic: academic,
		council:  council,
		file:     file,
		role:     role,
		thesis:   thesis,
		user:     user,
	}
}

// ConvertSearchRequestToPB converts GraphQL SearchRequestInput to Protobuf SearchRequest
func (c *Controller) ConvertSearchRequestToPB(input model.SearchRequestInput) *pb.SearchRequest {
	if input.Pagination == nil && (input.Filters == nil || len(input.Filters) == 0) {
		return nil
	}

	req := &pb.SearchRequest{}

	// Convert Pagination
	if input.Pagination != nil {
		req.Pagination = convertPaginationToPB(input.Pagination)
	}

	// Convert Filters
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

// convertPaginationToPB converts GraphQL PaginationInput to Protobuf Pagination
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

// convertFilterCriteriaToPB converts GraphQL FilterCriteriaInput to Protobuf FilterCriteria
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

// convertFilterConditionToPB converts GraphQL FilterConditionInput to Protobuf FilterCondition
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

// convertFilterGroupToPB converts GraphQL FilterGroupInput to Protobuf FilterGroup
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

// convertFilterOperatorToPB converts GraphQL FilterOperator to Protobuf FilterOperator
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

// convertLogicalConditionToPB converts GraphQL LogicalCondition to Protobuf LogicalCondition
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
