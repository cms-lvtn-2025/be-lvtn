CREATE TABLE `Semester` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Student` (
  `id` varchar(255) PRIMARY KEY,
  `email` varchar(255) NOT NULL,
  `phone` varchar(255) NOT NULL,
  `username` varchar(255) NOT NULL,
  `gender` ENUM ('male', 'female', 'other') NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `class_code` varchar(255),
  `semester_code` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Teacher` (
  `id` varchar(255) PRIMARY KEY,
  `email` varchar(255) NOT NULL,
  `username` varchar(255) NOT NULL,
  `gender` ENUM ('male', 'female', 'other') NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Faculty` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Major` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `faculty_code` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `RoleSystem` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `teacher_code` varchar(255) NOT NULL,
  `role` ENUM ('Academic_affairs_staff', 'Teacher', 'Department_Lecturer') NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `activate` boolean NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `File` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `file` varchar(255) NOT NULL,
  `status` ENUM ('pending', 'approved', 'rejected') NOT NULL,
  `table` ENUM ('topic', 'midterm', 'final', 'order') NOT NULL,
  `option` varchar(255),
  `table_id` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Midterm` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `grade` int,
  `status` ENUM ('not_submitted', 'submitted', 'pass', 'fail') NOT NULL,
  `feedback` text,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Enrollment` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `student_code` varchar(255) NOT NULL,
  `topic_council_code` varchar(255) NOT NULL,
  `final_code` varchar(255),
  `grade_review_code` varchar(255),
  `midterm_code` varchar(255),
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Topic` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `status` ENUM ('submit', 'pending', 'approved_1', 'approved_2', 'in_progress', 'completed', 'rejected') NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `percent_stage_1` int,
  `percent_stage_2` int,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Topic_council` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `stage` ENUM ('stage_dacn', 'stage_lvtn') NOT NULL,
  `topic_code` varchar(255) NOT NULL,
  `council_code` varchar(255),
  `time_start` datetime NOT NULL,
  `time_end` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Topic_council_supervisor` (
  `id` varchar(255) PRIMARY KEY,
  `teacher_supervisor_code` varchar(255) NOT NULL,
  `topic_council_code` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Council` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `time_start` datetime,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Defence` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `council_code` varchar(255) NOT NULL,
  `teacher_code` varchar(255) NOT NULL,
  `position` ENUM ('president', 'secretary', 'reviewer', 'member') NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Final` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `supervisor_grade` int,
  `department_grade` int,
  `final_grade` int,
  `status` ENUM ('pending', 'passed', 'failed', 'completed') NOT NULL,
  `notes` text,
  `completion_date` datetime,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Grade_review` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `review_grade` int,
  `teacher_code` varchar(255) NOT NULL,
  `status` ENUM ('pending', 'passed', 'failed', 'completed') NOT NULL,
  `notes` text,
  `completion_date` datetime,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Grade_defence` (
  `id` varchar(255) PRIMARY KEY,
  `defence_code` varchar(255) NOT NULL,
  `enrollment_code` varchar(255) NOT NULL,
  `note` varchar(255),
  `total_score` int,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

CREATE TABLE `Grade_defence_criterion` (
  `id` varchar(255) PRIMARY KEY,
  `grade_defence_code` varchar(255) NOT NULL,
  `name` varchar(255),
  `score` varchar(255),
  `maxScore` varchar(255),
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
);

ALTER TABLE `Student` ADD FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`);

ALTER TABLE `Student` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Teacher` ADD FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`);

ALTER TABLE `Teacher` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Major` ADD FOREIGN KEY (`faculty_code`) REFERENCES `Faculty` (`id`);

ALTER TABLE `RoleSystem` ADD FOREIGN KEY (`teacher_code`) REFERENCES `Teacher` (`id`);

ALTER TABLE `RoleSystem` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Enrollment` ADD FOREIGN KEY (`student_code`) REFERENCES `Student` (`id`);

ALTER TABLE `Topic` ADD FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`);

ALTER TABLE `Topic` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Council` ADD FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`);

ALTER TABLE `Council` ADD FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

ALTER TABLE `Defence` ADD FOREIGN KEY (`council_code`) REFERENCES `Council` (`id`);

ALTER TABLE `Defence` ADD FOREIGN KEY (`teacher_code`) REFERENCES `Teacher` (`id`);

ALTER TABLE `Topic_council` ADD FOREIGN KEY (`council_code`) REFERENCES `Council` (`id`);

ALTER TABLE `Grade_defence` ADD FOREIGN KEY (`defence_code`) REFERENCES `Defence` (`id`);

ALTER TABLE `Grade_defence_criterion` ADD FOREIGN KEY (`grade_defence_code`) REFERENCES `Grade_defence` (`id`);

ALTER TABLE `Enrollment` ADD FOREIGN KEY (`topic_council_code`) REFERENCES `Topic_council` (`id`);

ALTER TABLE `Topic_council_supervisor` ADD FOREIGN KEY (`teacher_supervisor_code`) REFERENCES `Teacher` (`id`);

ALTER TABLE `Topic_council_supervisor` ADD FOREIGN KEY (`topic_council_code`) REFERENCES `Topic_council` (`id`);

ALTER TABLE `Enrollment` ADD FOREIGN KEY (`final_code`) REFERENCES `Final` (`id`);

ALTER TABLE `Enrollment` ADD FOREIGN KEY (`midterm_code`) REFERENCES `Midterm` (`id`);

ALTER TABLE `Grade_review` ADD FOREIGN KEY (`teacher_code`) REFERENCES `Teacher` (`id`);

ALTER TABLE `Grade_defence` ADD FOREIGN KEY (`enrollment_code`) REFERENCES `Enrollment` (`id`);

ALTER TABLE `Topic_council` ADD FOREIGN KEY (`topic_code`) REFERENCES `Topic` (`id`);
