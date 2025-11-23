package services

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/shared/utils"
	"fmt"
	"os"

	btcbase58 "github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

// TronWalletManager gerencia criação e validação de carteiras TRON
type TronWalletManager struct{}

// NewTronWalletManager cria uma nova instância do gerenciador TRON
func NewTronWalletManager() *TronWalletManager {
	return &TronWalletManager{}
}

// GenerateWallet cria uma nova carteira TRON com private key e address
func (twm *TronWalletManager) GenerateWallet() (*entities.WalletInfo, error) {
	// Gerar par de chaves ECDSA usando secp256k1 (mesma curva do Bitcoin/Ethereum/TRON)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar chave privada: %v", err)
	}

	// Obter bytes da private key
	privKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privKeyBytes)

	// Criptografar a private key para armazenamento seguro
	encryptedPrivKey, err := utils.EncryptPrivateKey(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar private key: %v", err)
	}

	// Obter public key (65 bytes: 04 + 32 bytes X + 32 bytes Y)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("erro ao converter public key")
	}

	// Serializar public key (sem o prefixo 0x04)
	pubKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	// Gerar endereço TRON a partir da public key
	address := twm.generateTronAddress(pubKeyBytes)

	// DEBUG: Verificar endereço ANTES de retornar
	fmt.Printf("🔍 GenerateWallet FINAL: address='%s', len=%d\n", address, len(address))

	return &entities.WalletInfo{
		Address:          address,
		PublicKey:        hex.EncodeToString(pubKeyBytes),
		PrivateKey:       privateKeyHex,
		EncryptedPrivKey: encryptedPrivKey,
		Blockchain:       entities.BlockchainTRON,
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
func (twm *TronWalletManager) GetBlockchainType() entities.BlockchainType {
	return entities.BlockchainTRON
}

// generateTronAddress gera endereço TRON a partir da public key
func (twm *TronWalletManager) generateTronAddress(pubKeyBytes []byte) string {
	// TRON usa Keccak256 da public key (sem o prefixo 0x04)
	// Remover o primeiro byte (0x04) se presente
	originalLen := len(pubKeyBytes)
	if len(pubKeyBytes) == 65 && pubKeyBytes[0] == 0x04 {
		pubKeyBytes = pubKeyBytes[1:]
	}

	// Aplicar Keccak256 (SHA3-256)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(pubKeyBytes)
	addressHash := hash.Sum(nil)

	// Pegar os últimos 20 bytes
	address20 := addressHash[len(addressHash)-20:]

	// Adicionar versão TRON (0x41 para mainnet/testnet)
	versionedAddress := append([]byte{0x41}, address20...)

	// Calcular checksum (double SHA-256)
	checksum := sha256.Sum256(versionedAddress)
	checksum = sha256.Sum256(checksum[:])

	// Adicionar os primeiros 4 bytes do checksum
	addressBytes := append(versionedAddress, checksum[:4]...)

	// Codificar em Base58 usando biblioteca confiável
	encodedAddress := btcbase58.Encode(addressBytes)

	// DEBUG CRÍTICO: Print para STDOUT (deve aparecer sempre)
	fmt.Fprintf(os.Stdout, "[GENERATE_ADDRESS_DEBUG] originalPubKeyLen=%d, finalPubKeyLen=%d, addressHashLen=%d, address20Len=%d, versionedLen=%d, addressBytesLen=%d, ENCODED='%s', encodedLen=%d\n",
		originalLen, len(pubKeyBytes), len(addressHash), len(address20), len(versionedAddress), len(addressBytes), encodedAddress, len(encodedAddress))

	return encodedAddress
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
