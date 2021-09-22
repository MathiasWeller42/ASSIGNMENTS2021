package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
)

func EncryptToFile(filename string, key []byte) { //key type?
	//read plaintext from the file
	toBeEncrypted, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Something went wrong when reading the plaintext file", filename)
		fmt.Println("Error", err)
		return
	}

	//do some encrypting stuff
	block, _ := aes.NewCipher(key)
	iv := make([]byte, block.BlockSize())
	rand.Read(iv)
	stream := cipher.NewCTR(block, iv)
	cipherArray := make([]byte, len(toBeEncrypted))
	ciphertext := append(iv, cipherArray...)
	stream.XORKeyStream(ciphertext[block.BlockSize():], toBeEncrypted)

	//overwrite the original file (truncate it)
	file, _ := os.Create(filename)
	defer file.Close()

	//write ciphertext to the file
	file.WriteString(string(ciphertext))
}

func DecryptFromFile(filename string, key []byte) {
	//read ciphertext from the file
	toBeDecrypted, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Something went wrong when reading the ciphertext file", filename)
		fmt.Println("Error", err)
		return
	}

	//do some decrypting stuff
	block, _ := aes.NewCipher(key)
	plaintext := make([]byte, len(toBeDecrypted)-block.BlockSize())
	iv := toBeDecrypted[:block.BlockSize()]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, toBeDecrypted[block.BlockSize():])

	//overwrite the original file (truncate it)
	file, _ := os.Create(filename)
	defer file.Close()

	//write plaintext to the file
	file.WriteString(string(plaintext))
}

func generateAESKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	return key
}

func main() {
	example := "Hej med dig, Simon :D"
	runEncryption("example.txt", []byte(example), true)

	//encryption of RSA key (we only encrypt d, since e and n are publicly known)
	d, n, e := GenerateRSAKeys()

	fmt.Println("The key before encrypt/decrypt:", big.NewInt(1).SetBytes(d))

	//enrypting and decrypting d
	d = runEncryption("secretkey.txt", d, false)

	dAsBigInt := big.NewInt(1)
	dAsBigInt.SetBytes(d)

	fmt.Println("The key after encrypt/decrypt:", dAsBigInt)

	//an m to test the RSA key on
	m := big.NewInt(12345)
	c := Encrypt(m, e, n)

	//decrypting m with the encrypted/decrypted key d
	p := Decrypt(&c, *dAsBigInt, n)

	fmt.Println("These should be the same:", m, p)

}

func runEncryption(filename string, toBeEncrypted []byte, print bool) []byte {
	key := generateAESKey()

	f, _ := os.Create(filename)
	defer f.Close()
	f.Write(toBeEncrypted)
	m, _ := os.ReadFile(filename)
	if print {
		fmt.Println("Plaintext before:", string(m))
	}

	EncryptToFile(filename, key)
	c, _ := os.ReadFile(filename)
	if print {
		fmt.Println("After encryption the file contains:", string(c))
	}

	DecryptFromFile(filename, key)
	d, _ := os.ReadFile(filename)
	if print {
		fmt.Println("After decryption the file contains:", string(d))
	}
	return d
}
