output "raw_bucket_name" {
  value = aws_s3_bucket.raw.bucket
}

output "raw_bucket_arn" {
  value = aws_s3_bucket.raw.arn
}

output "golden_bucket_name" {
  value = aws_s3_bucket.golden.bucket
}

output "golden_bucket_arn" {
  value = aws_s3_bucket.golden.arn
}

output "script_bucket_name" {
  value = aws_s3_bucket.scripts.bucket
}

output "script_bucket_arn" {
  value = aws_s3_bucket.scripts.arn
}

output "access_logs_bucket_name" {
  value = aws_s3_bucket.access_logs.bucket
}

output "access_logs_bucket_arn" {
  value = aws_s3_bucket.access_logs.arn
}
