output "alb_dns" {
  value = module.alb.alb_dns
}

output "ecs_cluster_name" {
  value = aws_ecs_cluster.cluster.name
}

output "raw_bucket_name" {
  value = module.s3_buckets.raw_bucket_name
}

output "golden_bucket_name" {
  value = module.s3_buckets.golden_bucket_name
}
