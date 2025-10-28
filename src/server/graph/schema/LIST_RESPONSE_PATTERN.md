# List Response Pattern - Total + Data Structure

## Nguyên tắc

**Tất cả list fields phải có structure `{ total, data }` chứ không phải flat array!**

## Lý do

1. **Pagination**: Frontend cần biết total để tính số trang
2. **Consistency**: Tất cả APIs đều có format giống nhau
3. **Performance**: Có thể count total riêng mà không cần load hết data

## Pattern

### ❌ SAI - Flat Array
```graphql
type Semester {
    students: [Student!]  # ❌ SAI
}
```

### ✅ ĐÚNG - List Response
```graphql
type Semester {
    students: StudentListResponse!  # ✅ ĐÚNG
}

type StudentListResponse {
    total: Int!
    data: [Student!]!
}
```

## List Response Types đã tạo

### Base Types
```graphql
type StudentListResponse {
    total: Int!
    data: [Student!]!
}

type TeacherListResponse {
    total: Int!
    data: [Teacher!]!
}

type EnrollmentListResponse {
    total: Int!
    data: [Enrollment!]!
}

type TopicListResponse {
    total: Int!
    data: [Topic!]!
}

type TopicCouncilListResponse {
    total: Int!
    data: [TopicCouncil!]!
}

type TopicCouncilSupervisorListResponse {
    total: Int!
    data: [TopicCouncilSupervisor!]!
}

type FileListResponse {
    total: Int!
    data: [File!]!
}

type RoleSystemListResponse {
    total: Int!
    data: [RoleSystem!]!
}

type MajorListResponse {
    total: Int!
    data: [Major!]!
}

type FacultyListResponse {
    total: Int!
    data: [Faculty!]!
}

type SemesterListResponse {
    total: Int!
    data: [Semester!]!
}

type CouncilListResponse {
    total: Int!
    data: [Council!]!
}

type DefenceListResponse {
    total: Int!
    data: [Defence!]!
}

type GradeDefenceListResponse {
    total: Int!
    data: [GradeDefence!]!
}

type GradeDefenceCriterionListResponse {
    total: Int!
    data: [GradeDefenceCriterion!]!
}

type GradeReviewListResponse {
    total: Int!
    data: [GradeReview!]!
}

type MidtermListResponse {
    total: Int!
    data: [Midterm!]!
}

type FinalListResponse {
    total: Int!
    data: [Final!]!
}
```

### Student Custom Types
```graphql
type StudentEnrollmentListResponse {
    total: Int!
    data: [StudentEnrollment!]!
}

type StudentGradeDefenceListResponse {
    total: Int!
    data: [StudentGradeDefence!]!
}

type StudentTopicSupervisorListResponse {
    total: Int!
    data: [StudentTopicSupervisor!]!
}

type StudentDefenceInfoListResponse {
    total: Int!
    data: [StudentDefenceInfo!]!
}
```

### Supervisor Custom Types
```graphql
type SupervisorTopicCouncilAssignmentListResponse {
    total: Int!
    data: [SupervisorTopicCouncilAssignment!]!
}

type SupervisorEnrollmentListResponse {
    total: Int!
    data: [SupervisorEnrollment!]!
}

type SupervisorTopicCouncilListResponse {
    total: Int!
    data: [SupervisorTopicCouncil!]!
}
```

### Council Member Custom Types
```graphql
type CouncilDefenceListResponse {
    total: Int!
    data: [CouncilDefence!]!
}

type CouncilTopicCouncilListResponse {
    total: Int!
    data: [CouncilTopicCouncil!]!
}

type CouncilEnrollmentListResponse {
    total: Int!
    data: [CouncilEnrollment!]!
}
```

### Reviewer Custom Types
```graphql
type ReviewerGradeReviewListResponse {
    total: Int!
    data: [ReviewerGradeReview!]!
}
```

## Updated Types

