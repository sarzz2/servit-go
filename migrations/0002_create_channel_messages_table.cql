CREATE TABLE IF NOT EXISTS messaging.channel_messages (
    channel_id UUID,
    message_date text,  -- e.g. "2023-06-15"
    timestamp TIMESTAMP,
    message_id UUID,
    sender_id UUID,
    sender_username text,
    content text,
    PRIMARY KEY ((channel_id, message_date), timestamp, message_id)
) WITH CLUSTERING ORDER BY (timestamp DESC);
