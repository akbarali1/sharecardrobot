ALTER TABLE `cards`
    ADD COLUMN `bank_bin_id` INT NOT NULL DEFAULT 0 AFTER `last_four`;

UPDATE `cards` c
SET c.`bank_bin_id` = COALESCE((
    SELECT bb.`id`
    FROM `bank_bins` bb
    WHERE CAST(LEFT(c.`card_number`, CHAR_LENGTH(bb.`bin_code`)) AS BINARY) = CAST(bb.`bin_code` AS BINARY)
    ORDER BY CHAR_LENGTH(bb.`bin_code`) DESC, bb.`id` ASC
    LIMIT 1
), 0)
WHERE c.`bank_bin_id` = 0;

ALTER TABLE `cards`
    ADD KEY `idx_cards_bank_bin_id` (`bank_bin_id`);
