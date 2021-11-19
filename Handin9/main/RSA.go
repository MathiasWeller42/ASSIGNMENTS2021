package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
	"strings"
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

func MakeRSAWithKeys(n string, d string) *RSA {
	rsa := new(RSA)
	rsa.n = *ConvertStringToBigInt(n)
	rsa.d = *ConvertStringToBigInt(d)
	rsa.e = big.NewInt(3)
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

func Hash(message string) *big.Int {
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

func (rsa *RSA) FullSignStringKeys(message string, keyN string, keyD string) string {
	n := ConvertStringToBigInt(keyN)
	d := ConvertStringToBigInt(keyD)
	return ConvertBigIntToString(rsa.FullSign(message, *n, *d))
}

func (rsa *RSA) FullSign(message string, keyN big.Int, keyD big.Int) *big.Int {
	hashed := Hash(message)
	signature := rsa.SignWithKey(hashed, keyN, keyD)
	return &signature
}

func (rsa *RSA) FullSignTransaction(transaction *SignedTransaction, keyN string, keyD string) {
	n := ConvertStringToBigInt(keyN)
	d := ConvertStringToBigInt(keyD)
	stringToSign := transaction.ID + transaction.From + transaction.To + strconv.Itoa(transaction.Amount)
	transaction.Signature = ConvertBigIntToString(rsa.FullSign(stringToSign, *n, *d))
}

func (rsa *RSA) VerifyTransaction(transaction SignedTransaction) bool {
	stringToVerify := transaction.ID + transaction.From + transaction.To + strconv.Itoa(transaction.Amount)

	signature := ConvertStringToBigInt(transaction.Signature)

	keyN := *ConvertStringToBigInt(transaction.From)

	verified := rsa.VerifyWithKey(stringToVerify, *signature, keyN, big.NewInt(3))
	return verified
}

func (rsa *RSA) FullSignBlock(block Block, keyN big.Int, keyD big.Int) Block {
	stringToSign := strings.Join(block, ":")
	fmt.Println("The string to sign:", stringToSign)
	signature := ConvertBigIntToString(rsa.FullSign(stringToSign, keyN, keyD))
	block = append(block, signature)
	return block
}

func (rsa *RSA) CreateBlockSignature(slotNumber int, nextBlockData Block, prevBlockHash string) string {
	toSign := "BLOCK" + ":" + strconv.Itoa(slotNumber) + ":" + strings.Join(nextBlockData, ":") + ":" + prevBlockHash
	return ConvertBigIntToString(rsa.FullSign(toSign, rsa.n, rsa.d))
}

//Verify that decrypting sigma with keyN gives the checkString
func (rsa *RSA) VerifyBlockSignature(slotNumber int, nextBlockData Block, prevBlockHash string, sigma string, keyN string) bool {
	checkString := "BLOCK" + ":" + strconv.Itoa(slotNumber) + ":" + strings.Join(nextBlockData, ":") + ":" + prevBlockHash
	keyNreal := *ConvertStringToBigInt(keyN)
	verified := rsa.VerifyWithKey(checkString, *ConvertStringToBigInt(sigma), keyNreal, big.NewInt(3)) //Decrypt sigma with keyNreal and check that it gives checkString
	return verified
}

func (rsa *RSA) VerifyDraw(draw string, slotNumber int, seed int, keyN string) bool {
	fmt.Println("Verifying draw: ")
	fmt.Println("slotNumber: ", slotNumber)
	fmt.Println("seed: ", seed)
	fmt.Println("keyN: ", keyN)

	checkString := "LOTTERY:" + strconv.Itoa(seed) + ":" + strconv.Itoa(slotNumber)
	fmt.Println("checkString", checkString)
	keyNreal := *ConvertStringToBigInt(keyN)
	verified := rsa.VerifyWithKey(checkString, *ConvertStringToBigInt(draw), keyNreal, big.NewInt(3))
	return verified
}

func (rsa *RSA) VerifyBlock(block Block, keyN string) bool {
	signature := ConvertStringToBigInt(block[len(block)-2])
	block = block[:len(block)-2] //-2 because we also append the delimiter in the block
	stringToVerify := strings.Join(block, ":")

	keyNreal := *ConvertStringToBigInt(keyN)

	verified := rsa.VerifyWithKey(stringToVerify, *signature, keyNreal, big.NewInt(3))
	return verified
}

func (rsa *RSA) VerifyWithKey(m string, s big.Int, keyN big.Int, keyE *big.Int) bool {
	decryptedSignature := rsa.EncryptWithKey(&s, keyN, keyE)
	fmt.Println("drawstring:", decryptedSignature)
	hashed := Hash(m)
	fmt.Println("checkString:", rsa.FullSign(m, keyN, *keyE))
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

func ConvertStringToBigInt(str string) *big.Int {
	s := new(big.Int)
	s.SetString(str, 10)
	return s
}

func ConvertBigIntToString(big *big.Int) string {
	return big.String()
}

func (rsa *RSA) Verify(m string, s big.Int) bool {
	decryptedSignature := rsa.Encrypt(&s)
	hashed := Hash(m)
	var verified bool
	if bytes.Equal(hashed.Bytes(), decryptedSignature.Bytes()) {
		verified = true
	} else {
		verified = false
	}
	return verified
}
