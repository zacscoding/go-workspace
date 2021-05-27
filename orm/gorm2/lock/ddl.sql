DROP TABLE IF EXISTS app_lock;
CREATE TABLE app_lock (
    lock_id VARCHAR(100) PRIMARY KEY,
    expires TIMESTAMP NULL
);

DROP PROCEDURE IF EXISTS `AcquireLock`;
CREATE PROCEDURE `AcquireLock` (
    IN id VARCHAR(100),
    IN ttl INTEGER
)
BEGIN
    DECLARE nowT TIMESTAMP;
    DECLARE expiresT TIMESTAMP;
    SET nowT = now();
    SET expiresT = TIMESTAMPADD(SECOND, ttl, nowT);

    IF EXISTS (SELECT lock_id FROM app_lock WHERE lock_id = id AND expires > nowT) THEN
        SET expiresT = NULL;
    ELSEIF EXISTS(SELECT lock_id FROM app_lock WHERE lock_id = id) THEN
        UPDATE app_lock SET expires = expiresT  WHERE lock_id = id;
    ELSE
        INSERT INTO app_lock (lock_id, expires) VALUES (id, expiresT);
    END IF;

    SELECT expiresT;
END;

DROP PROCEDURE IF EXISTS `ReleaseLock`;
CREATE PROCEDURE `ReleaseLock` (
    IN id VARCHAR(100),
    IN expiresT TIMESTAMP
)
BEGIN
    IF NOT EXISTS (SELECT lock_id FROM app_lock WHERE lock_id = id AND expires = expiresT) THEN
        SELECT false;
    ELSE
        DELETE FROM app_lock WHERE lock_id = id;
        SELECT true;
    END IF;
END;