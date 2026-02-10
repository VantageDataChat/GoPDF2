package gopdf

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
)

// EncryptionMethod represents the PDF encryption algorithm.
type EncryptionMethod int

const (
	// EncryptRC4V1 is 40-bit RC4 encryption (PDF 1.1+).
	EncryptRC4V1 EncryptionMethod = iota
	// EncryptRC4V2 is up to 128-bit RC4 encryption (PDF 1.4+).
	EncryptRC4V2
	// EncryptAES128 is 128-bit AES encryption (PDF 1.6+).
	EncryptAES128
	// EncryptAES256 is 256-bit AES encryption (PDF 2.0+).
	EncryptAES256
)

// AESEncryptionConfig configures AES-based PDF encryption.
type AESEncryptionConfig struct {
	// Method selects the encryption algorithm.
	Method EncryptionMethod
	// UserPassword is the password required to open the document.
	UserPassword string
	// OwnerPassword is the password for full access. If empty, a random one is generated.
	OwnerPassword string
	// Permissions is a bitmask of allowed operations (PermissionsPrint, etc.).
	Permissions int
}

// aesEncryptionObj represents an AES encryption dictionary object.
type aesEncryptionObj struct {
	method  EncryptionMethod
	uValue  []byte
	oValue  []byte
	ueValue []byte // UE for AES-256
	oeValue []byte // OE for AES-256
	pValue  int
	keyLen  int
}

func (e *aesEncryptionObj) init(func() *GoPdf) {}

func (e *aesEncryptionObj) getType() string {
	return "Encryption"
}

func (e *aesEncryptionObj) write(w io.Writer, objID int) error {
	io.WriteString(w, "<<\n")
	io.WriteString(w, "/Filter /Standard\n")

	switch e.method {
	case EncryptAES128:
		io.WriteString(w, "/V 4\n")
		io.WriteString(w, "/R 4\n")
		fmt.Fprintf(w, "/Length %d\n", e.keyLen*8)
		io.WriteString(w, "/CF <</StdCF <</AuthEvent /DocOpen /CFM /AESV2 /Length 16>>>>\n")
		io.WriteString(w, "/StmF /StdCF\n")
		io.WriteString(w, "/StrF /StdCF\n")
	case EncryptAES256:
		io.WriteString(w, "/V 5\n")
		io.WriteString(w, "/R 6\n")
		fmt.Fprintf(w, "/Length %d\n", e.keyLen*8)
		io.WriteString(w, "/CF <</StdCF <</AuthEvent /DocOpen /CFM /AESV3 /Length 32>>>>\n")
		io.WriteString(w, "/StmF /StdCF\n")
		io.WriteString(w, "/StrF /StdCF\n")
	default:
		// RC4 fallback
		io.WriteString(w, "/V 1\n")
		io.WriteString(w, "/R 2\n")
	}

	fmt.Fprintf(w, "/O <%X>\n", e.oValue)
	fmt.Fprintf(w, "/U <%X>\n", e.uValue)
	fmt.Fprintf(w, "/P %d\n", e.pValue)

	if e.method == EncryptAES256 {
		if len(e.ueValue) > 0 {
			fmt.Fprintf(w, "/UE <%X>\n", e.ueValue)
		}
		if len(e.oeValue) > 0 {
			fmt.Fprintf(w, "/OE <%X>\n", e.oeValue)
		}
		io.WriteString(w, "/Perms <")
		perms := computePermsValue(e.pValue, e.uValue[:16])
		fmt.Fprintf(w, "%X", perms)
		io.WriteString(w, ">\n")
	}

	io.WriteString(w, ">>\n")
	return nil
}

// SetEncryption configures AES encryption for the document.
// Supports AES-128 and AES-256 in addition to the existing RC4.
//
// Example:
//
//	pdf.SetEncryption(gopdf.AESEncryptionConfig{
//	    Method:        gopdf.EncryptAES128,
//	    UserPassword:  "user123",
//	    OwnerPassword: "owner456",
//	    Permissions:   gopdf.PermissionsPrint | gopdf.PermissionsCopy,
//	})
func (gp *GoPdf) SetEncryption(config AESEncryptionConfig) error {
	switch config.Method {
	case EncryptRC4V1, EncryptRC4V2:
		// Delegate to existing RC4 protection.
		perms := config.Permissions
		if perms == 0 {
			perms = PermissionsPrint
		}
		p := gp.createProtection()
		return p.SetProtection(perms, []byte(config.UserPassword), []byte(config.OwnerPassword))

	case EncryptAES128:
		return gp.setupAES128(config)

	case EncryptAES256:
		return gp.setupAES256(config)

	default:
		return fmt.Errorf("unsupported encryption method: %d", config.Method)
	}
}

