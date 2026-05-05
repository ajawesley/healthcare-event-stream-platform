output "bucket_name" {
  value = aws_s3_bucket.this.bucket
}

output "bucket_arn" {
  value = aws_s3_bucket.this.arn
}

output "kms_key_arn" {
  value = aws_kms_key.this.arn
}

output "script_prefix" {
  value = "s3://${aws_s3_bucket.this.bucket}/scripts/"
}

output "temp_prefix" {
  value = "s3://${aws_s3_bucket.this.bucket}/tmp/"
}

