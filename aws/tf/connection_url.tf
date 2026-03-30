resource "aws_secretsmanager_secret" "db_admin_mss" {
  name                    = "${var.name}/conn_url"
  tags                    = var.tags
  recovery_window_in_days = var.is_prod_env ? 7 : 0
}

resource "aws_secretsmanager_secret_version" "db_admin_mss" {
  secret_id     = aws_secretsmanager_secret.db_admin_mss.id
  secret_string = "sqlserver://${urlencode(var.username)}:${urlencode(var.password)}@${var.host}:${var.port}?database=${urlencode(var.database)}"
}
