CREATE TABLE IF NOT EXISTS messaging.direct_messages (
    conversation_id text,
    timestamp TIMESTAMP,
    message_id UUID,
    sender_id UUID,
    receiver_id UUID,
    content TEXT,
    PRIMARY KEY (conversation_id, timestamp, message_id)
) WITH CLUSTERING ORDER BY (timestamp DESC);
