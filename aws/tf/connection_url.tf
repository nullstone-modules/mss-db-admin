resource "aws_secretsmanager_secret" "db_admin_mss" {
  name = "${var.name}/conn_url"
  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "db_admin_mss" {
  secret_id     = aws_secretsmanager_secret.db_admin_mss.id
  secret_string = "sqlserver://${urlencode(var.username)}:${urlencode(var.password)}@${var.host}:${var.port}/${urlencode(var.database)}"
}