### academic.graphqls
```graphql
type Semester {
    students: StudentListResponse!  # ✅ ĐÚNG
    teachers: TeacherListResponse!  # ✅ ĐÚNG
    topics: TopicListResponse!      # ✅ ĐÚNG
}

type Major {
    topics: TopicListResponse!      # ✅ ĐÚNG
}

type Faculty {
    majors: MajorListResponse!      # ✅ ĐÚNG
}
```

### user.graphqls
```graphql
type Student {
    enrollments: EnrollmentListResponse!  # ✅ ĐÚNG
}

type Teacher {
    roles: RoleSystemListResponse!        # ✅ ĐÚNG
}
```

### thesis.graphqls
```graphql
type Topic {
    files: FileListResponse!                        # ✅ ĐÚNG
    topicCouncils: TopicCouncilListResponse!        # ✅ ĐÚNG
}

type TopicCouncil {
    enrollments: EnrollmentListResponse!                    # ✅ ĐÚNG
    supervisors: TopicCouncilSupervisorListResponse!        # ✅ ĐÚNG
}

type Enrollment {
    gradeDefences: GradeDefenceListResponse!        # ✅ ĐÚNG
}
```

### council.graphqls
```graphql
type Council {
    defences: DefenceListResponse!                  # ✅ ĐÚNG
    topicCouncils: TopicCouncilListResponse!        # ✅ ĐÚNG
}

type Defence {
    gradeDefences: GradeDefenceListResponse!        # ✅ ĐÚNG
}

type GradeDefence {
    criteria: GradeDefenceCriterionListResponse!    # ✅ ĐÚNG
}
```

### student.graphqls
```graphql
type StudentEnrollment {
    gradeDefences: StudentGradeDefenceListResponse!  # ✅ ĐÚNG
}

type StudentTopicCouncil {
    supervisors: StudentTopicSupervisorListResponse! # ✅ ĐÚNG
}

type StudentTopic {
    files: FileListResponse!                         # ✅ ĐÚNG
}

type StudentCouncil {
    defences: StudentDefenceInfoListResponse!        # ✅ ĐÚNG
}

type StudentGradeDefence {
    criteria: GradeDefenceCriterionListResponse!     # ✅ ĐÚNG
}
```

### teacher_general.graphqls
```graphql
type SupervisorTopic {
    files: FileListResponse!                                # ✅ ĐÚNG
    topicCouncils: SupervisorTopicCouncilListResponse!      # ✅ ĐÚNG
}

type SupervisorEnrollment {
    gradeDefences: GradeDefenceListResponse!                # ✅ ĐÚNG
}

type SupervisorTopicCouncil {
    enrollments: SupervisorEnrollmentListResponse!          # ✅ ĐÚNG
    supervisors: TopicCouncilSupervisorListResponse!        # ✅ ĐÚNG
}

type CouncilMemberCouncil {
    defences: CouncilDefenceListResponse!                   # ✅ ĐÚNG
    topicCouncils: CouncilTopicCouncilListResponse!         # ✅ ĐÚNG
}

type CouncilDefence {
    gradeDefences: GradeDefenceListResponse!                # ✅ ĐÚNG
}

type CouncilTopicCouncil {
    enrollments: CouncilEnrollmentListResponse!             # ✅ ĐÚNG
    supervisors: TopicCouncilSupervisorListResponse!        # ✅ ĐÚNG
}

type CouncilEnrollment {
    gradeDefences: GradeDefenceListResponse!                # ✅ ĐÚNG
}

type ReviewerTopicCouncil {
    supervisors: TopicCouncilSupervisorListResponse!        # ✅ ĐÚNG
}

type ReviewerTopic {
    files: FileListResponse!                                # ✅ ĐÚNG
}
```

## Queries với List Response

### Student Queries
```graphql
extend type Query {
    getMyEnrollments(search: SearchRequestInput): StudentEnrollmentListResponse!  # ✅
    getMySemesters(search: SearchRequestInput): SemesterListResponse!             # ✅
}
```

