# Implementation Example - Custom Type Resolvers

## Overview

Sau khi generate, gqlgen đã tạo resolver functions cho tất cả custom types. Bây giờ cần implement các functions này với type mapping và DataLoader.

## Generated Resolver Functions

### Student Resolvers

File: `student.resolvers.go`

```go
// Query resolvers
func (r *queryResolver) GetMyEnrollments(ctx context.Context, search *model.SearchRequestInput) ([]*model.StudentEnrollment, error)
func (r *queryResolver) GetMyEnrollmentDetail(ctx context.Context, id string) (*model.StudentEnrollment, error)

// StudentEnrollment field resolvers
func (r *studentEnrollmentResolver) TopicCouncil(ctx context.Context, obj *model.StudentEnrollment) (*model.StudentTopicCouncil, error)
func (r *studentEnrollmentResolver) Midterm(ctx context.Context, obj *model.StudentEnrollment) (*model.Midterm, error)
func (r *studentEnrollmentResolver) Final(ctx context.Context, obj *model.StudentEnrollment) (*model.Final, error)
func (r *studentEnrollmentResolver) GradeReview(ctx context.Context, obj *model.StudentEnrollment) (*model.GradeReview, error)
func (r *studentEnrollmentResolver) GradeDefences(ctx context.Context, obj *model.StudentEnrollment) ([]*model.StudentGradeDefence, error)

// StudentTopicCouncil field resolvers
func (r *studentTopicCouncilResolver) Topic(ctx context.Context, obj *model.StudentTopicCouncil) (*model.StudentTopic, error)
func (r *studentTopicCouncilResolver) Supervisors(ctx context.Context, obj *model.StudentTopicCouncil) ([]*model.StudentTopicSupervisor, error)

// StudentTopicSupervisor field resolvers
func (r *studentTopicSupervisorResolver) Teacher(ctx context.Context, obj *model.StudentTopicSupervisor) (*model.StudentTeacherInfo, error)

// ... và nhiều functions khác
```

### Supervisor Resolvers

File: `teacher_general.resolvers.go`

```go
// Query resolvers
func (r *queryResolver) GetMySupervisedTopics(ctx context.Context, search *model.SearchRequestInput) ([]*model.SupervisorTopic, error)
func (r *queryResolver) GetMySupervisedEnrollments(ctx context.Context, search *model.SearchRequestInput) ([]*model.SupervisorEnrollment, error)

// SupervisorEnrollment field resolvers
func (r *supervisorEnrollmentResolver) Student(ctx context.Context, obj *model.SupervisorEnrollment) (*model.Student, error)
func (r *supervisorEnrollmentResolver) TopicCouncil(ctx context.Context, obj *model.SupervisorEnrollment) (*model.SupervisorTopicCouncil, error)

// ... và nhiều functions khác
```

## Implementation Pattern

### 1. Query Resolver - Fetch và Convert

```go
// src/graph/resolver/student.resolvers.go

func (r *queryResolver) GetMyEnrollments(
    ctx context.Context,
    search *model.SearchRequestInput,
) ([]*model.StudentEnrollment, error) {
    // 1. Lấy user từ context
    user := auth.GetUserFromContext(ctx)
    if user == nil {
        return nil, fmt.Errorf("unauthorized")
    }

    // 2. Fetch enrollments từ service (filter by student_code)
    enrollments, err := r.EnrollmentService.GetByStudentCode(ctx, user.ID, search)
    if err != nil {
        return nil, err
    }

    // 3. Convert từ full Enrollment type → StudentEnrollment type
    studentEnrollments := make([]*model.StudentEnrollment, len(enrollments))
    for i, enrollment := range enrollments {
        studentEnrollments[i] = toStudentEnrollment(enrollment)
    }

    return studentEnrollments, nil
}
```

### 2. Field Resolver - Load via DataLoader + Convert

