import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 10,
            maxVUs: 200
        }
    }
};

const JWT_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Imx5dmluaHRoYWkzMjFAZ21haWwuY29tIiwiZXhwIjoxNzYxMDQwMzc3LCJnb29nbGVfaWQiOiIxMDQ4NjUyNzA0NDQwODgzNzQ2MzEiLCJpYXQiOjE3NjEwMzk0NzcsImlkcyI6IlNFTV8yMDI1XzEtU1RVXzAwMDAwMixTRU1fMjAyNF8yLVNUVV8wMDAwMDMsU0VNXzIwMjNfMi1TVFVfMDAwMDA1LFNFTV8yMDIzXzEtU1RVXzAwMDAwMSwiLCJuYW1lIjoiIiwicm9sZSI6InN0dWRlbnQifQ.e8Vq44yoVslzNV23nIRSGIC5ZEgCxEXp4r5KMJ61FF0';

const GRAPHQL_ENDPOINT = 'http://localhost:8080/query';

const query = `
query GetMyEnrollments {
  getMyEnrollments {
    total
    data {
      id
      title
      studentCode
      topicCouncilCode
      finalCode
      gradeReviewCode
      midtermCode
      createdAt
      updatedAt
      createdBy
      updatedBy
      topicCouncil {
        id
        title
        stage
        topicCode
        councilCode
        timeStart
        timeEnd
        createdAt
        updatedAt
        council {
          id
          title
          majorCode
          semesterCode
          timeStart
          createdAt
          updatedAt
          major {
            id
            title
            facultyCode
          }
          semester {
            id
            title
          }
          defences {
            id
            title
            teacher_code
            position
            createdAt
            updatedAt
            teacher {
              id
              email
              username
              gender
              majorCode
            }
          }
        }
        topic {
          id
          title
          majorCode
          semesterCode
          status
          percentStage1
          percentStage2
          createdAt
          updatedAt
          major {
            id
            title
            facultyCode
          }
          semester {
            id
            title
          }
        }
        supervisors {
          id
          teacherSupervisorCode
          topicCouncilCode
          teacher {
            id
            email
            username
            gender
            majorCode
          }
        }
      }
      midterm {
        id
        title
        grade
        status
        feedback
        createdAt
        updatedAt
        createdBy
        updatedBy
      }
      final {
        id
        title
        supervisorGrade
        departmentGrade
        finalGrade
        status
        notes
        completionDate
        createdAt
        updatedAt
        createdBy
        updatedBy
      }
      gradeDefences {
        id
        defenceCode
        enrollmentCode
        note
        totalScore
        createdAt
        updatedAt
        criteria {
          id
          gradeDefenceCode
          name
          score
          maxScore
          createdAt
          updatedAt
          createdBy
          updatedBy
        }
        defence {
          id
          title
          teacher_code
          position
          createdAt
          updatedAt
          teacher {
            id
            email
            username
            gender
            majorCode
          }
        }
      }
      gradeReview {
        id
        title
        reviewGrade
        teacherCode
        status
        notes
        completionDate
        createdAt
        updatedAt
        createdBy
        updatedBy
      }
    }
  }
}
`;

export default function () {
    const payload = JSON.stringify({
        query: query,
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${JWT_TOKEN}`,
        },
    };

    const response = http.post(GRAPHQL_ENDPOINT, payload, params);

    check(response, {
        'status is 200': (r) => r.status === 200,
        'response has data': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.data && body.data.getMyEnrollments !== null;
            } catch (e) {
                return false;
            }
        },
        'no errors': (r) => {
            try {

            } catch (e) {
                return false;
            }
        },
    });

    sleep(0.1);
}