#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script tạo dữ liệu lớn cho hệ thống quản lý luận văn
Sử dụng: python generate_data.py [số_lượng_records]
Ví dụ: python generate_data.py 1000

Tác giả: Lý Vĩnh Thái
"""

import subprocess
import sys
import random

def generate_data(num_records):
    """Tạo dữ liệu với số lượng records chỉ định"""
    
    print(f"🚀 Đang tạo {num_records} records cho hệ thống quản lý luận văn...")
    
    # Tính toán số lượng cho từng loại
    num_teachers = max(num_records // 20, 10)  # Ít nhất 10 teachers, tỉ lệ 1:20
    num_students = num_records
    num_topics = num_records
    num_submissions = int(num_records * 0.6)  # 60% có submissions
    num_evaluations = int(num_records * 0.4)  # 40% có evaluations
    
    print(f"📊 Sẽ tạo:")
    print(f"   👥 Teachers: {num_teachers}")
    print(f"   🎓 Students: {num_students}")
    print(f"   📚 Topics: {num_topics}")
    print(f"   📝 Submissions: ~{num_submissions}")
    print(f"   📋 Evaluations: ~{num_evaluations}")
    print(f"   📈 Enrollments: {num_records}")
    
    confirm = input(f"\n⚠️  Cảnh báo: Script sẽ XÓA toàn bộ dữ liệu cũ (trừ admin teacher001)!\nTiếp tục tạo {num_records} records? (y/N): ")
    if confirm.lower() != 'y':
        print("❌ Đã hủy")
        return
    
    # 1. Xóa dữ liệu cũ
    print("\n🗑️  Đang xóa dữ liệu cũ...")
    clear_sql = """
SET FOREIGN_KEY_CHECKS = 0;
DELETE FROM defense_results;
DELETE FROM defense_schedules; 
DELETE FROM council_members;
DELETE FROM defense_councils;
DELETE FROM final_grades;
DELETE FROM final_evaluations;
DELETE FROM midterm_evaluations;
DELETE FROM submissions;
DELETE FROM topic_assignments;
DELETE FROM student_enrollments;
DELETE FROM teacher_permissions WHERE teacher_id != 'teacher001';
DELETE FROM topics;
DELETE FROM students;
DELETE FROM teachers WHERE id != 'teacher001';
SET FOREIGN_KEY_CHECKS = 1;
"""
    run_sql(clear_sql)
    
    # 2. Tạo semesters
    print("📅 Tạo semesters...")
    semesters_sql = """
