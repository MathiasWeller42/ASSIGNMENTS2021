package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

type AES struct {
}

func MakeAES() *AES {
	this_aes := new(AES)
	return this_aes
}

func (this_aes *AES) Encrypt(plaintext string, key []byte) []byte { //key type?
	//read plaintext from the file
	toBeEncrypted := []byte(plaintext)
	//do some encrypting stuff
	block, _ := aes.NewCipher(key)
	iv := make([]byte, block.BlockSize())
	rand.Read(iv)
	stream := cipher.NewCTR(block, iv)
	cipherArray := make([]byte, len(toBeEncrypted))
	ciphertext := append(iv, cipherArray...)
	stream.XORKeyStream(ciphertext[block.BlockSize():], toBeEncrypted)

	return ciphertext
}

func (this_aes *AES) Decrypt(encrypted string, key []byte) string {
	//do some decrypting stuff
	toBeDecrypted := []byte(encrypted)

	block, _ := aes.NewCipher(key)
	plaintext := make([]byte, len(toBeDecrypted)-block.BlockSize())
	iv := toBeDecrypted[:block.BlockSize()]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, toBeDecrypted[block.BlockSize():])
	plaintextString := string(plaintext)
	//write plaintext to the file
	return plaintextString
}

func (this_aes *AES) generateAESKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	return key
}

/*
func main() {
	fmt.Println("First we test RSA encryption on its own")
	RunRSA()

	fmt.Println("Now let's go to AES")
	example := "Hej med dig, Simon :D"
	runEncryption("example.txt", []byte(example), true)

	fmt.Println("And finally let's encrypt/decrypt an RSA key (d)")
	//encryption of RSA key (we only encrypt d, since e and n are publicly known)
	d, n, e := GenerateRSAKeys()

	fmt.Println("The key before encrypt/decrypt:", big.NewInt(1).SetBytes(d))

	//enrypting and decrypting d
	d = runEncryption("secretkey.txt", d, false)

	dAsBigInt := big.NewInt(1)
	dAsBigInt.SetBytes(d)

	fmt.Println("The key after encrypt/decrypt:", dAsBigInt)

	//an m to test the RSA key on
	test_m := 798493272
	fmt.Println("We use the decrypted RSA key to encrypt the number", test_m)
	m := big.NewInt(int64(test_m))
	c := Encrypt(m, e, n)

	//decrypting m with the encrypted/decrypted key d
	p := Decrypt(&c, *dAsBigInt, n)

	fmt.Println("These should be the same (except for some formatting):", m, p)

}
*/
