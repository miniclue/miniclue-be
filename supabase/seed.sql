-- Create Queues

SELECT pgmq.create('ingestion_queue');
SELECT pgmq.create('embedding_queue');
SELECT pgmq.create('explanation_queue');
SELECT pgmq.create('summary_queue');