INSERT INTO semesters (id, name, start_date, end_date, is_active, registration_deadline, submission_deadline, defense_start_date, defense_end_date) VALUES
('2023-1', 'Học kỳ I năm học 2023-2024', '2023-09-01', '2023-12-31', 0, '2023-10-15', '2023-11-30', '2023-12-01', '2023-12-20'),
('2023-2', 'Học kỳ II năm học 2023-2024', '2024-01-15', '2024-05-31', 0, '2024-02-28', '2024-04-30', '2024-05-01', '2024-05-25'),
('2024-1', 'Học kỳ I năm học 2024-2025', '2024-09-01', '2024-12-31', 1, '2024-10-15', '2024-11-30', '2024-12-01', '2024-12-20'),
('2024-2', 'Học kỳ II năm học 2024-2025', '2025-01-15', '2025-05-31', 0, '2025-02-28', '2025-04-30', '2025-05-01', '2025-05-25');
"""
    run_sql(semesters_sql)
    
    # 3. Tạo teachers
    print(f"👥 Tạo {num_teachers} teachers...")
    departments = ['Khoa học máy tính', 'Công nghệ thông tin', 'Kỹ thuật phần mềm', 'An toàn thông tin']
    positions = ['Tiến sĩ', 'Thạc sĩ', 'Phó giáo sư', 'Giáo sư']
    specializations = ['Machine Learning', 'Deep Learning', 'Blockchain', 'IoT', 'Cybersecurity', 
                      'Software Engineering', 'Database Systems', 'Computer Networks', 'Mobile Development', 'Web Development']
    
    teacher_values = []
    for i in range(2, num_teachers + 1):
        dept = random.choice(departments)
        pos = random.choice(positions)
        spec = random.choice(specializations)
        teacher_values.append(f"('teacher{i:03d}', 'GV{i:03d}', 'Giáo viên {i}', 'teacher{i}@hcmut.edu.vn', '090{random.randint(1000000, 9999999)}', '{dept}', '{pos}', '{spec}', 1)")
    
    insert_batch("teachers", 
                "(id, teacher_code, name, email, phone, department, position, specialization, is_active)", 
                teacher_values)
    
    # 4. Tạo students
    print(f"🎓 Tạo {num_students} students...")
    majors = ['Khoa học máy tính', 'Công nghệ thông tin', 'Kỹ thuật phần mềm']
    classes = ['CS2020', 'CS2021', 'CS2022', 'IT2020', 'IT2021', 'IT2022', 'SE2020', 'SE2021', 'SE2022']
    academic_years = ['2020-2024', '2021-2025', '2022-2026']
    genders = ['male', 'female']
    
    student_values = []
    for i in range(1, num_students + 1):
        major = random.choice(majors)
        class_code = random.choice(classes)
        academic_year = random.choice(academic_years)
        gender = random.choice(genders)
        gpa = round(random.uniform(2.0, 4.0), 2)
        birth_year = random.randint(2000, 2003)
        birth_month = random.randint(1, 12)
        birth_day = random.randint(1, 28)
        
        student_values.append(f"('student{i:03d}', 'SV{i:03d}', 'Sinh viên {i}', 'student{i}@hcmut.edu.vn', '098{random.randint(1000000, 9999999)}', '{birth_year}-{birth_month:02d}-{birth_day:02d}', '{gender}', 'Địa chỉ {i}, TP.HCM', '{major}', '{class_code}', '{academic_year}', {gpa})")
    
    insert_batch("students",
                "(id, student_code, name, email, phone, date_of_birth, gender, address, major, class_code, academic_year, gpa)",
                student_values)
    
    # 5. Tạo topics
    print(f"📚 Tạo {num_topics} topics...")
    categories = ['Machine Learning', 'Deep Learning', 'Blockchain', 'IoT', 'Cybersecurity', 
                 'Software Engineering', 'Database Systems', 'Computer Networks', 'Mobile Development', 'Web Development']
    statuses = ['approved', 'in_progress', 'end_progress', 'completed']
    semesters = ['2023-1', '2023-2', '2024-1', '2024-2']
    
    topic_values = []
    for i in range(1, num_topics + 1):
        category = random.choice(categories)
        status = random.choice(statuses)
        semester = random.choice(semesters)
        supervisor = f"teacher{random.randint(2, num_teachers):03d}"
        reviewer = f"teacher{random.randint(2, num_teachers):03d}"
        while reviewer == supervisor:
            reviewer = f"teacher{random.randint(2, num_teachers):03d}"
            
        topic_values.append(f"('topic{i:03d}', 'Nghiên cứu {category} ứng dụng {i}', 'Research {category} Application {i}', 'Mô tả chi tiết cho đề tài nghiên cứu {category} số {i}', '{category}', 1, '{status}', '2024-09-01', '2024-09-15', 'teacher001', '{semester}', '{supervisor}', '{reviewer}')")
    
    insert_batch("topics",
                "(id, title_vi, title_en, description, category, max_students, status, submitted_date, approved_date, approved_by, semester_id, supervisor_id, reviewer_id)",
                topic_values)
    
    # 6. Tạo enrollments
    print(f"📈 Tạo {num_records} enrollments...")
    enroll_values = []
    for i in range(1, min(num_students, num_topics) + 1):
        status = 'active' if random.random() < 0.8 else 'graduated'
        major = random.choice(majors)
        semester = random.choice(semesters)
        enroll_values.append(f"('enroll{i:03d}', 'student{i:03d}', '{semester}', '2024-09-01', '{status}', 'LVTN', '{major}', 'topic{i:03d}', '2024-10-01')")
    
    insert_batch("student_enrollments",
                "(id, student_id, semester_id, enrollment_date, status, thesis_type, major, topic_id, assigned_date)",
                enroll_values)
    
    # 7. Tạo submissions (60% sinh viên)
    print(f"📝 Tạo ~{num_submissions} submissions...")
    submission_types = ['draft', 'midterm', 'final', 'revision']
    submission_statuses = ['pending', 'approved', 'under_review', 'rejected']
    
    submission_values = []
    for i in range(1, min(num_submissions, num_students) + 1):
        sub_type = random.choice(submission_types)
        sub_status = random.choice(submission_statuses)
        file_size = random.randint(500000, 5000000)
        version = random.randint(1, 3)
        is_final = 1 if sub_type == 'final' else 0
        
        submission_values.append(f"('sub{i:03d}', 'topic{i:03d}', 'student{i:03d}', '{sub_type}', 'submission_{i}.pdf', '/uploads/sub_{i}.pdf', {file_size}, 'pdf', NOW(), '{sub_status}', {version}, {is_final})")
    
    insert_batch("submissions",
                "(id, topic_id, student_id, submission_type, file_name, file_path, file_size, file_type, submission_date, status, version, is_final)",
                submission_values)
    
    # 8. Tạo midterm evaluations (40% sinh viên)
    print(f"📋 Tạo ~{num_evaluations} evaluations...")
    eval_values = []
    for i in range(1, min(num_evaluations, num_students) + 1):
        grade = round(random.uniform(5.0, 10.0), 1)
        eval_status = 'pass' if grade >= 6.0 else 'fail'
        evaluator = f"teacher{random.randint(2, num_teachers):03d}"
        
        eval_values.append(f"('mid{i:03d}', 'topic{i:03d}', 'student{i:03d}', '{evaluator}', {grade}, '{eval_status}', 'Đánh giá giữa kỳ', '2024-11-15')")
    
    insert_batch("midterm_evaluations",
                "(id, topic_id, student_id, evaluator_id, grade, status, feedback, evaluation_date)",
                eval_values)
    
    # 9. Tạo teacher permissions
    print("🔐 Tạo teacher permissions...")
    permissions = ['submit_topic', 'approve_topic', 'midterm_evaluation', 'review', 'grading', 'manage_students']
    
    perm_values = []
    perm_id = 1
    for teacher_num in range(2, num_teachers + 1):
        for semester in semesters:
            for perm in random.sample(permissions, random.randint(2, 4)):  # Mỗi teacher có 2-4 permissions ngẫu nhiên
                perm_values.append(f"('perm{perm_id:04d}', 'teacher{teacher_num:03d}', '{semester}', '{perm}', 'teacher001', 1)")
                perm_id += 1
    
    insert_batch("teacher_permissions",
                "(id, teacher_id, semester_id, permission, granted_by, is_active)",
                perm_values)
    
    # 10. Thống kê cuối
    print("\n📊 THỐNG KÊ CUỐI:")
    stats_sql = """
