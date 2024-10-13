CREATE TABLE 
`students` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `score` decimal(10,2) NOT NULL,
  `grade` varchar(50) DEFAULT '',
  `email` varchar(100) DEFAULT '',
  `phone_number` varchar(20) DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_name_score` (`name`,`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
