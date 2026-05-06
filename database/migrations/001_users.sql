CREATE TABLE IF NOT EXISTS `users`
(
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `telegram_user_id` BIGINT          NOT NULL,
    `username`         VARCHAR(255)             DEFAULT NULL,
    `first_name`       VARCHAR(255)    NOT NULL DEFAULT '',
    `last_name`        VARCHAR(255)             DEFAULT NULL,
    `created_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_users_telegram_user_id` (`telegram_user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
