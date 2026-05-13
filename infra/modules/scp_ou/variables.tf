variable "scp_directory" {
  type        = string
  description = "Directory containing SCP JSON files"
}

variable "ou_id" {
  type        = string
  description = "Organizational Unit ID to attach SCPs to"
}