func (gp *GoPdf) setupAES128(config AESEncryptionConfig) error {
	userPass := []byte(config.UserPassword)
	ownerPass := []byte(config.OwnerPassword)
	if len(ownerPass) == 0 {
		ownerPass = make([]byte, 16)
		rand.Read(ownerPass)
	}

	perms := 192 | config.Permissions
	pValue := -((perms ^ 255) + 1)

	// Compute O value (owner hash).
	paddedOwner := padPassword(ownerPass)
	ownerHash := md5.Sum(paddedOwner)
	ownerKey := ownerHash[:]
	// MD5 50 iterations for R>=3.
	for i := 0; i < 50; i++ {
		h := md5.Sum(ownerKey[:16])
		ownerKey = h[:]
	}
	ownerKey = ownerKey[:16]

	paddedUser := padPassword(userPass)
	oValue := make([]byte, 32)
	copy(oValue, paddedUser)
	// RC4 encrypt with 20 iterations.
	for i := 0; i <= 19; i++ {
		tmpKey := make([]byte, 16)
		for j := range ownerKey {
			tmpKey[j] = ownerKey[j] ^ byte(i)
		}
		cip, err := newRC4Cipher(tmpKey)
		if err != nil {
			return err
		}
		cip.XORKeyStream(oValue, oValue)
	}

	// Compute encryption key.
	m := md5.New()
	m.Write(paddedUser)
	m.Write(oValue)
	pBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(pBytes, uint32(int32(pValue)))
	m.Write(pBytes)
	encKey := m.Sum(nil)
	for i := 0; i < 50; i++ {
		h := md5.Sum(encKey[:16])
		encKey = h[:]
	}
	encKey = encKey[:16]

	// Compute U value.
	um := md5.New()
	um.Write(protectionPadding)
	uHash := um.Sum(nil)
	uValue := make([]byte, 32)
	copy(uValue, uHash[:16])
	for i := 0; i <= 19; i++ {
		tmpKey := make([]byte, 16)
		for j := range encKey {
			tmpKey[j] = encKey[j] ^ byte(i)
		}
		cip, err := newRC4Cipher(tmpKey)
		if err != nil {
			return err
		}
		cip.XORKeyStream(uValue[:16], uValue[:16])
	}

	encObj := &aesEncryptionObj{
		method: EncryptAES128,
		uValue: uValue,
		oValue: oValue,
		pValue: pValue,
		keyLen: 16,
	}

	gp.encryptionObjID = gp.addObj(encObj) + 1
	return nil
}

func (gp *GoPdf) setupAES256(config AESEncryptionConfig) error {
	userPass := []byte(config.UserPassword)
	ownerPass := []byte(config.OwnerPassword)
	if len(ownerPass) == 0 {
		ownerPass = make([]byte, 16)
		rand.Read(ownerPass)
	}

	perms := 192 | config.Permissions
	pValue := -((perms ^ 255) + 1)

	// Generate random file encryption key (32 bytes).
	fileKey := make([]byte, 32)
	if _, err := rand.Read(fileKey); err != nil {
		return fmt.Errorf("generate file key: %w", err)
	}

	// User validation salt and key salt (8 bytes each).
	userValSalt := make([]byte, 8)
	userKeySalt := make([]byte, 8)
	rand.Read(userValSalt)
	rand.Read(userKeySalt)

	// U value = SHA-256(password + validation salt) + validation salt + key salt.
	uHash := sha256.Sum256(append(userPass, userValSalt...))
	uValue := make([]byte, 48)
	copy(uValue[:32], uHash[:])
	copy(uValue[32:40], userValSalt)
	copy(uValue[40:48], userKeySalt)

	// UE value = AES-256-CBC encrypt file key with SHA-256(password + key salt).
	ueKeyHash := sha256.Sum256(append(userPass, userKeySalt...))
	ueValue, err := aesEncryptCBC(ueKeyHash[:], fileKey)
	if err != nil {
		return fmt.Errorf("encrypt UE: %w", err)
	}

	// Owner validation salt and key salt.
	ownerValSalt := make([]byte, 8)
	ownerKeySalt := make([]byte, 8)
	rand.Read(ownerValSalt)
	rand.Read(ownerKeySalt)

	// O value = SHA-256(password + validation salt + U) + validation salt + key salt.
	oInput := append(ownerPass, ownerValSalt...)
	oInput = append(oInput, uValue[:48]...)
	oHash := sha256.Sum256(oInput)
	oValue := make([]byte, 48)
	copy(oValue[:32], oHash[:])
	copy(oValue[32:40], ownerValSalt)
	copy(oValue[40:48], ownerKeySalt)

	// OE value = AES-256-CBC encrypt file key with SHA-256(password + key salt + U).
	oeInput := append(ownerPass, ownerKeySalt...)
	oeInput = append(oeInput, uValue[:48]...)
	oeKeyHash := sha256.Sum256(oeInput)
	oeValue, err := aesEncryptCBC(oeKeyHash[:], fileKey)
	if err != nil {
		return fmt.Errorf("encrypt OE: %w", err)
	}

	encObj := &aesEncryptionObj{
		method:  EncryptAES256,
		uValue:  uValue,
		oValue:  oValue,
		ueValue: ueValue,
		oeValue: oeValue,
		pValue:  pValue,
		keyLen:  32,
	}

	gp.encryptionObjID = gp.addObj(encObj) + 1
	return nil
}

