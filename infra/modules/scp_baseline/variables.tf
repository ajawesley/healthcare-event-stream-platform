variable "scp_directory" {
  description = "Directory containing SCP JSON files"
  type        = string
}

variable "root_id" {
  description = "AWS Organizations Root ID (e.g., r-abcd)"
  type        = string
}
