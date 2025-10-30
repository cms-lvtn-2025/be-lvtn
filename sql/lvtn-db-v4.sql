-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- Máy chủ: mysql:3306
-- Thời gian đã tạo: Th10 28, 2025 lúc 09:22 AM
-- Phiên bản máy phục vụ: 8.0.43
-- Phiên bản PHP: 8.2.27

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Cơ sở dữ liệu: `lvtn-db-v4`
--

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Council`
--

CREATE TABLE `Council` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `time_start` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Defence`
--

CREATE TABLE `Defence` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `council_code` varchar(255) NOT NULL,
  `teacher_code` varchar(255) NOT NULL,
  `position` enum('president','secretary','reviewer','member') NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Enrollment`
--

CREATE TABLE `Enrollment` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `student_code` varchar(255) NOT NULL,
  `topic_council_code` varchar(255) NOT NULL,
  `final_code` varchar(255) DEFAULT NULL,
  `grade_review_code` varchar(255) DEFAULT NULL,
  `midterm_code` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Faculty`
--

CREATE TABLE `Faculty` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `File`
--

CREATE TABLE `File` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `file` varchar(255) NOT NULL,
  `status` enum('approved','rejected','file_pending') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `table` enum('topic','midterm','final','order') NOT NULL,
  `option` varchar(255) DEFAULT NULL,
  `table_id` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Final`
--

CREATE TABLE `Final` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `supervisor_grade` int DEFAULT NULL,
  `department_grade` int DEFAULT NULL,
  `final_grade` int DEFAULT NULL,
  `status` enum('pending','passed','failed','completed') NOT NULL,
  `notes` text,
  `completion_date` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Grade_defence`
--