// aesEncryptCBC encrypts data using AES-CBC with PKCS7 padding.
func aesEncryptCBC(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// PKCS7 padding.
	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padded := make([]byte, len(plaintext)+padding)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	// Generate random IV.
	iv := make([]byte, blockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	ciphertext := make([]byte, blockSize+len(padded))
	copy(ciphertext[:blockSize], iv)

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[blockSize:], padded)

	return ciphertext, nil
}

// aesDecryptCBC decrypts AES-CBC encrypted data (IV prepended).
func aesDecryptCBC(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(ciphertext) < blockSize*2 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:blockSize]
	data := ciphertext[blockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)

	// Remove PKCS7 padding.
	if len(plaintext) > 0 {
		padLen := int(plaintext[len(plaintext)-1])
		if padLen > 0 && padLen <= blockSize && padLen <= len(plaintext) {
			plaintext = plaintext[:len(plaintext)-padLen]
		}
	}

	return plaintext, nil
}

// computePermsValue computes the /Perms value for AES-256 (R=6).
func computePermsValue(pValue int, encKey []byte) []byte {
	perms := make([]byte, 16)
	binary.LittleEndian.PutUint32(perms[0:4], uint32(int32(pValue)))
	perms[4] = 0xFF
	perms[5] = 0xFF
	perms[6] = 0xFF
	perms[7] = 0xFF
	perms[8] = 'T' // EncryptMetadata = true
	perms[9] = 'a'
	perms[10] = 'd'
	perms[11] = 'b'
	// 12-15: random
	rand.Read(perms[12:16])

	// AES-ECB encrypt with file encryption key.
	if len(encKey) >= 16 {
		block, err := aes.NewCipher(encKey[:16])
		if err == nil {
			dst := make([]byte, 16)
			block.Encrypt(dst, perms)
			return dst
		}
	}
	return perms
}

// newRC4Cipher is a helper to create an RC4 cipher (wraps crypto/rc4).
func newRC4Cipher(key []byte) (*rc4Cipher, error) {
	// Use a simple RC4 implementation to avoid import cycle.
	if len(key) < 1 || len(key) > 256 {
		return nil, fmt.Errorf("invalid RC4 key length: %d", len(key))
	}
	var s [256]byte
	for i := range s {
		s[i] = byte(i)
	}
	j := 0
	for i := 0; i < 256; i++ {
		j = (j + int(s[i]) + int(key[i%len(key)])) & 0xFF
		s[i], s[j] = s[j], s[i]
	}
	return &rc4Cipher{s: s}, nil
}

type rc4Cipher struct {
	s    [256]byte
	i, j uint8
}

func (c *rc4Cipher) XORKeyStream(dst, src []byte) {
	for k := range src {
		c.i++
		c.j += c.s[c.i]
		c.s[c.i], c.s[c.j] = c.s[c.j], c.s[c.i]
		dst[k] = src[k] ^ c.s[c.s[c.i]+c.s[c.j]]
	}
}
