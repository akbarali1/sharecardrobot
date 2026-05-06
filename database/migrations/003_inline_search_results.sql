CREATE TABLE IF NOT EXISTS `inline_search_results`
(
    `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `result_id`      VARCHAR(64)     NOT NULL,
    `user_id`        BIGINT UNSIGNED NOT NULL,
    `card_id`        BIGINT UNSIGNED NOT NULL,
    `title`          VARCHAR(255)    NOT NULL,
    `description`    VARCHAR(255)    NOT NULL DEFAULT '',
    `message_text`   TEXT            NOT NULL,
    `shown_count`    INT UNSIGNED    NOT NULL DEFAULT 0,
    `chosen_count`   INT UNSIGNED    NOT NULL DEFAULT 0,
    `last_query`     VARCHAR(255)             DEFAULT NULL,
    `last_chosen_at` DATETIME                 DEFAULT NULL,
    `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_inline_search_results_result_id` (`result_id`),
    KEY `idx_inline_search_results_user_id` (`user_id`),
    KEY `idx_inline_search_results_card_id` (`card_id`),
    KEY `idx_inline_search_results_popular` (`user_id`, `chosen_count`, `shown_count`, `updated_at`),
    CONSTRAINT `fk_inline_search_results_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
            ON DELETE CASCADE,
    CONSTRAINT `fk_inline_search_results_card_id`
        FOREIGN KEY (`card_id`) REFERENCES `cards` (`id`)
            ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
