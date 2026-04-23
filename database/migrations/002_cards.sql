CREATE TABLE IF NOT EXISTS `cards`
(
    `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id`            BIGINT UNSIGNED NOT NULL,
    `title`              VARCHAR(255)    NOT NULL,
    `card_number`        VARCHAR(32)     NOT NULL,
    `card_number_masked` VARCHAR(64)     NOT NULL,
    `last_four`          VARCHAR(8)      NOT NULL,
    `created_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_cards_user_number` (`user_id`, `card_number`),
    KEY `idx_cards_user_id` (`user_id`),
    KEY `idx_cards_title` (`title`),
    CONSTRAINT `fk_cards_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
            ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
