############################################
# Enable Inspector in Delegated Admin Account
############################################

resource "aws_inspector2_account" "this" {
  # No arguments required — enabling Inspector in this account
}

############################################
# Enable Inspector Scanning Features
############################################

resource "aws_inspector2_enabler" "this" {
  account_ids = [var.account_id]

  resource_types = [
    "EC2",
    "ECR",
    "LAMBDA"
  ]
}