```go
// src/graph/resolver/student.resolvers.go

func (r *studentEnrollmentResolver) TopicCouncil(
    ctx context.Context,
    obj *model.StudentEnrollment,
) (*model.StudentTopicCouncil, error) {
    // 1. Load full TopicCouncil via DataLoader (batch loading)
    loaders := dataloader.GetLoaders(ctx)
    topicCouncil, err := loaders.TopicCouncilByID.Load(obj.TopicCouncilCode)
    if err != nil {
        return nil, err
    }

    // 2. Convert full TopicCouncil → StudentTopicCouncil (chỉ copy allowed fields)
    return toStudentTopicCouncil(topicCouncil), nil
}

func (r *studentEnrollmentResolver) Midterm(
    ctx context.Context,
    obj *model.StudentEnrollment,
) (*model.Midterm, error) {
    if obj.MidtermCode == nil {
        return nil, nil
    }

    // Load via DataLoader (batch loading)
    loaders := dataloader.GetLoaders(ctx)
    return loaders.MidtermByID.Load(*obj.MidtermCode)
}
```

### 3. Nested Type Resolver

```go
// src/graph/resolver/student.resolvers.go

func (r *studentTopicCouncilResolver) Topic(
    ctx context.Context,
    obj *model.StudentTopicCouncil,
) (*model.StudentTopic, error) {
    // 1. Load full Topic via DataLoader
    loaders := dataloader.GetLoaders(ctx)
    topic, err := loaders.TopicByID.Load(obj.TopicCode)
    if err != nil {
        return nil, err
    }

    // 2. Convert full Topic → StudentTopic (restricted)
    return toStudentTopic(topic), nil
}

func (r *studentTopicCouncilResolver) Supervisors(
    ctx context.Context,
    obj *model.StudentTopicCouncil,
) ([]*model.StudentTopicSupervisor, error) {
    // 1. Load full TopicCouncilSupervisors
    supervisors, err := r.TopicCouncilSupervisorService.GetByTopicCouncilCode(ctx, obj.ID)
    if err != nil {
        return nil, err
    }

    // 2. Convert to StudentTopicSupervisor
    studentSupervisors := make([]*model.StudentTopicSupervisor, len(supervisors))
    for i, sup := range supervisors {
        studentSupervisors[i] = toStudentTopicSupervisor(sup)
    }

    return studentSupervisors, nil
}
```

## Type Mapping Functions

Tạo file `src/graph/resolver/type_mapper.go`:

