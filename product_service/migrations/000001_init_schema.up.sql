CREATE TABLE `categories` (
  `category_id` varchar(255) NOT NULL,
  `category_name` varchar(100) NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `products` (
  `product_id` varchar(255) NOT NULL,
  `product_code` varchar(50) NOT NULL,
  `product_name` varchar(150) NOT NULL,
  `category_id` varchar(255) NOT NULL,
  `price` bigint NOT NULL,
  `stock` int unsigned NOT NULL DEFAULT '0',
  `description` varchar(500) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`product_id`),
  UNIQUE KEY `product_code` (`product_code`),
  KEY `idx_category` (`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `product_images` (
  `image_id` varchar(255) NOT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `image_url` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`image_id`),
  KEY `idx_product_image` (`product_id`),
  CONSTRAINT `product_images_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `products` (`product_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `product_flash_sales` (
  `flash_sale_id` varchar(255) NOT NULL,
  `product_id` varchar(255) NOT NULL,
  `flash_sale_price` bigint NOT NULL,
  `flash_sale_stock` int NOT NULL,
  `max_purchase_per_user` int DEFAULT '1',
  `start_time` timestamp NOT NULL,
  `end_time` timestamp NOT NULL,
  `version` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`flash_sale_id`),
  UNIQUE KEY `product_id` (`product_id`),
  KEY `idx_product_time` (`product_id`,`start_time`,`end_time`),
  CONSTRAINT `product_flash_sales_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `products` (`product_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
