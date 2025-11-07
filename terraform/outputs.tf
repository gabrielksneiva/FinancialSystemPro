output "public_ip" {
  description = "IP público da instância EC2"
  value       = aws_instance.financialsystempro.public_ip
}

output "app_url" {
  description = "URL pública da aplicação"
  value       = "http://${aws_instance.financialsystempro.public_ip}:3000"
}