SELECT 'Teachers:' as type, COUNT(*) as count FROM teachers
UNION ALL
SELECT 'Students:', COUNT(*) FROM students  
UNION ALL
SELECT 'Topics:', COUNT(*) FROM topics
UNION ALL
SELECT 'Enrollments:', COUNT(*) FROM student_enrollments
UNION ALL
SELECT 'Submissions:', COUNT(*) FROM submissions
UNION ALL
SELECT 'Evaluations:', COUNT(*) FROM midterm_evaluations
UNION ALL
SELECT 'Permissions:', COUNT(*) FROM teacher_permissions;
"""
    
    result = run_sql(stats_sql, capture_output=True)
    print(result.stdout)
    
    print("✅ Hoàn thành tạo dữ liệu!")
    print(f"🎯 Đã tạo thành công {num_records} records với đầy đủ relationships")

def run_sql(sql, capture_output=False):
    """Thực thi SQL command"""
    cmd = ['docker', 'exec', 'mysql_container', 'mysql', '-u', 'root', '-pTh@i2004', '-D', 'thesis_management_system', '-e', sql]
    if capture_output:
        return subprocess.run(cmd, capture_output=True, text=True, stderr=subprocess.DEVNULL)
    else:
        subprocess.run(cmd, stderr=subprocess.DEVNULL)

def insert_batch(table, columns, values, batch_size=100):
    """Insert data theo batch để tránh lỗi quá lớn"""
    base_sql = f"INSERT INTO {table} {columns} VALUES "
    
    for i in range(0, len(values), batch_size):
        batch = values[i:i+batch_size]
        batch_sql = base_sql + ",".join(batch) + ";"
        run_sql(batch_sql)
        print(f"   ✓ Đã tạo {min(i+batch_size, len(values))}/{len(values)} records")

def main():
    if len(sys.argv) != 2:
        print("❌ Cách sử dụng:")
        print("   python generate_data.py [số_lượng_records]")
        print("\n📋 Ví dụ:")
        print("   python generate_data.py 500   # Tạo 500 records")
        print("   python generate_data.py 1000  # Tạo 1000 records")
        print("   python generate_data.py 5000  # Tạo 5000 records")
        print("\n⚠️  Lưu ý: Số lượng records càng lớn thì thời gian tạo càng lâu")
        sys.exit(1)
    
    try:
        num_records = int(sys.argv[1])
        if num_records <= 0:
            raise ValueError("Số lượng records phải > 0")
        if num_records > 10000:
            print("⚠️  Cảnh báo: Số lượng lớn hơn 10,000 có thể mất nhiều thời gian!")
            
        generate_data(num_records)
        
    except ValueError as e:
        print(f"❌ Lỗi: {e}")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n❌ Đã hủy bởi người dùng")
        sys.exit(1)
    except Exception as e:
        print(f"❌ Lỗi không mong muốn: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()