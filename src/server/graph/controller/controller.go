package controller

import (
	"context"
	"fmt"
	"strings"
	pb "thaily/proto/common"
	"thaily/src/server/client"
	"thaily/src/server/graph/helper"
	"thaily/src/server/graph/model"

	"github.com/golang-jwt/jwt/v5"
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

func (c *Controller) DefaultPagination() *model.PaginationInput {
	Page := int32(1)
	PageSize := int32(10)
	SortBy := "created_at"
	Descending := true
	return &model.PaginationInput{
		Page:       &Page,
		PageSize:   &PageSize,
		SortBy:     &SortBy,
		Descending: &Descending,
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

func (c *Controller) GetInfoRequest(ctx context.Context) (id *string, role *string, err error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("not authorized")
	}
	roleSystem, ok := claims["role"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("not authorized")
	}
	semester, ok := ctx.Value("semester").(string)

	idsArr := strings.Split(claims["ids"].(string), ",")
	fmt.Println(idsArr)
	myId := ""
	if semester == "" {
		myId = strings.Split(idsArr[0], "-")[1]
	} else {
		for _, id := range idsArr {
			if strings.HasPrefix(id, semester+"-") {
				myId = strings.Split(id, "-")[1]
			}
		}
	}
	fmt.Println(myId)
	if myId == "" {
		return nil, nil, fmt.Errorf("no teacher found for semester %s", semester)
	}

	return &myId, &roleSystem, nil
}

func (c *Controller) GetInfoAllRequest(ctx context.Context) (ids *[]string, semesters *[]string, role *string, err error) {
	claims, ok := ctx.Value(helper.Auth).(jwt.MapClaims)
	if !ok {
		return nil, nil, nil, fmt.Errorf("not authorized")
	}
	roleSystem, ok := claims["role"].(string)
	if !ok {
		return nil, nil, nil, fmt.Errorf("not authorized")
	}
	idsArr := strings.Split(claims["ids"].(string), ",")
	var myIds []string
	var mySemesters []string
	for _, id := range idsArr {
		parts := strings.Split(id, "-")
		if len(parts) == 2 {
			myIds = append(myIds, strings.Split(id, "-")[1])
			mySemesters = append(mySemesters, strings.Split(id, "-")[0])
		}

	}
	if len(myIds) == 0 {
		return nil, nil, nil, fmt.Errorf("no id found for semester %s", roleSystem)
	}
	return &myIds, &mySemesters, &roleSystem, nil

}