CREATE TABLE `Grade_defence` (
  `id` varchar(255) NOT NULL,
  `defence_code` varchar(255) NOT NULL,
  `enrollment_code` varchar(255) NOT NULL,
  `note` varchar(255) DEFAULT NULL,
  `total_score` int DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Grade_defence_criterion`
--

CREATE TABLE `Grade_defence_criterion` (
  `id` varchar(255) NOT NULL,
  `grade_defence_code` varchar(255) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `score` varchar(255) DEFAULT NULL,
  `maxScore` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Grade_review`
--

CREATE TABLE `Grade_review` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `review_grade` int DEFAULT NULL,
  `teacher_code` varchar(255) NOT NULL,
  `status` enum('pending','passed','failed','completed') NOT NULL,
  `notes` text,
  `completion_date` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Major`
--

CREATE TABLE `Major` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `faculty_code` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Midterm`
--

CREATE TABLE `Midterm` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `grade` int DEFAULT NULL,
  `status` enum('not_submitted','submitted','pass','fail') NOT NULL,
  `feedback` text,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `RoleSystem`
--

CREATE TABLE `RoleSystem` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `teacher_code` varchar(255) NOT NULL,
  `role` enum('Academic_affairs_staff','Department_Lecturer','Teacher') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `activate` tinyint(1) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Semester`
--

CREATE TABLE `Semester` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Student`
--

CREATE TABLE `Student` (
  `id` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `phone` varchar(255) NOT NULL,
  `username` varchar(255) NOT NULL,
  `gender` enum('male','female','other') NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `class_code` varchar(255) DEFAULT NULL,
  `semester_code` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Teacher`
--

CREATE TABLE `Teacher` (
  `id` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `username` varchar(255) NOT NULL,
  `gender` enum('male','female','other') NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Topic`
--

CREATE TABLE `Topic` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `title_en` varchar(255) NOT NULL DEFAULT 'topic english',
  `description` longtext NOT NULL,
  `major_code` varchar(255) NOT NULL,
  `semester_code` varchar(255) NOT NULL,
  `status` enum('submit','pending','approved_1','approved_2','in_progress','completed','rejected') NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `percent_stage_1` int DEFAULT NULL,
  `percent_stage_2` int DEFAULT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Topic_council`
--

CREATE TABLE `Topic_council` (
  `id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `stage` enum('stage_dacn','stage_lvtn') NOT NULL,
  `topic_code` varchar(255) NOT NULL,
  `council_code` varchar(255) DEFAULT NULL,
  `time_start` datetime NOT NULL,
  `time_end` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Cấu trúc bảng cho bảng `Topic_council_supervisor`
--

CREATE TABLE `Topic_council_supervisor` (
  `id` varchar(255) NOT NULL,
  `teacher_supervisor_code` varchar(255) NOT NULL,
  `topic_council_code` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_by` varchar(255) NOT NULL,
  `updated_by` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Chỉ mục cho các bảng đã đổ
--

--
-- Chỉ mục cho bảng `Council`
--
ALTER TABLE `Council`
  ADD PRIMARY KEY (`id`),
  ADD KEY `major_code` (`major_code`),
  ADD KEY `semester_code` (`semester_code`);

--
-- Chỉ mục cho bảng `Defence`
--
ALTER TABLE `Defence`
  ADD PRIMARY KEY (`id`),
  ADD KEY `council_code` (`council_code`),
  ADD KEY `teacher_code` (`teacher_code`);

--
-- Chỉ mục cho bảng `Enrollment`
--
ALTER TABLE `Enrollment`
  ADD PRIMARY KEY (`id`),
  ADD KEY `student_code` (`student_code`),
  ADD KEY `topic_council_code` (`topic_council_code`),
  ADD KEY `final_code` (`final_code`),
  ADD KEY `midterm_code` (`midterm_code`);

--
-- Chỉ mục cho bảng `Faculty`
--
ALTER TABLE `Faculty`
  ADD PRIMARY KEY (`id`);

--
-- Chỉ mục cho bảng `File`
--
ALTER TABLE `File`
  ADD PRIMARY KEY (`id`);

--
-- Chỉ mục cho bảng `Final`
--
ALTER TABLE `Final`
  ADD PRIMARY KEY (`id`);

--
-- Chỉ mục cho bảng `Grade_defence`
--
ALTER TABLE `Grade_defence`
  ADD PRIMARY KEY (`id`),
  ADD KEY `defence_code` (`defence_code`),
  ADD KEY `enrollment_code` (`enrollment_code`);

--
-- Chỉ mục cho bảng `Grade_defence_criterion`
--
ALTER TABLE `Grade_defence_criterion`
  ADD PRIMARY KEY (`id`),
  ADD KEY `grade_defence_code` (`grade_defence_code`);

--
-- Chỉ mục cho bảng `Grade_review`
--
ALTER TABLE `Grade_review`
  ADD PRIMARY KEY (`id`),
  ADD KEY `teacher_code` (`teacher_code`);

--
-- Chỉ mục cho bảng `Major`
--
ALTER TABLE `Major`
  ADD PRIMARY KEY (`id`),
  ADD KEY `faculty_code` (`faculty_code`);

--
-- Chỉ mục cho bảng `Midterm`
--
ALTER TABLE `Midterm`
  ADD PRIMARY KEY (`id`);

--
-- Chỉ mục cho bảng `RoleSystem`
--
ALTER TABLE `RoleSystem`
  ADD PRIMARY KEY (`id`),
  ADD KEY `teacher_code` (`teacher_code`),
  ADD KEY `semester_code` (`semester_code`);

--
-- Chỉ mục cho bảng `Semester`
--
ALTER TABLE `Semester`
  ADD PRIMARY KEY (`id`);

--
-- Chỉ mục cho bảng `Student`
--
ALTER TABLE `Student`
  ADD PRIMARY KEY (`id`),
  ADD KEY `major_code` (`major_code`),
  ADD KEY `semester_code` (`semester_code`);

--
-- Chỉ mục cho bảng `Teacher`
--
ALTER TABLE `Teacher`
  ADD PRIMARY KEY (`id`),
  ADD KEY `major_code` (`major_code`),
  ADD KEY `semester_code` (`semester_code`);

--
-- Chỉ mục cho bảng `Topic`
--
ALTER TABLE `Topic`
  ADD PRIMARY KEY (`id`),
  ADD KEY `major_code` (`major_code`),
  ADD KEY `semester_code` (`semester_code`);

--
-- Chỉ mục cho bảng `Topic_council`
--
ALTER TABLE `Topic_council`
  ADD PRIMARY KEY (`id`),
  ADD KEY `council_code` (`council_code`),
  ADD KEY `topic_code` (`topic_code`);

--
-- Chỉ mục cho bảng `Topic_council_supervisor`
--
ALTER TABLE `Topic_council_supervisor`
  ADD PRIMARY KEY (`id`),
  ADD KEY `teacher_supervisor_code` (`teacher_supervisor_code`),
  ADD KEY `topic_council_code` (`topic_council_code`);

--
-- Ràng buộc đối với các bảng kết xuất
--

--
-- Ràng buộc cho bảng `Council`
--
ALTER TABLE `Council`
  ADD CONSTRAINT `Council_ibfk_1` FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`),
  ADD CONSTRAINT `Council_ibfk_2` FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

--
-- Ràng buộc cho bảng `Defence`
--
ALTER TABLE `Defence`
  ADD CONSTRAINT `Defence_ibfk_1` FOREIGN KEY (`council_code`) REFERENCES `Council` (`id`),
  ADD CONSTRAINT `Defence_ibfk_2` FOREIGN KEY (`teacher_code`) REFERENCES `Teacher` (`id`);

--
-- Ràng buộc cho bảng `Enrollment`
--
ALTER TABLE `Enrollment`
  ADD CONSTRAINT `Enrollment_ibfk_1` FOREIGN KEY (`student_code`) REFERENCES `Student` (`id`),
  ADD CONSTRAINT `Enrollment_ibfk_2` FOREIGN KEY (`topic_council_code`) REFERENCES `Topic_council` (`id`),
  ADD CONSTRAINT `Enrollment_ibfk_3` FOREIGN KEY (`final_code`) REFERENCES `Final` (`id`),
  ADD CONSTRAINT `Enrollment_ibfk_4` FOREIGN KEY (`midterm_code`) REFERENCES `Midterm` (`id`);

--
-- Ràng buộc cho bảng `Grade_defence`
--
ALTER TABLE `Grade_defence`
  ADD CONSTRAINT `Grade_defence_ibfk_1` FOREIGN KEY (`defence_code`) REFERENCES `Defence` (`id`),
  ADD CONSTRAINT `Grade_defence_ibfk_2` FOREIGN KEY (`enrollment_code`) REFERENCES `Enrollment` (`id`);

--
-- Ràng buộc cho bảng `Grade_defence_criterion`
--
ALTER TABLE `Grade_defence_criterion`
  ADD CONSTRAINT `Grade_defence_criterion_ibfk_1` FOREIGN KEY (`grade_defence_code`) REFERENCES `Grade_defence` (`id`);

--
-- Ràng buộc cho bảng `Grade_review`
--
ALTER TABLE `Grade_review`
  ADD CONSTRAINT `Grade_review_ibfk_1` FOREIGN KEY (`teacher_code`) REFERENCES `Teacher` (`id`);

--
-- Ràng buộc cho bảng `Major`
--
ALTER TABLE `Major`
  ADD CONSTRAINT `Major_ibfk_1` FOREIGN KEY (`faculty_code`) REFERENCES `Faculty` (`id`);

--
-- Ràng buộc cho bảng `RoleSystem`
--
ALTER TABLE `RoleSystem`
  ADD CONSTRAINT `RoleSystem_ibfk_1` FOREIGN KEY (`teacher_code`) REFERENCES `Teacher` (`id`),
  ADD CONSTRAINT `RoleSystem_ibfk_2` FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

--
-- Ràng buộc cho bảng `Student`
--
ALTER TABLE `Student`
  ADD CONSTRAINT `Student_ibfk_1` FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`),
  ADD CONSTRAINT `Student_ibfk_2` FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

--
-- Ràng buộc cho bảng `Teacher`
--
ALTER TABLE `Teacher`
  ADD CONSTRAINT `Teacher_ibfk_1` FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`),
  ADD CONSTRAINT `Teacher_ibfk_2` FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

--
-- Ràng buộc cho bảng `Topic`
--
ALTER TABLE `Topic`
  ADD CONSTRAINT `Topic_ibfk_1` FOREIGN KEY (`major_code`) REFERENCES `Major` (`id`),
  ADD CONSTRAINT `Topic_ibfk_2` FOREIGN KEY (`semester_code`) REFERENCES `Semester` (`id`);

--
-- Ràng buộc cho bảng `Topic_council`
--
ALTER TABLE `Topic_council`
  ADD CONSTRAINT `Topic_council_ibfk_1` FOREIGN KEY (`council_code`) REFERENCES `Council` (`id`),
  ADD CONSTRAINT `Topic_council_ibfk_2` FOREIGN KEY (`topic_code`) REFERENCES `Topic` (`id`);

--
-- Ràng buộc cho bảng `Topic_council_supervisor`
--
ALTER TABLE `Topic_council_supervisor`
  ADD CONSTRAINT `Topic_council_supervisor_ibfk_1` FOREIGN KEY (`teacher_supervisor_code`) REFERENCES `Teacher` (`id`),
  ADD CONSTRAINT `Topic_council_supervisor_ibfk_2` FOREIGN KEY (`topic_council_code`) REFERENCES `Topic_council` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
