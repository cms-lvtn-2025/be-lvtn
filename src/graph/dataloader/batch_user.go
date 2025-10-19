package dataloader

import (
	"context"
	"log"
	"thaily/src/graph/convert"
	"thaily/src/graph/model"
	"thaily/src/server/client"
)

func createTeacherInfoBatchFunc(client *client.GRPCUser) BatchFunc[string, *model.StudentTeacherInfo] {
	return func(ctx context.Context, ids []string) (map[string]*model.StudentTeacherInfo, error) {
		result := make(map[string]*model.StudentTeacherInfo)

		if len(ids) == 0 {
			return result, nil
		}
		resp, err := client.GetTeachersByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				Teacher, err := client.GetTeacherById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Teacher %s: %v", id, err)
					continue
				}

				if Teacher != nil && Teacher.Teacher != nil {
					result[id] = convert.PbTeacherToStudentTeacherInfo(Teacher.Teacher)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Teachers != nil {
			for _, pbTeacher := range resp.Teachers {
				if pbTeacher != nil {
					result[pbTeacher.Id] = convert.PbTeacherToStudentTeacherInfo(pbTeacher)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d Teachers successfully", len(result), len(ids))
		return result, nil
	}
}
