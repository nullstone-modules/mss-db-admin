resource "aws_lambda_function" "db_admin" {
  function_name    = var.name
  tags             = var.tags
  role             = aws_iam_role.db_admin.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename          = "${path.module}/files/mss-db-admin.zip"
  source_code_hash = filebase64sha256("${path.module}/files/mss-db-admin.zip")
  // This can take ~5s to create a db sometimes
  timeout          = 10

  environment {
    variables = {
      DB_CONN_URL_SECRET_ID = aws_secretsmanager_secret.db_admin_mss.id
    }
  }

  vpc_config {
    security_group_ids = concat([aws_security_group.db_admin.id], var.network.security_group_ids)
    subnet_ids         = var.network.subnet_ids
  }
}
