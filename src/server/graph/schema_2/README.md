# Schema 2 - Nested Query + Unified Types + Field-level RBAC

## Tong quan

Schema moi duoc to chuc voi cau truc **Nested Query** va **Unified Types**:
- Moi role co namespace rieng (student, teacher, affair, department)
- Chi dung 1 type cho moi entity (khong con StudentEnrollment, SupervisorEnrollment, etc.)
- RBAC duoc check o field-level resolver, tra ve `null` neu khong co quyen

## Cau truc thu muc

```
schema_2/
├── base.graphqls          # Scalars, Enums, Directives, Common Inputs
├── types.graphqls         # Unified Entity Types + Info Types + List Responses
├── query.graphqls         # Root Query + Role Namespaces
├── mutation.graphqls      # Root Mutation + Role Namespaces
├── inputs.graphqls        # Tat ca Input Types
├── subscription.graphqls  # Subscription
└── README.md              # File nay
```

## Nguyen tac thiet ke

### 1. Unified Types
- Tat ca deu dung chung `Enrollment`, `TopicCouncil`, `Topic`, `Defence`, `Council`, etc.
- **KHONG** con custom types nhu `StudentEnrollment`, `SupervisorEnrollment`, `CouncilEnrollment`
- Giam so luong types tu ~30 xuong ~15

### 2. Field-level RBAC
- Security duoc xu ly o **resolver level**, khong phai schema level
- Neu khong co quyen truy cap 1 field relation -> tra ve `null`

## Cau truc Nested Query

### Root Query

```graphql
type Query {
    student: StudentQuery!      # Set role = 'student'
    teacher: TeacherQuery!      # Set role = 'teacher'
    affair: AffairQuery!        # Set role = 'affair'
    department: DepartmentQuery! # Set role = 'department'
}
```

### Teacher co sub-namespaces

```graphql
type TeacherQuery {
    me: Teacher!
    supervisor: SupervisorQuery!    # Set role = 'supervisor'
    council: CouncilMemberQuery!    # Set role = 'council'
    reviewer: ReviewerQuery!        # Set role = 'reviewer'
}
```

## Vi du su dung

### Student query enrollments

```graphql
query {
    student {
        me { id username }
        enrollments(search: { ... }) {
            total
            data {
                id
                title
                # student = null (RBAC: student khong xem duoc student khac)
                topicCouncil {
                    id
                    title
                    # enrollments = null (RBAC: student khong xem duoc enrollment nguoi khac)
                }
                midterm { grade status }
            }
        }
    }
}
```

### Teacher (Supervisor) query topic councils

```graphql
query {
    teacher {
        me { id username }
        supervisor {
            topicCouncils(search: { ... }) {
                total
                data {
                    id
                    title
                    # enrollments = co data (supervisor duoc xem)
                    enrollments {
                        id
                        student { id username }  # supervisor duoc xem student
                    }
                }
            }
        }
    }
}
```

### Affair query full access

```graphql
query {
    affair {
        teachers(search: { ... }) { total data { id username roles { role } } }
        students(search: { ... }) { total data { id username mssv } }
        topics(search: { ... }) {
            total
            data {
                id
                title
                topicCouncils {  # affair duoc xem
                    enrollments {  # affair duoc xem
                        student { id }
                    }
                }
            }
        }
    }
}
```

## RBAC Rules trong Resolver

### Pattern co ban

```go
// Entry point resolver - Set role vao context
func (r *queryResolver) Student(ctx context.Context) (*model.StudentQuery, error) {
    directive.SetRole(ctx, "student")
    return &model.StudentQuery{}, nil
}

func (r *teacherQueryResolver) Supervisor(ctx context.Context, obj *model.TeacherQuery) (*model.SupervisorQuery, error) {
    directive.SetRole(ctx, "supervisor")
    return &model.SupervisorQuery{}, nil
}

// Field resolver - Check role va tra ve null neu khong co quyen
func (r *enrollmentResolver) Student(ctx context.Context, obj *model.Enrollment) (*model.Student, error) {
    // Student role khong duoc xem thong tin student khac
    access, _ := r.RbacAccess(ctx, []string{"supervisor", "council", "department", "affair"})
    if !access {
        return nil, nil  // tra ve null
    }
    return r.Ctrl.GetStudentById(ctx, obj.StudentCode)
}

func (r *topicCouncilResolver) Enrollments(ctx context.Context, obj *model.TopicCouncil) ([]*model.Enrollment, error) {
    // Student role khong duoc xem enrollments cua nguoi khac
    access, _ := r.RbacAccess(ctx, []string{"supervisor", "council", "department", "affair"})
    if !access {
        return nil, nil  // tra ve null
    }
    return r.Ctrl.GetEnrollmentsByTopicCouncilId(ctx, obj.ID)
}
```

### RBAC Rules Table

| Field | Allowed Roles | student | supervisor | council | reviewer | department | affair |
|-------|---------------|---------|------------|---------|----------|------------|--------|
| Enrollment.student | !student | null | ✓ | ✓ | ✓ | ✓ | ✓ |
| Enrollment.gradeDefences | student,supervisor,dept,affair | ✓ | ✓ | null | null | ✓ | ✓ |
| TopicCouncil.enrollments | !student | null | ✓ | ✓ | ✓ | ✓ | ✓ |
| Topic.topicCouncils | !student | null | ✓ | ✓ | ✓ | ✓ | ✓ |
| Council.defences | council,dept,affair | null | null | ✓ | null | ✓ | ✓ |
| Council.topicCouncils | !student | null | ✓ | ✓ | ✓ | ✓ | ✓ |
| Defence.council | council,dept,affair | null | null | ✓ | null | ✓ | ✓ |
| Defence.gradeDefences | council,dept,affair | null | null | ✓ | null | ✓ | ✓ |
| Teacher.roles | !student | null | ✓ | ✓ | ✓ | ✓ | ✓ |
| GradeReview.enrollment | reviewer | null | null | null | ✓ | null | null |

## So sanh voi Schema cu

| Aspect | Schema cu | Schema 2 |
|--------|-----------|----------|
| Query structure | Flat (extend type Query) | Nested namespaces |
| Types | ~30 custom types | ~15 unified types |
| RBAC | Tao types rieng cho role | Field-level check |
| Maintainability | Kho (duplicate code) | De (1 source of truth) |
| Query naming | `getMyEnrollments`, `getDepartmentEnrollments` | `student { enrollments }`, `department { enrollments }` |

## Migration tu schema cu

1. **Resolvers**: Update entry points va field-level RBAC checks
2. **Client queries**: Wrap trong namespace tuong ung
3. **Models**: Xoa cac custom types duplicate
4. **Dataloader**: Update de dung unified types

## Loi ich

1. **Don gian hoa**: 1 type = 1 entity, khong duplicate
2. **Ro rang**: Moi role co namespace rieng
3. **De bao tri**: Thay doi 1 field chi sua 1 noi
4. **Flexible**: RBAC co the thay doi ma khong can sua schema
5. **Type safety**: GraphQL van dam bao types
