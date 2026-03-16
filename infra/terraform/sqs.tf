# ── SQS Queues ──────────────────────────────────────

resource "aws_sqs_queue" "tweet_embedding_dlq" {
  name = "${var.project_name}-tweet-embedding-dlq"
}

resource "aws_sqs_queue" "tweet_embedding" {
  name                       = "${var.project_name}-tweet-embedding-queue"
  visibility_timeout_seconds = 180 # 6x Lambda timeout (30s) to prevent duplicate processing

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.tweet_embedding_dlq.arn
    maxReceiveCount     = 3
  })
}
