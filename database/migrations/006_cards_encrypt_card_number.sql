ALTER TABLE `cards`
    MODIFY COLUMN `card_number` VARCHAR(128) NOT NULL;

UPDATE `inline_search_results`
SET `message_text` = '';
