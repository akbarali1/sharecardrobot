ALTER TABLE `cards`
    MODIFY COLUMN `card_number` VARCHAR(128) NOT NULL,
    MODIFY COLUMN `expiry_date` VARCHAR(128) DEFAULT NULL;

UPDATE `inline_search_results`
SET `message_text` = '';
