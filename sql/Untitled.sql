CREATE TABLE `Semester` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Student` (
  `id` varchar(255) PRIMARY KEY,
  `email` varchar(255) NOT NULL,
  `phone` varchar(255) NOT NULL,
  `username` varchar(255) NOT NULL,
  `gender` ENUM ('male', 'female', 'other'),
  `major_code` varchar(255) NOT NULL,
  `class_code` varchar(255),
  `semester_code` varchar(255) NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Teacher` (
  `id` varchar(255) PRIMARY KEY,
  `email` varchar(255) NOT NULL,
  `username` varchar(255) NOT NULL,
  `gender` ENUM ('male', 'female', 'other'),
  `major_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Faculty` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Major` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `faculty_code` varchar(255) NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `RoleSystem` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `teacher_code` varchar(255),
  `role` ENUM ('Academic_affairs_staff', 'Supervisor_lecturer', 'Department_Lecturer', 'Reviewer_Lecturer') NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `activate` boolean NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `File` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `file` varchar(255),
  `status` ENUM ('pending', 'approved', 'rejected') NOT NULL,
  `table` ENUM ('topic', 'midterm', 'final', 'order') NOT NULL,
  `option` varchar(255),
  `table_id` varchar(255) NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Midterm` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `grade` int,
  `status` ENUM ('not_submitted', 'submitted', 'graded') NOT NULL,
  `feedback` text,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Enrollment` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `student_code` varchar(255) NOT NULL,
  `midterm_code` varchar(255),
  `final_code` varchar(255),
  `grade_code` varchar(255),
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Topic` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `enrollment_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `teacher_supervisor_code` varchar(255) NOT NULL,
  `grade_defences_code` varchar(255),
  `status` ENUM ('pending', 'approved', 'in_progress', 'completed', 'rejected') NOT NULL,
  `time_start` datetime NOT NULL,
  `time_end` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Council` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Defence` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `council_code` varchar(255) NOT NULL,
  `teacher_code` varchar(255) NOT NULL,
  `position` ENUM ('president', 'secretary', 'reviewer', 'member') NOT NULL
);

CREATE TABLE `Final` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `supervisor_grade` int,
  `reviewer_grade` int,
  `defense_grade` int,
  `final_grade` int,
  `status` ENUM ('pending', 'passed', 'failed', 'completed') NOT NULL,
  `notes` text,
  `completion_date` datetime,
  `created_at` datetime,
  `updated_at` datetime,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `Grade_defences` (
  `id` varchar(255) PRIMARY KEY,
  `council` int,
  `secretary` int,
  `created_at` datetime,
  `updated_at` datetime
);

CREATE TABLE `councils_schedule` (
  `id` varchar(255) PRIMARY KEY,
  `councils_code` varchar(255),
  `topic_code` varchar(255),
  `time_start` datetime,
  `time_end` datetime,
  `created_at` datetime,
  `updated_at` datetime,
  `status` boolean NOT NULL
);

ALTER TABLE `Student` ADD FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`);

ALTER TABLE `Student` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Teacher` ADD FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`);

ALTER TABLE `Teacher` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Major` ADD FOREIGN KEY (`faculty_code`) REFERENCES `Faculty` (`id`);

ALTER TABLE `RoleSystem` ADD FOREIGN KEY (`teacher_code`) REFERENCES `Teacher` (`id`);

ALTER TABLE `RoleSystem` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Enrollment` ADD FOREIGN KEY (`student_code`) REFERENCES `Student` (`id`);

ALTER TABLE `Enrollment` ADD FOREIGN KEY (`midterm_code`) REFERENCES `Midterm` (`id`);

ALTER TABLE `Enrollment` ADD FOREIGN KEY (`final_code`) REFERENCES `Final` (`id`);

ALTER TABLE `Topic` ADD FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`);

ALTER TABLE `Topic` ADD FOREIGN KEY (`enrollment_code`) REFERENCES `Enrollment` (`id`);

ALTER TABLE `Topic` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Topic` ADD FOREIGN KEY (`teacher_supervisor_code`) REFERENCES `Teacher` (`id`);

ALTER TABLE `Council` ADD FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`);

ALTER TABLE `Council` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Defence` ADD FOREIGN KEY (`council_code`) REFERENCES `Council` (`id`);

ALTER TABLE `Defence` ADD FOREIGN KEY (`teacher_code`) REFERENCES `Teacher` (`id`);

ALTER TABLE `councils_schedule` ADD FOREIGN KEY (`councils_code`) REFERENCES `Council` (`id`);

ALTER TABLE `councils_schedule` ADD FOREIGN KEY (`topic_code`) REFERENCES `Topic` (`id`);

ALTER TABLE `Topic` ADD FOREIGN KEY (`grade_defences_code`) REFERENCES `Grade_defences` (`id`);

ALTER TABLE `File` ADD FOREIGN KEY (`table_id`) REFERENCES `Final` (`id`);

ALTER TABLE `File` ADD FOREIGN KEY (`table_id`) REFERENCES `Topic` (`id`);

ALTER TABLE `File` ADD FOREIGN KEY (`table_id`) REFERENCES `Midterm` (`id`);
