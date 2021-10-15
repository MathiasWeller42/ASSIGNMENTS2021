package main

import (
	"bytes"
	"crypto/bcrypt"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"os"
	"strconv"
)

type RSA struct {
	d big.Int
	e *big.Int
	n big.Int
	p *big.Int
	q *big.Int
}

func MakeRSA(k int) *RSA {
	rsa := new(RSA)
	rsa.KeyGen(k)
	return rsa
}

func (rsa *RSA) KeyGen(k int) {
	//bit length of p*q = n is k
	rsa.e = big.NewInt(3) //given by the hand-in text
	rsa.p = rsa.GeneratePrime(k)
	rsa.q = rsa.GeneratePrime(k)
	rsa.n.Mul(rsa.p, rsa.q)
	rsa.d = rsa.GenerateD()
}

func (rsa *RSA) Encrypt(message *big.Int) big.Int {
	//c = m ^ e mod n
	var cipher big.Int
	cipher.Exp(message, rsa.e, &rsa.n)
	return cipher
}

func (rsa *RSA) Decrypt(cipher *big.Int) big.Int {
	//m = c ^ d mod n
	var message big.Int
	message.Exp(cipher, &rsa.d, &rsa.n)
	return message
}

func (rsa *RSA) DecryptWithKey(cipher *big.Int, keyN big.Int, keyD big.Int) big.Int {
	//m = c ^ d mod n
	var message big.Int
	message.Exp(cipher, &keyD, &keyN)
	return message
}

func (rsa *RSA) GenerateD() big.Int {
	var x, y, z big.Int
	x.Mul(y.Sub(rsa.p, big.NewInt(1)), z.Sub(rsa.q, big.NewInt(1)))
	x.ModInverse(rsa.e, &x)
	return x
}

func (rsa *RSA) GeneratePrime(k int) *big.Int {
	for {
		prime, err := rand.Prime(rand.Reader, k/2) //set the length to k/2
		if err != nil {
			fmt.Print(err) //print the error and try again
			continue
		}
		var x big.Int
		x.Mod(x.Sub(prime, big.NewInt(1)), rsa.e)
		mutuallyPrime := (x) //this is (prime-1) mod e

		//prime should be a probable prime
		//mutuallyPrime should not be 0 (this means that gcd(prime-1,e) is 1 when e=3)
		if prime.ProbablyPrime(1) && (mutuallyPrime.Cmp(big.NewInt(0)) != 0) {
			return prime
		}
	}
}

func (rsa *RSA) Hash(message string) *big.Int {
	messageAsByte := []byte(message)
	hashed := sha256.Sum256(messageAsByte)
	x := big.NewInt(0).SetBytes(hashed[:])
	return x
}

func (rsa *RSA) Sign(m *big.Int) big.Int {
	signature := rsa.Decrypt(m)
	return signature
}

func (rsa *RSA) SignWithKey(m *big.Int, keyN big.Int, keyD big.Int) big.Int {
	signature := rsa.DecryptWithKey(m, keyN, keyD)
	return signature
}

func (rsa *RSA) FullSign(message string, keyN big.Int, keyD big.Int) *big.Int {
	hashed := rsa.Hash(message)
	signature := rsa.SignWithKey(hashed, keyN, keyD)
	return &signature
}

func (rsa *RSA) FullSignTransaction(transaction *SignedTransaction, keyN string, keyD string) {
	n := rsa.ConvertStringToBigInt(keyN)
	d := rsa.ConvertStringToBigInt(keyD)
	stringToSign := transaction.ID + transaction.From + transaction.To + strconv.Itoa(transaction.Amount)
	transaction.Signature = rsa.ConvertBigIntToString(rsa.FullSign(stringToSign, *n, *d))
}

func (rsa *RSA) VerifyTransaction(transaction SignedTransaction) bool {
	stringToVerify := transaction.ID + transaction.From + transaction.To + strconv.Itoa(transaction.Amount)

	signature := rsa.ConvertStringToBigInt(transaction.Signature)

	keyN := *rsa.ConvertStringToBigInt(transaction.From)

	verified := rsa.VerifyWithKey(stringToVerify, *signature, keyN, big.NewInt(3))
	return verified
}

func (rsa *RSA) VerifyWithKey(m string, s big.Int, keyN big.Int, keyE *big.Int) bool {
	decryptedSignature := rsa.EncryptWithKey(&s, keyN, keyE)
	hashed := rsa.Hash(m)
	var verified bool
	if bytes.Equal(hashed.Bytes(), decryptedSignature.Bytes()) {
		verified = true
	} else {
		verified = false
	}
	return verified
}

func (rsa *RSA) EncryptWithKey(message *big.Int, keyN big.Int, keyE *big.Int) big.Int {
	//c = m ^ e mod n
	var cipher big.Int
	cipher.Exp(message, keyE, &keyN)
	return cipher
}

func (rsa *RSA) ConvertStringToBigInt(str string) *big.Int {
	s := new(big.Int)
	s.SetString(str, 10)
	return s
}

func (rsa *RSA) ConvertBigIntToString(big *big.Int) string {
	return big.String()
}

func (rsa *RSA) Verify(m string, s big.Int) bool {
	decryptedSignature := rsa.Encrypt(&s)
	hashed := rsa.Hash(m)
	var verified bool
	if bytes.Equal(hashed.Bytes(), decryptedSignature.Bytes()) {
		verified = true
	} else {
		verified = false
	}
	return verified
}

/**
func Generate(filename string, password string) string {
	f, _ := os.Create(filename)
	defer f.Close()
	rsa := MakeRSA(2000)
	secretKeyString := rsa.ConvertBigIntToString(&rsa.d)

	hashedPW := rsa.Hash(password).Bytes()

	aes := MakeAES()

	return rsa.ConvertBigIntToString(rsa.e) + ":" + rsa.ConvertBigIntToString(&rsa.n)
}

func Sign(filename string, password string, msg []byte) string {

	encryptedD, _ := os.ReadFile(filename)
	encryptedDAsString := string(encryptedD)

	return ""
}
*/
func main() {
	//Generate("hej", "pass123")
	//print(Sign("hej", "pass123", []byte("Mathias Weller er den sejeste <3<3<3")))
	filename := "hej"
	password := "p4SSw0rd1234!"
	f, _ := os.Create(filename)
	defer f.Close()
	rsa := MakeRSA(2048)
	secretKeyString := rsa.ConvertBigIntToString(&rsa.d) + ":" + rsa.ConvertBigIntToString(&rsa.n)

	hashedPW := rsa.Hash(password).Bytes()
	passwordHashToSave := bcrypt.GenerateFromPassword(hashedPW, 10)

	aes := MakeAES()
	encrypted := aes.Encrypt(secretKeyString, hashedPW)

	toWrite := bytes.append(encrypted, passwordHashToSave)
	f.Write(toWrite)

	read_bytes, _ := os.ReadFile(filename)

	decrypted := aes.Decrypt(string(read_bytes[:32]), hashedPW)

	if bytes.Equal(toWrite, read_bytes) {
		fmt.Println("The encrypted and written to file byte-arrays are equal")
	} else {
		fmt.Println("This is encrypted + password hash:", toWrite)
		fmt.Println("This is read from file:", read_bytes)
	}
	if decrypted == secretKeyString {
		fmt.Println("The string to be encrypted and the resulting decrypted string read from file are equal")
	} else {
		fmt.Println("This is SK string:", secretKeyString)
		fmt.Println("This is decrypted SK string:", decrypted)
	}

}