```go
package resolver

import "thaily/src/server/graph/model"

// ============================================
// ENROLLMENT MAPPERS
// ============================================

// toStudentEnrollment converts full Enrollment → StudentEnrollment
func toStudentEnrollment(e *model.Enrollment) *model.StudentEnrollment {
	if e == nil {
		return nil
	}

	return &model.StudentEnrollment{
		ID:               e.ID,
		Title:            e.Title,
		StudentCode:      e.StudentCode,
		TopicCouncilCode: e.TopicCouncilCode,
		FinalCode:        e.FinalCode,
		GradeReviewCode:  e.GradeReviewCode,
		MidtermCode:      e.MidtermCode,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
		CreatedBy:        e.CreatedBy,
		UpdatedBy:        e.UpdatedBy,
		// Relationships will be resolved by field resolvers
	}
}

// toSupervisorEnrollment converts full Enrollment → SupervisorEnrollment
func toSupervisorEnrollment(e *model.Enrollment) *model.SupervisorEnrollment {
	if e == nil {
		return nil
	}

	return &model.SupervisorEnrollment{
		ID:               e.ID,
		Title:            e.Title,
		StudentCode:      e.StudentCode,
		TopicCouncilCode: e.TopicCouncilCode,
		FinalCode:        e.FinalCode,
		GradeReviewCode:  e.GradeReviewCode,
		MidtermCode:      e.MidtermCode,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
		CreatedBy:        e.CreatedBy,
		UpdatedBy:        e.UpdatedBy,
		// Relationships will be resolved by field resolvers
	}
}

// ============================================
// TOPIC COUNCIL MAPPERS
// ============================================

// toStudentTopicCouncil converts full TopicCouncil → StudentTopicCouncil
func toStudentTopicCouncil(tc *model.TopicCouncil) *model.StudentTopicCouncil {
	if tc == nil {
		return nil
	}

	return &model.StudentTopicCouncil{
		ID:        tc.ID,
		Title:     tc.Title,
		Stage:     tc.Stage,
		TopicCode: tc.TopicCode,
		// KHÔNG copy CouncilCode - student không được xem
		TimeStart: tc.TimeStart,
		TimeEnd:   tc.TimeEnd,
		CreatedAt: tc.CreatedAt,
		UpdatedAt: tc.UpdatedAt,
	}
}

// toSupervisorTopicCouncil converts full TopicCouncil → SupervisorTopicCouncil
func toSupervisorTopicCouncil(tc *model.TopicCouncil) *model.SupervisorTopicCouncil {
	if tc == nil {
		return nil
	}

	return &model.SupervisorTopicCouncil{
		ID:          tc.ID,
		Title:       tc.Title,
		Stage:       tc.Stage,
		TopicCode:   tc.TopicCode,
		CouncilCode: tc.CouncilCode, // ✅ Supervisor được xem councilCode
		TimeStart:   tc.TimeStart,
		TimeEnd:     tc.TimeEnd,
		CreatedAt:   tc.CreatedAt,
		UpdatedAt:   tc.UpdatedAt,
	}
}

// ============================================
// TOPIC MAPPERS
// ============================================

// toStudentTopic converts full Topic → StudentTopic
func toStudentTopic(t *model.Topic) *model.StudentTopic {
	if t == nil {
		return nil
	}

	return &model.StudentTopic{
		ID:            t.ID,
		Title:         t.Title,
		MajorCode:     t.MajorCode,
		SemesterCode:  t.SemesterCode,
		Status:        t.Status,
		PercentStage1: t.PercentStage1,
		PercentStage2: t.PercentStage2,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		// KHÔNG copy sensitive fields
	}
}

// toSupervisorTopic converts full Topic → SupervisorTopic
func toSupervisorTopic(t *model.Topic) *model.SupervisorTopic {
	if t == nil {
		return nil
	}

	return &model.SupervisorTopic{
		ID:            t.ID,
		Title:         t.Title,
		MajorCode:     t.MajorCode,
		SemesterCode:  t.SemesterCode,
		Status:        t.Status,
		PercentStage1: t.PercentStage1,
		PercentStage2: t.PercentStage2,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		CreatedBy:     t.CreatedBy,
		UpdatedBy:     t.UpdatedBy,
		// Supervisor gets full fields
	}
}

// ============================================
// TEACHER MAPPERS
// ============================================

// toStudentTeacherInfo converts full Teacher → StudentTeacherInfo
func toStudentTeacherInfo(t *model.Teacher) *model.StudentTeacherInfo {
	if t == nil {
		return nil
	}

	return &model.StudentTeacherInfo{
		ID:        t.ID,
		Email:     t.Email,
		Username:  t.Username,
		Gender:    t.Gender,
		MajorCode: t.MajorCode,
		// KHÔNG copy roles, semester, etc - student chỉ xem thông tin cơ bản
	}
}

// toStudentTopicSupervisor converts full TopicCouncilSupervisor → StudentTopicSupervisor
func toStudentTopicSupervisor(tcs *model.TopicCouncilSupervisor) *model.StudentTopicSupervisor {
	if tcs == nil {
		return nil
	}

	return &model.StudentTopicSupervisor{
		ID:                    tcs.ID,
		TeacherSupervisorCode: tcs.TeacherSupervisorCode,
		TopicCouncilCode:      tcs.TopicCouncilCode,
		// Relationships resolved by field resolvers
	}
}

// ============================================
// GRADE DEFENCE MAPPERS
// ============================================

// toStudentGradeDefence converts full GradeDefence → StudentGradeDefence
func toStudentGradeDefence(gd *model.GradeDefence) *model.StudentGradeDefence {
	if gd == nil {
		return nil
	}

	return &model.StudentGradeDefence{
		ID:             gd.ID,
		DefenceCode:    gd.DefenceCode,
		EnrollmentCode: gd.EnrollmentCode,
		Note:           gd.Note,
		TotalScore:     gd.TotalScore,
		CreatedAt:      gd.CreatedAt,
		UpdatedAt:      gd.UpdatedAt,
	}
}

// toStudentDefenceInfo converts full Defence → StudentDefenceInfo
func toStudentDefenceInfo(d *model.Defence) *model.StudentDefenceInfo {
	if d == nil {
		return nil
	}

	return &model.StudentDefenceInfo{
		ID:       d.ID,
		Title:    d.Title,
		Position: d.Position,
		// Teacher resolved by field resolver
	}
}
```

