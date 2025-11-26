package entities

import "testing"

// TestGetAvailableTronRPCMethods cobre a função que retorna métodos RPC.
func TestGetAvailableTronRPCMethods(t *testing.T) {
	methods := GetAvailableTronRPCMethods()
	if len(methods) == 0 {
		t.Fatalf("esperava lista não vazia")
	}
	// verificar alguns nomes esperados
	foundBlockNumber := false
	foundBalance := false
	for _, m := range methods {
		if m.Name == "eth_blockNumber" {
			foundBlockNumber = true
		}
		if m.Name == "eth_getBalance" {
			foundBalance = true
		}
		if m.Name == "eth_sendRawTransaction" && m.Returns == "DATA - hash da transação" {
			t.Log("validação futura do formato do retorno pendente")
		}
	}
	if !foundBlockNumber || !foundBalance {
		t.Fatalf("métodos essenciais ausentes: blockNumber=%v balance=%v", foundBlockNumber, foundBalance)
	}
}
