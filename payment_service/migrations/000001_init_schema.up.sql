CREATE TABLE `payments` (
  `payment_id` varchar(255) NOT NULL,
  `order_id` varchar(255) DEFAULT NULL,
  `payment_method` varchar(30) DEFAULT NULL,
  `payment_amount` bigint NOT NULL,
  `payment_status` varchar(20) DEFAULT NULL,
  `payment_time` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`payment_id`),
  UNIQUE KEY `order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;