### Supervisor Queries
```graphql
extend type Query {
    getMySupervisedTopicCouncils(search: SearchRequestInput): SupervisorTopicCouncilAssignmentListResponse!  # ✅
}
```

### Council Member Queries
```graphql
extend type Query {
    getMyDefences(search: SearchRequestInput): CouncilDefenceListResponse!  # ✅
}
```

### Reviewer Queries
```graphql
extend type Query {
    getMyGradeReviews(search: SearchRequestInput): ReviewerGradeReviewListResponse!  # ✅
}
```

### Academic Affairs Queries
```graphql
extend type Query {
    getListTeachers(search: SearchRequestInput!): TeacherListResponse!         # ✅
    getListStudents(search: SearchRequestInput!): StudentListResponse!         # ✅
    getAllSemesters(search: SearchRequestInput!): SemesterListResponse!        # ✅
    getAllMajors(search: SearchRequestInput!): MajorListResponse!              # ✅
    getAllFaculties(search: SearchRequestInput!): FacultyListResponse!         # ✅
    getAllTopics(search: SearchRequestInput!): TopicListResponse!              # ✅
    getAllEnrollments(search: SearchRequestInput!): EnrollmentListResponse!    # ✅
    getAllCouncils(search: SearchRequestInput!): CouncilListResponse!          # ✅
    getDefencesByCouncil(councilId: ID!): DefenceListResponse!                 # ✅
    getAllGradeDefences(search: SearchRequestInput!): GradeDefenceListResponse!  # ✅
}
```

## Example Query & Response

### Query
```graphql
query GetAllSemesters {
    getAllSemesters {
        total  # ← Tổng số records
        data {  # ← List data
            id
            title
            students {
                total  # ← Nested list cũng có total
                data {
                    id
                    username
                }
            }
        }
    }
}
```

### Response
```json
{
  "data": {
    "getAllSemesters": {
      "total": 10,
      "data": [
        {
          "id": "sem1",
          "title": "Học kỳ 1 2024",
          "students": {
            "total": 150,
            "data": [
              {
                "id": "sv1",
                "username": "Nguyễn Văn A"
              },
              {
                "id": "sv2",
                "username": "Trần Thị B"
              }
            ]
          }
        }
      ]
    }
  }
}
```

## Generated Go Types

```go
type Semester struct {
    ID        string               `json:"id"`
    Title     string               `json:"title"`
    Students  *StudentListResponse `json:"students"`   // ✅ List response
    Teachers  *TeacherListResponse `json:"teachers"`   // ✅ List response
    Topics    *TopicListResponse   `json:"topics"`     // ✅ List response
}

type StudentListResponse struct {
    Total int32      `json:"total"`   // ✅ Total count
    Data  []*Student `json:"data"`    // ✅ Actual data
}
```

## Resolver Implementation

```go
func (r *semesterResolver) Students(ctx context.Context, obj *model.Semester) (*model.StudentListResponse, error) {
    // Query students WHERE semester_code = obj.ID
    students, total, err := r.studentService.GetBySemesterCode(ctx, obj.ID)
    if err != nil {
        return nil, err
    }

    return &model.StudentListResponse{
        Total: int32(total),  // ✅ Return total
        Data:  students,      // ✅ Return data
    }, nil
}
```

## Benefits

✅ **Pagination Support**: Frontend knows total for calculating pages
✅ **Consistency**: All APIs have same format
✅ **Performance**: Can count total separately without loading all data
✅ **Type Safety**: TypeScript auto-generated types include total + data
✅ **Flexible**: Can add more metadata fields later (hasMore, cursor, etc.)

## Summary

- ✅ Created 40+ list response types
- ✅ Updated all base types to use list responses
- ✅ Updated all custom types to use list responses
- ✅ Updated all queries to return list responses
- ✅ Regenerated GraphQL code successfully

**All list fields now have `{ total, data }` structure!**

---

**Created:** 2025-10-18
**Status:** ✅ Complete
**Pattern:** `type XyzListResponse { total: Int!, data: [Xyz!]! }`
