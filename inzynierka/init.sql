SET FOREIGN_KEY_CHECKS=0;


-- Tworzymy drugą bazę
CREATE DATABASE IF NOT EXISTS kalendarz;

-- Nadajemy Twojemu użytkownikowi pełne prawa do tej bazy
GRANT ALL PRIVILEGES ON kalendarz.* TO 'CZEMPU'@'%';

-- Zapisujemy uprawnienia
FLUSH PRIVILEGES;


USE `RCP`;

--
-- Table structure for table `Nominals`
--

DROP TABLE IF EXISTS `Nominals`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Nominals` (
  `year` int NOT NULL,
  `month` int NOT NULL,
  `hours` int NOT NULL DEFAULT '0',
  `minutes` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`year`,`month`),
  CONSTRAINT `chk_nominals_month` CHECK ((`month` between 1 and 12))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `Nominals`
--

LOCK TABLES `Nominals` WRITE;
/*!40000 ALTER TABLE `Nominals` DISABLE KEYS */;
INSERT INTO `Nominals` VALUES (2025,1,168,10080),(2025,2,160,9600),(2025,3,168,10080),(2025,4,168,10080),(2025,5,168,10080),(2025,6,160,9600),(2025,7,184,11040),(2025,8,160,9600),(2025,9,176,10560),(2025,10,184,11040),(2025,11,152,9120),(2025,12,160,9600),(2026,1,160,9600),(2026,2,160,9600),(2026,3,176,10560),(2026,4,168,10080),(2026,5,160,9600),(2026,6,168,10080),(2026,7,184,11040),(2026,8,168,10080),(2026,9,176,10560),(2026,10,176,10560),(2026,11,160,9600),(2026,12,168,10080);
/*!40000 ALTER TABLE `Nominals` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `approval_policies`
--

DROP TABLE IF EXISTS `approval_policies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `approval_policies` (
  `id` int NOT NULL AUTO_INCREMENT,
  `resource` varchar(64) NOT NULL,
  `conditions` json DEFAULT NULL,
  `workflow_id` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `approval_policies_index_16` (`resource`),
  KEY `workflow_id` (`workflow_id`),
  CONSTRAINT `approval_policies_ibfk_1` FOREIGN KEY (`workflow_id`) REFERENCES `approval_workflows` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `approval_policies`
--

LOCK TABLES `approval_policies` WRITE;
/*!40000 ALTER TABLE `approval_policies` DISABLE KEYS */;
INSERT INTO `approval_policies` VALUES (1,'leave_requests',NULL,1),(2,'sick_leave_requests',NULL,NULL);
/*!40000 ALTER TABLE `approval_policies` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `approval_request_steps`
--

DROP TABLE IF EXISTS `approval_request_steps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `approval_request_steps` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `request_id` bigint NOT NULL,
  `step_no` int NOT NULL,
  `approver_id` int DEFAULT NULL,
  `status` enum('pending','waiting','approved','rejected') NOT NULL DEFAULT 'waiting',
  `decided_at` datetime DEFAULT NULL,
  `note` text,
  `approver_type` varchar(50) NOT NULL DEFAULT 'manager_of_requester',
  `approver_org_unit_id` int DEFAULT NULL,
  `approver_role_id` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_req_step_approver` (`request_id`,`step_no`,`approver_id`),
  KEY `ix_ars_req_step` (`request_id`,`step_no`),
  KEY `ix_ars_approver` (`approver_id`),
  CONSTRAINT `approval_request_steps_ibfk_1` FOREIGN KEY (`request_id`) REFERENCES `approval_requests` (`id`),
  CONSTRAINT `approval_request_steps_ibfk_2` FOREIGN KEY (`approver_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=201 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `approval_request_steps`
--

LOCK TABLES `approval_request_steps` WRITE;
/*!40000 ALTER TABLE `approval_request_steps` DISABLE KEYS */;
INSERT INTO `approval_request_steps` VALUES (1,1,1,1,'approved','2026-01-03 06:53:09',NULL,'manager_of_requester',NULL,NULL),(2,2,1,1,'approved','2026-01-04 09:11:58',NULL,'manager_of_requester',NULL,NULL);
/*!40000 ALTER TABLE `approval_request_steps` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `approval_requests`
--

DROP TABLE IF EXISTS `approval_requests`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `approval_requests` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `resource` varchar(64) NOT NULL,
  `target_id` bigint NOT NULL,
  `requester_id` int NOT NULL,
  `status` enum('pending','approved','rejected','cancelled') NOT NULL,
  `created_at` datetime NOT NULL,
  `closed_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_resource_target` (`resource`,`target_id`),
  KEY `ix_ar_requester` (`requester_id`),
  KEY `ix_ar_status` (`status`),
  CONSTRAINT `approval_requests_ibfk_1` FOREIGN KEY (`requester_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=116 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `approval_requests`
--

LOCK TABLES `approval_requests` WRITE;
/*!40000 ALTER TABLE `approval_requests` DISABLE KEYS */;
INSERT INTO `approval_requests` VALUES (1,'leave_requests',1,2,'approved','2026-01-03 06:12:13','2026-01-03 06:53:09'),(2,'leave_requests',2,1,'approved','2026-01-04 09:10:56','2026-01-04 09:11:58');
/*!40000 ALTER TABLE `approval_requests` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `approval_workflow_steps`
--

DROP TABLE IF EXISTS `approval_workflow_steps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `approval_workflow_steps` (
  `workflow_id` int NOT NULL,
  `step_no` int NOT NULL,
  `completion_rule` varchar(8) NOT NULL,
  `approver_type` enum('manager_of_requester','manager_of_unit','user_of_unit','specific_user','role_global') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `slug` varchar(64) DEFAULT NULL,
  `levels_up` int DEFAULT NULL,
  `org_scope` enum('requester_unit','unit_chain_up','whole_org') DEFAULT NULL,
  `required` tinyint(1) NOT NULL,
  PRIMARY KEY (`workflow_id`,`step_no`),
  UNIQUE KEY `uq_workflow_step` (`workflow_id`,`step_no`),
  CONSTRAINT `approval_workflow_steps_ibfk_1` FOREIGN KEY (`workflow_id`) REFERENCES `approval_workflows` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `approval_workflow_steps`
--

LOCK TABLES `approval_workflow_steps` WRITE;
/*!40000 ALTER TABLE `approval_workflow_steps` DISABLE KEYS */;
INSERT INTO `approval_workflow_steps` VALUES (1,1,'all','specific_user','1',0,'requester_unit',1),(1,2,'all','manager_of_unit','2',0,'requester_unit',1),(1,3,'all','role_global','19',0,'requester_unit',1),(1,4,'all','manager_of_requester',NULL,0,'requester_unit',1),(1,5,'all','manager_of_unit','1',0,'requester_unit',1);
/*!40000 ALTER TABLE `approval_workflow_steps` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `approval_workflows`
--

DROP TABLE IF EXISTS `approval_workflows`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `approval_workflows` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(160) NOT NULL,
  `resource` varchar(64) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `approval_workflows`
--

LOCK TABLES `approval_workflows` WRITE;
/*!40000 ALTER TABLE `approval_workflows` DISABLE KEYS */;
INSERT INTO `approval_workflows` VALUES (1,'Standardowy obieg urlopowy','leave_requests');
/*!40000 ALTER TABLE `approval_workflows` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `calendar_holidays`
--

DROP TABLE IF EXISTS `calendar_holidays`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `calendar_holidays` (
  `id` int NOT NULL AUTO_INCREMENT,
  `work_calendar_id` int NOT NULL,
  `day` date NOT NULL,
  `name` varchar(120) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_cal_day` (`work_calendar_id`,`day`),
  CONSTRAINT `calendar_holidays_ibfk_1` FOREIGN KEY (`work_calendar_id`) REFERENCES `work_calendars` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=30 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `calendar_holidays`
--

LOCK TABLES `calendar_holidays` WRITE;
/*!40000 ALTER TABLE `calendar_holidays` DISABLE KEYS */;
INSERT INTO `calendar_holidays` VALUES (2,1,'2025-01-01','Nowy Rok'),(3,1,'2025-01-06','Trzech Króli'),(4,1,'2025-04-20','Wielkanoc'),(5,1,'2025-04-21','Poniedziałek Wielkanocny'),(6,1,'2025-05-01','Święto Pracy'),(7,1,'2025-05-03','Święto Konstytucji 3 Maja'),(8,1,'2025-06-08','Zielone Świątki'),(9,1,'2025-06-19','Boże Ciało'),(10,1,'2025-08-15','Wniebowzięcie NMP'),(11,1,'2025-11-01','Wszystkich Świętych'),(12,1,'2025-11-11','Narodowe Święto Niepodległości'),(13,1,'2025-12-24','Wigilia'),(14,1,'2025-12-25','Boże Narodzenie (dzień pierwszy)'),(15,1,'2025-12-26','Boże Narodzenie (dzień drugi)'),(16,1,'2026-01-01','Nowy Rok'),(17,1,'2026-01-06','Trzech Króli'),(18,1,'2026-04-05','Wielkanoc'),(19,1,'2026-04-06','Poniedziałek Wielkanocny'),(20,1,'2026-05-01','Święto Pracy'),(21,1,'2026-05-03','Święto Konstytucji 3 Maja'),(22,1,'2026-05-24','Zielone Świątki'),(23,1,'2026-06-04','Boże Ciało'),(24,1,'2026-08-15','Wniebowzięcie NMP'),(25,1,'2026-11-01','Wszystkich Świętych'),(26,1,'2026-11-11','Narodowe Święto Niepodległości'),(27,1,'2026-12-24','Wigilia'),(28,1,'2026-12-25','Boże Narodzenie (dzień pierwszy)'),(29,1,'2026-12-26','Boże Narodzenie (dzień drugi)');
/*!40000 ALTER TABLE `calendar_holidays` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `employee_assignments`
--

DROP TABLE IF EXISTS `employee_assignments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `employee_assignments` (
  `id` int NOT NULL,
  `employee_id` int NOT NULL,
  `work_calendar_id` int NOT NULL,
  `flex_policy_id` int NOT NULL,
  `effective_from` date NOT NULL,
  `effective_to` date DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_emp_assign` (`employee_id`,`effective_from`,`effective_to`),
  CONSTRAINT `employee_assignments_ibfk_1` FOREIGN KEY (`employee_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `employee_assignments`
--

LOCK TABLES `employee_assignments` WRITE;
/*!40000 ALTER TABLE `employee_assignments` DISABLE KEYS */;
INSERT INTO `employee_assignments` VALUES (1,1,1,1,'2025-01-01',NULL),(2,2,1,1,'2025-01-01',NULL);
/*!40000 ALTER TABLE `employee_assignments` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `employee_sickness`
--

DROP TABLE IF EXISTS `employee_sickness`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `employee_sickness` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `employee_id` int NOT NULL,
  `from_date` date NOT NULL,
  `to_date` date NOT NULL,
  `legal_basis` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `certificate_number` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `percent_coverage` tinyint unsigned NOT NULL DEFAULT '80',
  `is_occupational_accident` tinyint(1) NOT NULL DEFAULT '0',
  `status` enum('pending','approved','rejected','cancelled') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending',
  `reported_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_employee_sickness_emp_date` (`employee_id`,`from_date`,`to_date`),
  KEY `idx_employee_sickness_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `employee_sickness`
--

LOCK TABLES `employee_sickness` WRITE;
/*!40000 ALTER TABLE `employee_sickness` DISABLE KEYS */;
INSERT INTO `employee_sickness` VALUES (1,2,'2026-01-20','2026-01-22','Art. 33','ZLA/1234/2026',80,0,'approved','2026-01-20 08:00:00','2026-01-20 08:00:00','2026-01-20 08:00:00');
/*!40000 ALTER TABLE `employee_sickness` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `employment_terms`
--

DROP TABLE IF EXISTS `employment_terms`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `employment_terms` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `effective_from` date NOT NULL,
  `effective_to` date DEFAULT NULL,
  `fte_percent` decimal(5,2) NOT NULL,
  `position` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `ix_terms_rng` (`user_id`,`effective_from`,`effective_to`),
  CONSTRAINT `employment_terms_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `employment_terms`
--

LOCK TABLES `employment_terms` WRITE;
/*!40000 ALTER TABLE `employment_terms` DISABLE KEYS */;
INSERT INTO `employment_terms` VALUES (1,1,'2025-01-01',NULL,1.00,'Administrator IT'),(2,2,'2025-01-01',NULL,1.00,'Specjalista IT');
/*!40000 ALTER TABLE `employment_terms` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `leave_request_items`
--

DROP TABLE IF EXISTS `leave_request_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `leave_request_items` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `request_id` bigint NOT NULL,
  `start_at` datetime(6) NOT NULL,
  `end_at` datetime(6) NOT NULL,
  `unit` enum('day','hour','minute') NOT NULL,
  `quantity` decimal(8,2) NOT NULL,
  `effective_minutes` int NOT NULL,
  `note` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_li_req` (`request_id`),
  KEY `idx_li_req_time` (`request_id`,`start_at`,`end_at`),
  CONSTRAINT `fk_li_req` FOREIGN KEY (`request_id`) REFERENCES `leave_requests` (`id`) ON DELETE CASCADE,
  CONSTRAINT `leave_request_items_ibfk_1` FOREIGN KEY (`request_id`) REFERENCES `leave_requests` (`id`),
  CONSTRAINT `chk_li_duration` CHECK ((`end_at` > `start_at`))
) ENGINE=InnoDB AUTO_INCREMENT=108 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `leave_request_items`
--

LOCK TABLES `leave_request_items` WRITE;
/*!40000 ALTER TABLE `leave_request_items` DISABLE KEYS */;
INSERT INTO `leave_request_items` VALUES (1,1,'2026-01-15 00:00:00.000000','2026-01-16 23:59:00.000000','day',2.00,2880,'Urlop wypoczynkowy'),(2,2,'2026-01-05 08:00:00.000000','2026-01-05 16:00:00.000000','hour',8.00,480,'Urlop na żądanie');
/*!40000 ALTER TABLE `leave_request_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `leave_requests`
--

DROP TABLE IF EXISTS `leave_requests`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `leave_requests` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `employee_id` int NOT NULL,
  `type_code` varchar(8) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_lr_emp` (`employee_id`,`created_at`),
  KEY `type_code` (`type_code`),
  CONSTRAINT `fk_lr_type` FOREIGN KEY (`type_code`) REFERENCES `leave_types` (`code`),
  CONSTRAINT `fk_lr_user` FOREIGN KEY (`employee_id`) REFERENCES `users` (`id`),
  CONSTRAINT `leave_requests_ibfk_1` FOREIGN KEY (`employee_id`) REFERENCES `users` (`id`),
  CONSTRAINT `leave_requests_ibfk_2` FOREIGN KEY (`type_code`) REFERENCES `leave_types` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=83 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `leave_requests`
--

LOCK TABLES `leave_requests` WRITE;
/*!40000 ALTER TABLE `leave_requests` DISABLE KEYS */;
INSERT INTO `leave_requests` VALUES (1,2,'vacay','2026-01-03 06:12:13'),(2,1,'ondmd','2026-01-04 09:10:56');
/*!40000 ALTER TABLE `leave_requests` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `leave_types`
--

DROP TABLE IF EXISTS `leave_types`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `leave_types` (
  `code` varchar(8) NOT NULL,
  `name` varchar(120) NOT NULL,
  `affects_balance` tinyint(1) NOT NULL DEFAULT '1',
  `count_strategy` enum('business','calendar') NOT NULL DEFAULT 'business',
  `allow_days` tinyint(1) NOT NULL DEFAULT '1',
  `allow_hours` tinyint(1) NOT NULL DEFAULT '1',
  `allow_minutes` tinyint(1) NOT NULL DEFAULT '0',
  `min_inc_days` int NOT NULL DEFAULT '1',
  `min_inc_hours` int NOT NULL DEFAULT '60',
  `min_inc_minutes` int NOT NULL DEFAULT '1',
  `nominal_day_minutes` int DEFAULT NULL,
  `approval_policies_id` int NOT NULL,
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `leave_types`
--

LOCK TABLES `leave_types` WRITE;
/*!40000 ALTER TABLE `leave_types` DISABLE KEYS */;
INSERT INTO `leave_types` VALUES ('brvmt','Urlop okolicznościowy',1,'business',1,1,0,1,60,1,0,1),('matrn','Urlop macierzyński',0,'business',1,1,0,1,60,1,0,1),('ondmd','Urlop na żądanie',1,'business',1,1,0,1,60,1,0,1),('prntl','Urlop wychowawczy',0,'business',1,1,0,1,60,1,0,1),('ptnry','Urlop ojcowski',0,'business',1,1,0,1,60,1,0,1),('unpld','Urlop bezpłatny',0,'business',1,1,0,1,60,1,0,1),('vacay','Urlop wypoczynkowy',1,'business',1,1,0,1,60,1,0,1);
/*!40000 ALTER TABLE `leave_types` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `module`
--

DROP TABLE IF EXISTS `module`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `module` (
  `id` int NOT NULL AUTO_INCREMENT,
  `module_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `description` varchar(256) NOT NULL,
  `icon` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `module`
--

LOCK TABLES `module` WRITE;
/*!40000 ALTER TABLE `module` DISABLE KEYS */;
INSERT INTO `module` VALUES (1,'dashboard','Panel Główny','/public/lib/img/home.svg'),(2,'schedule','Grafik','/public/lib/img/calendar-clock.svg'),(3,'leaves','Urlopy','/public/lib/img/document.svg'),(4,'approvals','Akceptacje','/public/lib/img/checkbox.svg'),(5,'organization','Organizacja','/public/lib/img/building.svg'),(6,'users','Użytkownicy','/public/lib/img/users.svg'),(7,'permissions','Uprawnienia','/public/lib/img/unlock.svg'),(9,'timepolicy','Polityka Czasu','/public/lib/img/database.svg'),(10,'sick','Choroby Pracownika','/public/lib/img/stethoscope.svg'),(11,'workflow','Obieg Akceptacji','/public/lib/img/workflow-setting-alt.svg'),(12,'docs','Dokumentacja','/public/lib/img/settings.svg');
/*!40000 ALTER TABLE `module` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `org_unit_managers`
--

DROP TABLE IF EXISTS `org_unit_managers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `org_unit_managers` (
  `org_unit_id` int NOT NULL,
  `user_id` int NOT NULL,
  `since_date` date DEFAULT NULL,
  `until_date` date DEFAULT NULL,
  `is_default` tinyint(1) NOT NULL,
  PRIMARY KEY (`org_unit_id`,`user_id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `org_unit_managers_ibfk_1` FOREIGN KEY (`org_unit_id`) REFERENCES `org_units` (`id`),
  CONSTRAINT `org_unit_managers_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `org_unit_managers`
--

LOCK TABLES `org_unit_managers` WRITE;
/*!40000 ALTER TABLE `org_unit_managers` DISABLE KEYS */;
INSERT INTO `org_unit_managers` VALUES (1,1,'2026-01-05',NULL,1);
/*!40000 ALTER TABLE `org_unit_managers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `org_units`
--

DROP TABLE IF EXISTS `org_units`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `org_units` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(160) NOT NULL,
  `parent_id` int DEFAULT NULL,
  `code` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `org_units_index_14` (`code`),
  KEY `org_units_index_13` (`parent_id`),
  CONSTRAINT `org_units_ibfk_1` FOREIGN KEY (`parent_id`) REFERENCES `org_units` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `org_units`
--

LOCK TABLES `org_units` WRITE;
/*!40000 ALTER TABLE `org_units` DISABLE KEYS */;
INSERT INTO `org_units` VALUES (1,'IT',14,'DEPT-ITT'),(2,'HR',14,'DEPT-HR'),(4,'Księgowość',14,'DEPT-AC'),(5,'Sprzedaż',4,'DEPT-SL'),(14,'Zarząd',NULL,'dasdas');
/*!40000 ALTER TABLE `org_units` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `permissions`
--

DROP TABLE IF EXISTS `permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permissions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `resource` varchar(64) NOT NULL,
  `action` varchar(32) NOT NULL,
  `scope` enum('own','team','org') NOT NULL,
  `conditions` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `ix_perm_triplet` (`resource`,`action`,`scope`)
) ENGINE=InnoDB AUTO_INCREMENT=138 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permissions`
--

LOCK TABLES `permissions` WRITE;
/*!40000 ALTER TABLE `permissions` DISABLE KEYS */;
INSERT INTO `permissions` VALUES (1,'schedule','read','own','{}'),(4,'users','read','own','{}'),(5,'schedule','read','team','{}'),(8,'users','read','team','{}'),(9,'leaves','read','team','{}'),(10,'schedule','read','org','{}'),(11,'schedule','update','org','{}'),(12,'users','read','org','{}'),(13,'leaves','read','org','{}'),(14,'organization','read','own','{}'),(19,'*','*','org','{}'),(20,'users','update','own','{}'),(21,'users','update','team','{}'),(22,'users','update','org','{}'),(25,'users','create','org','{}'),(38,'permissions','update','org','{}'),(42,'permissions','read','org','{}'),(54,'organization','update','org','{}'),(58,'organization','read','org','{}'),(59,'leaves','update','own','{}'),(65,'leaves','read','own','{}'),(84,'approvals','update','org','{}'),(89,'approvals','read','org','{}'),(96,'dashboard','read','own','{}'),(101,'workflow','update','org','{}'),(106,'workflow','read','org','{}'),(110,'timepolicy','update','org','{}'),(115,'timepolicy','read','org','{}'),(119,'sick','update','org','{}'),(123,'sick','read','own','{}'),(124,'sick','read','org','{}'),(136,'schedule','time','own','{}'),(137,'docs','read','own','{}');
/*!40000 ALTER TABLE `permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `role_permissions`
--

DROP TABLE IF EXISTS `role_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `role_permissions` (
  `role_id` int NOT NULL,
  `permission_id` int NOT NULL,
  PRIMARY KEY (`role_id`,`permission_id`),
  KEY `permission_id` (`permission_id`),
  CONSTRAINT `fk_rp_perm` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_rp_role` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `role_permissions_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`),
  CONSTRAINT `role_permissions_ibfk_2` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `role_permissions`
--

LOCK TABLES `role_permissions` WRITE;
/*!40000 ALTER TABLE `role_permissions` DISABLE KEYS */;
INSERT INTO `role_permissions` VALUES (1,1),(6,1),(1,4),(6,4),(6,5),(1,8),(6,8),(15,8),(6,9),(6,10),(14,10),(15,10),(6,11),(14,11),(6,12),(14,12),(15,12),(6,13),(14,13),(15,13),(1,14),(6,14),(14,14),(7,19),(6,20),(6,21),(15,21),(6,22),(14,22),(1,25),(6,25),(14,25),(6,38),(6,42),(15,42),(6,54),(14,54),(6,58),(14,58),(15,58),(1,59),(6,59),(14,59),(1,65),(6,65),(1,84),(6,84),(14,84),(15,84),(1,89),(6,89),(14,89),(15,89),(1,96),(6,96),(14,96),(15,96),(6,101),(6,106),(15,106),(6,110),(1,115),(6,115),(15,115),(6,119),(14,119),(1,123),(6,123),(6,124),(14,124),(15,124),(1,136),(6,136),(14,136),(1,137),(6,137),(15,137);
/*!40000 ALTER TABLE `role_permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `roles`
--

DROP TABLE IF EXISTS `roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `roles` (
  `id` int NOT NULL AUTO_INCREMENT,
  `slug` varchar(64) NOT NULL,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `slug` (`slug`)
) ENGINE=InnoDB AUTO_INCREMENT=20 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `roles`
--

LOCK TABLES `roles` WRITE;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;
INSERT INTO `roles` VALUES (1,'employee','Pracownik'),(6,'admin','Admin'),(7,'superadmin','Super Admin'),(14,'hr','HR'),(15,'custom','Uprawnienia Testowe'),(19,'dsadas','dadsa');
/*!40000 ALTER TABLE `roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `seq_0_to_31`
--

DROP TABLE IF EXISTS `seq_0_to_31`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `seq_0_to_31` (
  `seq` int NOT NULL,
  PRIMARY KEY (`seq`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `seq_0_to_31`
--

LOCK TABLES `seq_0_to_31` WRITE;
/*!40000 ALTER TABLE `seq_0_to_31` DISABLE KEYS */;
INSERT INTO `seq_0_to_31` VALUES (0),(1),(2),(3),(4),(5),(6),(7),(8),(9),(10),(11),(12),(13),(14),(15),(16),(17),(18),(19),(20),(21),(22),(23),(24),(25),(26),(27),(28),(29),(30),(31);
/*!40000 ALTER TABLE `seq_0_to_31` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `time_policies`
--

DROP TABLE IF EXISTS `time_policies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `time_policies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `effective_from` date NOT NULL,
  `effective_to` date DEFAULT NULL,
  `earliest_start` time DEFAULT NULL,
  `latest_start` time DEFAULT NULL,
  `latest_end` time DEFAULT NULL,
  `flex` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ux_time_policies_effective_range` (`effective_from`,`effective_to`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `time_policies`
--

LOCK TABLES `time_policies` WRITE;
/*!40000 ALTER TABLE `time_policies` DISABLE KEYS */;
INSERT INTO `time_policies` VALUES (5,'2025-01-01',NULL,'08:00:00','08:00:00','16:00:00',1,'2026-02-11 05:00:41','2026-02-11 05:02:39');
/*!40000 ALTER TABLE `time_policies` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `timesheet_entries`
--

DROP TABLE IF EXISTS `timesheet_entries`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `timesheet_entries` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `started_at` datetime NOT NULL,
  `ended_at` datetime DEFAULT NULL,
  `type` enum('work','break','overtime') NOT NULL DEFAULT 'work',
  `note` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_start` (`user_id`,`started_at`),
  CONSTRAINT `timesheet_entries_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `chk_times` CHECK (((`ended_at` is null) or (`ended_at` > `started_at`)))
) ENGINE=InnoDB AUTO_INCREMENT=303 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `timesheet_entries`
--

LOCK TABLES `timesheet_entries` WRITE;
/*!40000 ALTER TABLE `timesheet_entries` DISABLE KEYS */;
INSERT INTO `timesheet_entries` VALUES (1,1,'2026-02-10 08:00:00','2026-02-10 16:00:00','work','','2026-02-10 08:00:00','2026-02-10 16:00:00'),(2,1,'2026-02-11 08:00:00','2026-02-11 16:00:00','work','','2026-02-11 08:00:00','2026-02-11 16:00:00'),(3,2,'2026-02-10 08:00:00','2026-02-10 16:00:00','work','','2026-02-10 08:00:00','2026-02-10 16:00:00'),(4,2,'2026-02-11 08:00:00','2026-02-11 16:00:00','work','','2026-02-11 08:00:00','2026-02-11 16:00:00');
/*!40000 ALTER TABLE `timesheet_entries` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_org_memberships`
--

DROP TABLE IF EXISTS `user_org_memberships`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_org_memberships` (
  `user_id` int NOT NULL,
  `org_unit_id` int NOT NULL,
  `is_primary` tinyint(1) NOT NULL,
  `title` varchar(160) DEFAULT NULL,
  `since_date` date DEFAULT NULL,
  `until_date` date DEFAULT NULL,
  PRIMARY KEY (`user_id`),
  KEY `user_org_memberships_index_15` (`org_unit_id`),
  CONSTRAINT `user_org_memberships_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_org_memberships`
--

LOCK TABLES `user_org_memberships` WRITE;
/*!40000 ALTER TABLE `user_org_memberships` DISABLE KEYS */;
INSERT INTO `user_org_memberships` VALUES (1,1,1,'Dyrektor IT','2025-12-28',NULL),(2,1,1,'Pracownik IT','2025-12-28',NULL);
/*!40000 ALTER TABLE `user_org_memberships` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_roles`
--

DROP TABLE IF EXISTS `user_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_roles` (
  `user_id` int NOT NULL,
  `role_id` int NOT NULL,
  `team_id` int DEFAULT NULL,
  PRIMARY KEY (`user_id`,`role_id`),
  UNIQUE KEY `uq_user` (`user_id`),
  KEY `ix_ur_user` (`user_id`),
  KEY `role_id` (`role_id`),
  CONSTRAINT `fk_ur_role` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ur_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `user_roles_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `user_roles_ibfk_2` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_roles`
--

LOCK TABLES `user_roles` WRITE;
/*!40000 ALTER TABLE `user_roles` DISABLE KEYS */;
INSERT INTO `user_roles` VALUES (1,7,NULL),(2,1,NULL);
/*!40000 ALTER TABLE `user_roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_schedule_overrides`
--

DROP TABLE IF EXISTS `user_schedule_overrides`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_schedule_overrides` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `day` date NOT NULL,
  `planned_minutes` int NOT NULL,
  `earliest_start` time DEFAULT NULL,
  `latest_start` time DEFAULT NULL,
  `latest_end` time DEFAULT NULL,
  `start_time` time DEFAULT NULL,
  `end_time` time DEFAULT NULL,
  `spans_midnight` tinyint(1) NOT NULL DEFAULT '0',
  `flex` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_user_day` (`user_id`,`day`),
  CONSTRAINT `user_schedule_overrides_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=56 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_schedule_overrides`
--

LOCK TABLES `user_schedule_overrides` WRITE;
/*!40000 ALTER TABLE `user_schedule_overrides` DISABLE KEYS */;
INSERT INTO `user_schedule_overrides` VALUES (1,1,'2026-02-10',0,NULL,NULL,NULL,NULL,NULL,0,1),(2,2,'2026-02-10',0,NULL,NULL,NULL,NULL,NULL,0,1);
/*!40000 ALTER TABLE `user_schedule_overrides` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_schedules_weekly`
--

DROP TABLE IF EXISTS `user_schedules_weekly`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_schedules_weekly` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `effective_from` date NOT NULL,
  `effective_to` date DEFAULT NULL,
  `dow` tinyint NOT NULL,
  `planned_minutes` int NOT NULL,
  `earliest_start` time DEFAULT NULL,
  `latest_start` time DEFAULT NULL,
  `latest_end` time DEFAULT NULL,
  `spans_midnight` tinyint(1) NOT NULL DEFAULT '0',
  `flex` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_usw_user_rng` (`user_id`,`effective_from`,`effective_to`,`dow`),
  CONSTRAINT `user_schedules_weekly_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=94 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_schedules_weekly`
--

LOCK TABLES `user_schedules_weekly` WRITE;
/*!40000 ALTER TABLE `user_schedules_weekly` DISABLE KEYS */;
INSERT INTO `user_schedules_weekly` VALUES (1,1,'2025-01-01',NULL,1,480,'08:00:00','09:00:00','17:00:00',0,1),(2,1,'2025-01-01',NULL,2,480,'08:00:00','09:00:00','17:00:00',0,1),(3,1,'2025-01-01',NULL,3,480,'08:00:00','09:00:00','17:00:00',0,1),(4,1,'2025-01-01',NULL,4,480,'08:00:00','09:00:00','17:00:00',0,1),(5,1,'2025-01-01',NULL,5,480,'08:00:00','09:00:00','17:00:00',0,1),(6,1,'2025-01-01',NULL,6,0,NULL,NULL,NULL,0,1),(7,1,'2025-01-01',NULL,7,0,NULL,NULL,NULL,0,1),(8,2,'2025-01-01',NULL,1,480,'08:00:00','09:00:00','17:00:00',0,1),(9,2,'2025-01-01',NULL,2,480,'08:00:00','09:00:00','17:00:00',0,1),(10,2,'2025-01-01',NULL,3,480,'08:00:00','09:00:00','17:00:00',0,1),(11,2,'2025-01-01',NULL,4,480,'08:00:00','09:00:00','17:00:00',0,1),(12,2,'2025-01-01',NULL,5,480,'08:00:00','09:00:00','17:00:00',0,1),(13,2,'2025-01-01',NULL,6,0,NULL,NULL,NULL,0,1),(14,2,'2025-01-01',NULL,7,0,NULL,NULL,NULL,0,1);
/*!40000 ALTER TABLE `user_schedules_weekly` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL,
  `email` varchar(100) NOT NULL,
  `password` varchar(255) NOT NULL,
  `is_employee` tinyint(1) NOT NULL DEFAULT '1',
  `country_code` char(2) NOT NULL DEFAULT 'PL',
  `nominal_day_minutes` int NOT NULL DEFAULT '480',
  `timezone` varchar(64) NOT NULL DEFAULT 'Europe/Warsaw',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(100) NOT NULL DEFAULT '',
  `surname` varchar(100) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_users_username` (`username`),
  KEY `idx_users_email` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'admin','admin@example.com','$2a$10$tkmDbMd.VL.9A6H0uA9Juurxe4m2Bj7neXiHpnwDL787/YqntUK/G',1,'PL',480,'Europe/Warsaw','2025-09-06 08:36:37','2026-02-09 22:48:24','Administrator','Systemu'),(2,'user','user@example.com','$2a$10$3gE16ARJyCpBIIxGFKn3de.awo8N3haNLTPcmZUlmrnXcaXkbb0OW',1,'PL',480,'Europe/Warsaw','2025-10-30 01:17:16','2026-01-25 17:33:00','Jan','Kowalski');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Temporary view structure for view `v_user_permissions`
--

DROP TABLE IF EXISTS `v_user_permissions`;
/*!50001 DROP VIEW IF EXISTS `v_user_permissions`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `v_user_permissions` AS SELECT 
 1 AS `user_id`,
 1 AS `resource`,
 1 AS `action`,
 1 AS `scope`,
 1 AS `conditions`*/;
SET character_set_client = @saved_cs_client;

--
-- Table structure for table `whitelist`
--

DROP TABLE IF EXISTS `whitelist`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `whitelist` (
  `IP` varchar(15) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `whitelist`
--

LOCK TABLES `whitelist` WRITE;
/*!40000 ALTER TABLE `whitelist` DISABLE KEYS */;
INSERT INTO `whitelist` VALUES ('127.0.0.1');
/*!40000 ALTER TABLE `whitelist` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `work_calendars`
--

DROP TABLE IF EXISTS `work_calendars`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `work_calendars` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `timezone` varchar(128) NOT NULL,
  `is_default` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `work_calendars`
--

LOCK TABLES `work_calendars` WRITE;
/*!40000 ALTER TABLE `work_calendars` DISABLE KEYS */;
INSERT INTO `work_calendars` VALUES (1,'Standard PL','Europe/Warsaw',1);
/*!40000 ALTER TABLE `work_calendars` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Current Database: `kalendarz`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `kalendarz` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;

USE `kalendarz`;

--
-- Table structure for table `calendar`
--

DROP TABLE IF EXISTS `calendar`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `calendar` (
  `person` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `date_str` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` int NOT NULL,
  PRIMARY KEY (`person`,`date_str`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `calendar`
--

LOCK TABLES `calendar` WRITE;
/*!40000 ALTER TABLE `calendar` DISABLE KEYS */;
INSERT INTO `calendar` VALUES ('Jan','2026-04-14',0),('Administrator','2026-04-14',0);
/*!40000 ALTER TABLE `calendar` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tickets`
--

DROP TABLE IF EXISTS `tickets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tickets` (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `assignees` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tickets`
--

LOCK TABLES `tickets` WRITE;
/*!40000 ALTER TABLE `tickets` DISABLE KEYS */;
INSERT INTO `tickets` VALUES (1,'Wdrożenie modułu rejestracji czasu pracy','','Zadanie','realizowany','Administrator'),(2,'Przegląd wniosków urlopowych','','Zadanie','nowe','Jan');
/*!40000 ALTER TABLE `tickets` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Current Database: `RCP`
--

USE `RCP`;

--
-- Final view structure for view `v_user_permissions`
--

/*!50001 DROP VIEW IF EXISTS `v_user_permissions`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_0900_ai_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=`CZEMPU`@`%` SQL SECURITY DEFINER */
/*!50001 VIEW `v_user_permissions` AS select `ur`.`user_id` AS `user_id`,`p`.`resource` AS `resource`,`p`.`action` AS `action`,`p`.`scope` AS `scope`,`p`.`conditions` AS `conditions` from ((`user_roles` `ur` join `role_permissions` `rp` on((`rp`.`role_id` = `ur`.`role_id`))) join `permissions` `p` on((`p`.`id` = `rp`.`permission_id`))) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;

--
-- Current Database: `kalendarz`
--

USE `kalendarz`;

-- Dump completed on 2026-07-28 20:22:47
