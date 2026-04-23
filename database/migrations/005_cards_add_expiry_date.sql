ALTER TABLE `cards`
    ADD COLUMN `expiry_date` VARCHAR(5) DEFAULT NULL AFTER `card_number_masked`;
