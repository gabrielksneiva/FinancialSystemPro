package services

import (
	"os"
	"testing"
)

func TestAESEncryptionProviderRoundTrip(t *testing.T) {
	// 32 bytes master key base64
	master := make([]byte, 32)
	for i := range master {
		master[i] = byte(i + 1)
	}
	os.Setenv("ENCRYPTION_MASTER_KEY", encodeBase64(master))
	prov, err := NewAESEncryptionProviderFromEnv()
	if err != nil {
		t.Fatalf("erro criando provider: %v", err)
	}
	aesProv, ok := prov.(*AESEncryptionProvider)
	if !ok {
		t.Fatalf("esperava AESEncryptionProvider, obtido %T", prov)
	}
	plain := "segredo-teste-123"
	enc, err := aesProv.Encrypt(plain)
	if err != nil {
		t.Fatalf("erro encrypt: %v", err)
	}
	if enc == plain {
		t.Fatalf("ciphertext igual ao plaintext")
	}
	dec, err := aesProv.Decrypt(enc)
	if err != nil {
		t.Fatalf("erro decrypt: %v", err)
	}
	if dec != plain {
		t.Fatalf("mismatch esperado %s got %s", plain, dec)
	}
}

// encodeBase64 evita importar encoding/base64 no teste para manter escopo simples.
// (Na prática poderíamos usar diretamente base64.StdEncoding)
func encodeBase64(b []byte) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	out := make([]byte, 0, ((len(b)+2)/3)*4)
	for i := 0; i < len(b); i += 3 {
		var c0, c1, c2 byte
		c0 = b[i]
		if i+1 < len(b) {
			c1 = b[i+1]
		}
		if i+2 < len(b) {
			c2 = b[i+2]
		}
		out = append(out, table[c0>>2])
		out = append(out, table[((c0&0x03)<<4)|(c1>>4)])
		if i+1 < len(b) {
			out = append(out, table[((c1&0x0F)<<2)|(c2>>6)])
		} else {
			out = append(out, '=')
		}
		if i+2 < len(b) {
			out = append(out, table[c2&0x3F])
		} else {
			out = append(out, '=')
		}
	}
	return string(out)
}
