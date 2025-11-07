#!/bin/bash
# Atualiza pacotes e instala dependências
sudo apt update -y
sudo apt install -y docker.io docker-compose git

# Inicia Docker
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker ubuntu

# Clona o repositório
cd /home/ubuntu
git clone https://github.com/gabrielksneiva/FinancialSystemPro.git
cd FinancialSystemPro

# Cria o arquivo .env
cat <<EOF > .env
DB_HOST=postgres
DB_ADMIN=admin
DB_PASSWORD=g123
DB_NAME=database
DB_PORT=5432
EOF

# Sobe o container
sudo docker-compose up -d