## Complete Implementation Example

### Student Enrollment Flow

```go
// 1. User queries
query GetMyEnrollments {
    getMyEnrollments {
        id
        title
        topicCouncil {
            topic {
                title
            }
            supervisors {
                teacher {
                    username
                }
            }
        }
    }
}

// 2. Execution flow
GetMyEnrollments (query resolver)
  ↓ fetch enrollments from DB (filter by student_code)
  ↓ convert to StudentEnrollment
  ↓
StudentEnrollment.topicCouncil (field resolver)
  ↓ load TopicCouncil via DataLoader (batch)
  ↓ convert to StudentTopicCouncil
  ↓
StudentTopicCouncil.topic (field resolver)
  ↓ load Topic via DataLoader (batch)
  ↓ convert to StudentTopic
  ↓
StudentTopicCouncil.supervisors (field resolver)
  ↓ load TopicCouncilSupervisors
  ↓ convert to StudentTopicSupervisor
  ↓
StudentTopicSupervisor.teacher (field resolver)
  ↓ load Teacher via DataLoader (batch)
  ↓ convert to StudentTeacherInfo
```

### DataLoader Benefits

```go
// BAD: N+1 queries without DataLoader
for enrollment in enrollments {
    topicCouncil = db.GetTopicCouncil(enrollment.topicCouncilCode)  // N queries!
}

// GOOD: Batch loading with DataLoader
for enrollment in enrollments {
    topicCouncil = dataloader.Load(enrollment.topicCouncilCode)  // 1 batch query!
}

// DataLoader automatically batches:
// SELECT * FROM topic_council WHERE id IN ('id1', 'id2', 'id3', ...)
```

## Testing

```go
// src/graph/resolver/student_test.go

func TestStudentCannotAccessSensitiveFields(t *testing.T) {
    // Test 1: StudentTopicCouncil không có field councilCode
    tc := &model.TopicCouncil{
        ID:          "tc1",
        CouncilCode: strPtr("council123"),
    }

    studentTC := toStudentTopicCouncil(tc)

    // StudentTopicCouncil không có field CouncilCode
    // Compile error nếu try access: studentTC.CouncilCode
    assert.Equal(t, "tc1", studentTC.ID)
}

func TestSupervisorCanAccessAllFields(t *testing.T) {
    // Test 2: SupervisorTopicCouncil có field councilCode
    tc := &model.TopicCouncil{
        ID:          "tc1",
        CouncilCode: strPtr("council123"),
    }

    supervisorTC := toSupervisorTopicCouncil(tc)

    assert.Equal(t, "tc1", supervisorTC.ID)
    assert.Equal(t, "council123", *supervisorTC.CouncilCode) // ✅ OK
}
```

## Next Steps

1. ✅ gqlgen.yml updated - DONE
2. ✅ GraphQL generated - DONE
3. ⏳ Implement type mapper functions
4. ⏳ Implement query resolvers
5. ⏳ Implement field resolvers with DataLoader
6. ⏳ Add missing DataLoaders
7. ⏳ Write tests
8. ⏳ Test với GraphQL playground

## File Structure

```
src/graph/resolver/
├── type_mapper.go                      # ⏳ CREATE - Type conversion functions
├── student.resolvers.go                # ⏳ IMPLEMENT - Student resolvers
├── teacher_general.resolvers.go        # ⏳ IMPLEMENT - Teacher resolvers
├── academic_affairs.resolvers.go       # ⏳ IMPLEMENT - Admin resolvers
├── department_lecturer.resolvers.go    # ⏳ IMPLEMENT - Dept resolvers
└── *_test.go                          # ⏳ CREATE - Tests
```

Resolver files đã được generate với các function signatures đúng. Chỉ cần implement logic bên trong! 🚀
