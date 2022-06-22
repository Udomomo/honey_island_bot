USE honey_island;

DROP TABLE IF EXISTS puzzle;
CREATE TABLE puzzle(
    puzzle_id MEDIUMINT UNSIGNED NOT NULL UNIQUE
);

DROP TABLE IF EXISTS solved_puzzle;
CREATE table solved_puzzle(
    id MEDIUMINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    puzzle_id MEDIUMINT UNSIGNED NOT NULL,
    server_time_sec BIGINT UNSIGNED NOT NULL,
    FOREIGN KEY(puzzle_id) REFERENCES puzzle(puzzle_id)
);
