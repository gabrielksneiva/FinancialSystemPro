package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"financial-system-pro/domain"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

// TronWalletManager gerencia criação e validação de carteiras TRON
type TronWalletManager struct{}

// NewTronWalletManager cria uma nova instância do gerenciador TRON
func NewTronWalletManager() *TronWalletManager {
	return &TronWalletManager{}
}

// GenerateWallet cria uma nova carteira TRON com private key e address
func (twm *TronWalletManager) GenerateWallet() (*domain.WalletInfo, error) {
	// Gerar private key aleatória (32 bytes)
	privKeyBytes := make([]byte, 32)
	_, err := rand.Read(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar private key: %v", err)
	}

	// Gerar public key a partir da private key (usar hash SHA-256)
	publicKeyHash := sha256.Sum256(privKeyBytes)
	pubKeyBytes := publicKeyHash[:]

	// Gerar endereço TRON a partir da public key
	address := twm.generateTronAddress(pubKeyBytes)

	return &domain.WalletInfo{
		Address:    address,
		PublicKey:  hex.EncodeToString(pubKeyBytes),
		Blockchain: domain.BlockchainTRON,
	}, nil
}

// ValidateAddress valida se um endereço é um endereço TRON válido
func (twm *TronWalletManager) ValidateAddress(address string) bool {
	// Endereços TRON começam com 'T' e têm 34 caracteres
	if len(address) != 34 {
		return false
	}
	if address[0] != 'T' {
		return false
	}

	// Verificar se é base58 válido
	decoded, err := twm.base58Decode(address)
	if err != nil {
		return false
	}

	// Endereço TRON tem 25 bytes (21 bytes do hash + 4 bytes de checksum)
	if len(decoded) != 25 {
		return false
	}

	// Validar checksum
	hash := sha256.Sum256(decoded[:21])
	hash = sha256.Sum256(hash[:])

	return string(hash[:4]) == string(decoded[21:])
}

// GetBlockchainType retorna o tipo de blockchain
func (twm *TronWalletManager) GetBlockchainType() domain.BlockchainType {
	return domain.BlockchainTRON
}

// generateTronAddress gera endereço TRON a partir da public key
func (twm *TronWalletManager) generateTronAddress(pubKeyBytes []byte) string {
	// TRON usa SHA-256 então RIPEMD-160 da public key
	sha256Hash := sha256.Sum256(pubKeyBytes)

	ripemd := ripemd160.New()
	ripemd.Write(sha256Hash[:])
	addressHash := ripemd.Sum(nil)

	// Adicionar versão TRON (0x41)
	versionedHash := append([]byte{0x41}, addressHash...)

	// Adicionar checksum (primeiros 4 bytes do double SHA-256)
	checksum := sha256.Sum256(versionedHash)
	checksum = sha256.Sum256(checksum[:])

	addressBytes := append(versionedHash, checksum[:4]...)

	// Codificar em base58
	return twm.base58Encode(addressBytes)
}

// base58 alphabet (Bitcoin style, sem 0, O, I, l)
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// base58Encode codifica bytes em base58
func (twm *TronWalletManager) base58Encode(data []byte) string {
	result := ""

	// Contar zeros à esquerda
	zeroes := 0
	for _, v := range data {
		if v != 0 {
			break
		}
		zeroes++
	}

	// Adicionar '1' para cada zero à esquerda
	for i := 0; i < zeroes; i++ {
		result += "1"
	}

	// Se todos os bytes são zero
	if zeroes == len(data) {
		return result
	}

	// Converter restante dos dados
	num := twm.bytesToBigEndianInt(data[zeroes:])
	for num > 0 {
		digit := num % 58
		result = string(base58Alphabet[digit]) + result
		num = num / 58
	}

	return result
}

// base58Decode decodifica string base58 para bytes
func (twm *TronWalletManager) base58Decode(s string) ([]byte, error) {
	result := []byte{}

	// Contar '1's à esquerda
	zeroes := 0
	for _, c := range s {
		if c != '1' {
			break
		}
		zeroes++
	}

	// Adicionar zeros
	for i := 0; i < zeroes; i++ {
		result = append(result, 0)
	}

	if zeroes == len(s) {
		return result, nil
	}

	// Decodificar o resto
	num := 0
	for _, c := range s[zeroes:] {
		num = num * 58
		digit := -1
		for i, char := range base58Alphabet {
			if char == c {
				digit = i
				break
			}
		}
		if digit == -1 {
			return nil, fmt.Errorf("caractere inválido em base58: %c", c)
		}
		num = num + digit
	}

	// Converter número para bytes big-endian
	for num > 0 {
		result = append(result, byte(num%256))
		num = num / 256
	}

	// Reverter para ordem correta
	for i, j := zeroes, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// bytesToBigEndianInt converte bytes para número inteiro
func (twm *TronWalletManager) bytesToBigEndianInt(data []byte) int {
	result := 0
	for _, b := range data {
		result = result*256 + int(b)
	}
	return result
}
