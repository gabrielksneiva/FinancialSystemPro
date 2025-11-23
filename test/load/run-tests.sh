#!/bin/bash

# Load Testing Runner - FinancialSystemPro
# Este script executa diferentes tipos de testes de carga usando k6

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variáveis
BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_DIR="$(dirname "$0")"
RESULTS_DIR="${TEST_DIR}/results"

# Criar diretório de resultados
mkdir -p "$RESULTS_DIR"

# Função para imprimir cabeçalho
print_header() {
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}================================${NC}"
}

# Função para verificar se k6 está instalado
check_k6() {
    if ! command -v k6 &> /dev/null; then
        echo -e "${RED}❌ k6 não está instalado!${NC}"
        echo -e "${YELLOW}Instale com: brew install k6 (macOS) ou https://k6.io/docs/getting-started/installation/${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ k6 encontrado: $(k6 version)${NC}"
}

# Função para verificar se o servidor está rodando
check_server() {
    echo -e "${YELLOW}Verificando servidor em $BASE_URL...${NC}"
    if curl -s -f "$BASE_URL/health" > /dev/null; then
        echo -e "${GREEN}✅ Servidor está rodando!${NC}"
    else
        echo -e "${RED}❌ Servidor não está acessível em $BASE_URL${NC}"
        echo -e "${YELLOW}Inicie o servidor antes de executar os testes.${NC}"
        exit 1
    fi
}

# Função para executar teste
run_test() {
    local test_name=$1
    local test_file=$2
    local output_file="${RESULTS_DIR}/${test_name}_$(date +%Y%m%d_%H%M%S).json"
    
    print_header "Executando: $test_name"
    
    k6 run \
        --out json="$output_file" \
        --summary-export="${output_file%.json}_summary.json" \
        -e BASE_URL="$BASE_URL" \
        "$test_file"
    
    echo -e "${GREEN}✅ Teste concluído! Resultados salvos em: $output_file${NC}\n"
}

# Menu principal
show_menu() {
    echo -e "${YELLOW}Qual teste você deseja executar?${NC}"
    echo "1) Load Test - Teste de carga gradual (5min)"
    echo "2) Stress Test - Teste de estresse até o limite (20min)"
    echo "3) Spike Test - Teste de picos repentinos (5min)"
    echo "4) Soak Test - Teste de resistência (4h)"
    echo "5) Smoke Test - Teste rápido de fumaça (30s)"
    echo "6) Executar TODOS os testes"
    echo "0) Sair"
    echo ""
    read -p "Escolha uma opção: " choice
}

# Teste smoke (rápido)
run_smoke_test() {
    print_header "Smoke Test"
    k6 run \
        --vus 5 \
        --duration 30s \
        -e BASE_URL="$BASE_URL" \
        "${TEST_DIR}/load-test.js"
}

# Main
main() {
    check_k6
    check_server
    
    if [ $# -eq 1 ]; then
        # Executar teste específico via argumento
        case $1 in
            load)
                run_test "load-test" "${TEST_DIR}/load-test.js"
                ;;
            stress)
                run_test "stress-test" "${TEST_DIR}/stress-test.js"
                ;;
            spike)
                run_test "spike-test" "${TEST_DIR}/spike-test.js"
                ;;
            soak)
                run_test "soak-test" "${TEST_DIR}/soak-test.js"
                ;;
            smoke)
                run_smoke_test
                ;;
            all)
                run_smoke_test
                run_test "load-test" "${TEST_DIR}/load-test.js"
                run_test "spike-test" "${TEST_DIR}/spike-test.js"
                echo -e "${YELLOW}⏭️  Pulando stress e soak (muito longos)${NC}"
                ;;
            *)
                echo -e "${RED}Teste desconhecido: $1${NC}"
                echo "Opções: load, stress, spike, soak, smoke, all"
                exit 1
                ;;
        esac
    else
        # Menu interativo
        while true; do
            show_menu
            case $choice in
                1)
                    run_test "load-test" "${TEST_DIR}/load-test.js"
                    ;;
                2)
                    run_test "stress-test" "${TEST_DIR}/stress-test.js"
                    ;;
                3)
                    run_test "spike-test" "${TEST_DIR}/spike-test.js"
                    ;;
                4)
                    echo -e "${YELLOW}⚠️  Este teste dura 4 horas. Deseja continuar? (y/n)${NC}"
                    read -p "" confirm
                    if [ "$confirm" = "y" ]; then
                        run_test "soak-test" "${TEST_DIR}/soak-test.js"
                    fi
                    ;;
                5)
                    run_smoke_test
                    ;;
                6)
                    run_smoke_test
                    run_test "load-test" "${TEST_DIR}/load-test.js"
                    run_test "spike-test" "${TEST_DIR}/spike-test.js"
                    echo -e "${YELLOW}⏭️  Pulando stress e soak (muito longos)${NC}"
                    ;;
                0)
                    echo -e "${GREEN}Saindo...${NC}"
                    exit 0
                    ;;
                *)
                    echo -e "${RED}Opção inválida!${NC}"
                    ;;
            esac
        done
    fi
}

# Executar
main "$@"
