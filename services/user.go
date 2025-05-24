package services

import (
	"fmt"
	"strings"
)

type ServiceX struct {
	Teste string
}

func (s *ServiceX) Teste1() string {
	return fmt.Sprint(s.Teste)
}

func (s *ServiceX) QqrMerda() (string, error) {
	frase_completa := fmt.Sprintf("qualquer coisa")

	if !strings.EqualFold(frase_completa, "errado") {
		return "", fmt.Errorf("erro")
	}

	return frase_completa, nil
}
