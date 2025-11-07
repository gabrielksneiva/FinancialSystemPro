resource "aws_security_group" "financialsystempro_sg" {
  name_prefix = "financialsystempro-"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "financialsystempro" {
  ami                    = "ami-0e86e20dae9224db8" # Ubuntu 22.04 Free Tier (us-east-1)
  instance_type          = var.instance_type
  key_name               = var.key_name
  vpc_security_group_ids = [aws_security_group.financialsystempro_sg.id]

  user_data = file("${path.module}/user_data.sh")

  tags = {
    Name = "FinancialSystemPro"
  }
}
