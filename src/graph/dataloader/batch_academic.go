package dataloader

import (
	"context"
	"log"
	"thaily/src/graph/convert"
	"thaily/src/graph/model"
	"thaily/src/server/client"
)

func createMajorInfoBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, *model.MajorInfo] {
	return func(ctx context.Context, ids []string) (map[string]*model.MajorInfo, error) {
		result := make(map[string]*model.MajorInfo)

		if len(ids) == 0 {
			return result, nil
		}
		resp, err := client.GetMajorsByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				Major, err := client.GetMajorById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Major %s: %v", id, err)
					continue
				}

				if Major != nil && Major.Major != nil {
					result[id] = convert.PbMajorToMajorInfo(Major.Major)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Majors != nil {
			for _, pbMajor := range resp.Majors {
				if pbMajor != nil {
					result[pbMajor.Id] = convert.PbMajorToMajorInfo(pbMajor)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d Majors successfully", len(result), len(ids))
		return result, nil
	}
}

func createSemesterInfoBatchFunc(client *client.GRPCAcadamicClient) BatchFunc[string, *model.SemesterInfo] {
	return func(ctx context.Context, ids []string) (map[string]*model.SemesterInfo, error) {
		result := make(map[string]*model.SemesterInfo)
		if len(ids) == 0 {
			return result, nil
		}
		resp, err := client.GetSemestersByIds(ctx, ids)
		if err != nil {
			log.Printf("[DataLoader] Batch fetch failed, falling back to individual: %v", err)
			for _, id := range ids {
				Semester, err := client.GetSemesterById(ctx, id)
				if err != nil {
					log.Printf("[DataLoader] Failed to fetch Semester %s: %v", id, err)
					continue
				}

				if Semester != nil && Semester.Semester != nil {
					result[id] = convert.PbSemesterToSemesterInfo(Semester.Semester)
				}
			}
			log.Printf("[DataLoader] Individual fetch completed: %d/%d successful", len(result), len(ids))
			return result, nil
		}

		if resp != nil && resp.Semesters != nil {
			for _, pbSemester := range resp.Semesters {
				if pbSemester != nil {
					result[pbSemester.Id] = convert.PbSemesterToSemesterInfo(pbSemester)
				}
			}
		}

		log.Printf("[DataLoader] Batch loaded %d/%d Semesters successfully", len(result), len(ids))
		return result, nil
	}
}
