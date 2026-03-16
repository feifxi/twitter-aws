# ── SQS Queues ──────────────────────────────────────

resource "aws_sqs_queue" "tweet_embedding_dlq" {
  name = "${var.project_name}-tweet-embedding-dlq"

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = ["arn:aws:sqs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${var.project_name}-tweet-embedding-queue"]
  })
}

resource "aws_sqs_queue" "tweet_embedding" {
  name                       = "${var.project_name}-tweet-embedding-queue"
  visibility_timeout_seconds = 180 # 6x Lambda timeout (30s) to prevent duplicate processing

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.tweet_embedding_dlq.arn
    maxReceiveCount     = 3
  })
}
