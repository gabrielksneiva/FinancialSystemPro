variable "key_name" {
  description = "Nome da key pair existente na AWS"
  type        = string
  default     = "FinancialSystemPro-EC2"
}

variable "instance_type" {
  description = "Tipo da instância EC2"
  type        = string
  default     = "t3.micro"
}

variable "repo_url" {
  description = "URL do repositório GitHub"
  type        = string
  default     = "https://github.com/gabrielksneiva/FinancialSystemPro.git"
}
