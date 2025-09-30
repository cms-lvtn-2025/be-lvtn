package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"thaily/src/graph/model"
	"thaily/src/helper"

	"cloud.google.com/go/auth/credentials/idtoken"
)

func (c *Controller) Login(ctx context.Context, input model.Login) (*model.CommonResponse, error) {
	payload, err := idtoken.Validate(ctx, input.Token, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		return nil, err
	}
	email, emailOk := payload.Claims["email"].(string)
	if !emailOk {
		return nil, fmt.Errorf("email not found in token")
	}

	procedure := fmt.Sprintf("CALL login('%s', '%s', '%s');", email, input.Semester, input.Role)

	resp, err := c.common.CallProcerdure(ctx, procedure)
	if err != nil {
		return nil, err
	}

	// Direct mapping - structpb.Struct tự động convert sang map[string]interface{}
	var data map[string]interface{}
	if resp.Data != nil {
		data = resp.Data.AsMap()

		// Lấy thông tin user từ database response
		if resultsSlice, ok := data["results"].([]interface{}); ok && len(resultsSlice) > 0 {
			if results, ok := resultsSlice[0].(map[string]interface{}); ok {
				// Generate custom JWT
				userID := ""
				if id, ok := results["id"].(string); ok {
					userID = id
				}

				name, _ := payload.Claims["name"].(string)

				// Parse permissions field if it's a JSON string
				if permissionsStr, ok := results["permissions"].(string); ok && permissionsStr != "" {
					var permissions []string
					if err := json.Unmarshal([]byte(permissionsStr), &permissions); err == nil {
						results["permissions"] = permissions
					}
				}

				// Generate JWT với thông tin từ DB
				token, err := helper.GenerateJWT(userID, email, name, string(input.Role), input.Semester)
				if err != nil {
					return nil, fmt.Errorf("failed to generate token: %v", err)
				}

				data["token"] = token

				data["results"] = results
			}
		}
	}

	return &model.CommonResponse{
		Success: resp.Success,
		Message: resp.Message,
		Data:    data,
	}, nil
}

func (c *Controller) Info(ctx context.Context, input model.Info) (*model.CommonResponse, error) {
	authValue := ctx.Value(helper.Auth)
	if authValue == nil {
		return &model.CommonResponse{
			Success: false,
			Message: "Unauthorized: No auth info",
			Data:    nil,
		}, nil
	}
	claims, ok := authValue.(*helper.Claims)
	if !ok {
		return &model.CommonResponse{
			Success: false,
			Message: "Unauthorized: Invalid auth info",
			Data:    nil,
		}, nil
	}

	procedure := fmt.Sprintf("CALL login('%s', '%s', '%s');", claims.Email, input.Semester, claims.Role)

	resp, err := c.common.CallProcerdure(ctx, procedure)
	if err != nil {
		return nil, err
	}

	// Direct mapping - structpb.Struct tự động convert sang map[string]interface{}
	var data map[string]interface{}
	if resp.Data != nil {
		data = resp.Data.AsMap()

		// Lấy thông tin user từ database response
		if resultsSlice, ok := data["results"].([]interface{}); ok && len(resultsSlice) > 0 {
			if results, ok := resultsSlice[0].(map[string]interface{}); ok {
				// Parse permissions field if it's a JSON string
				if permissionsStr, ok := results["permissions"].(string); ok && permissionsStr != "" {
					var permissions []string
					if err := json.Unmarshal([]byte(permissionsStr), &permissions); err == nil {
						results["permissions"] = permissions
					}
				}
				data["results"] = results
			}
		}
	}

	return &model.CommonResponse{
		Success: resp.Success,
		Message: resp.Message,
		Data:    data,
	}, nil
}

func (c *Controller) GetTopic(ctx context.Context, input model.GetTopicInput) (*model.CommonResponse, error) {
	// Validate JWT from context
	authValue := ctx.Value(helper.Auth)
	if authValue == nil {
		return &model.CommonResponse{
			Success: false,
			Message: "Unauthorized: No auth info",
			Data:    nil,
		}, nil
	}
	claims, ok := authValue.(*helper.Claims)
	if !ok {
		return &model.CommonResponse{
			Success: false,
			Message: "Unauthorized: Invalid auth info",
			Data:    nil,
		}, nil
	}

	procedure := fmt.Sprintf("CALL getTopic('%s', '%s', '%s', %s);", claims.UserID, input.Semester, claims.Role, "NULL")
	if input.Status != nil {
		procedure = fmt.Sprintf("CALL getTopic('%s', '%s', '%s', '%s');", claims.UserID, input.Semester, claims.Role, *input.Status)
	}
	resp, err := c.common.CallProcerdure(ctx, procedure)
	if err != nil {
		return nil, err
	}

	// Direct mapping - structpb.Struct tự động convert sang map[string]interface{}
	var data map[string]interface{}
	if resp.Data != nil {
		data = resp.Data.AsMap()
	}

	return &model.CommonResponse{
		Success: resp.Success,
		Message: resp.Message,
		Data:    data,
	}, nil
}

func (c *Controller) Semester(ctx context.Context) (*model.CommonResponse, error) {
	procedure := "CALL semesters();"

	resp, err := c.common.CallProcerdure(ctx, procedure)
	if err != nil {
		return nil, err
	}

	// Direct mapping - structpb.Struct tự động convert sang map[string]interface{}
	var data map[string]interface{}
	if resp.Data != nil {
		data = resp.Data.AsMap()
	}

	return &model.CommonResponse{
		Success: resp.Success,
		Message: resp.Message,
		Data:    data,
	}, nil
}

func (c *Controller) GetCouncil(ctx context.Context, input model.GetCouncilInput) (*model.CommonResponse, error) {
	procedure := fmt.Sprintf("CALL getCouncil('%s');", input.Semester)

	resp, err := c.common.CallProcerdure(ctx, procedure)
	if err != nil {
		return nil, err
	}

	// Direct mapping - structpb.Struct tự động convert sang map[string]interface{}
	var data map[string]interface{}
	if resp.Data != nil {
		data = resp.Data.AsMap()
	}

	return &model.CommonResponse{
		Success: resp.Success,
		Message: resp.Message,
		Data:    data,
	}, nil
}
