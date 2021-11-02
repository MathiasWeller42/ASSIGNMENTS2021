package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
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

func (rsa *RSA) VerifyWithKey(m string, s big.Int, keyN big.Int, keyE *big.Int) bool {
	decryptedSignature := rsa.EncryptWithKey(&s, keyN, keyE)
	hashed := Hash(m)
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

func Generate(filename string, password string) string {
	f, _ := os.Create(filename)
	defer f.Close()
	rsa := MakeRSA(2048)
	secretKeyString := ConvertBigIntToString(&rsa.d) + ":" + ConvertBigIntToString(&rsa.n)

	hashedPW := Hash(password).Bytes()

	passwordHashToSave, _ := bcrypt.GenerateFromPassword(hashedPW, bcrypt.DefaultCost)

	aes := MakeAES()

	encrypted := aes.Encrypt(secretKeyString, hashedPW)

	toWrite := append(passwordHashToSave, encrypted...)
	f.Write(toWrite)

	return ConvertBigIntToString(rsa.e) + ":" + ConvertBigIntToString(&rsa.n)
}

func Sign(filename string, password string, msg []byte) (string, error) {

	read_bytes, err := os.ReadFile(filename)

	if err != nil {
		return "Yo wtf", errors.New("Tried to read non-existing file")
	}

	hashedPW := Hash(password).Bytes()

	aes := MakeAES()
	rsa := MakeRSA(2048)

	savedPasswordHash := read_bytes[:60]

	if bcrypt.CompareHashAndPassword(savedPasswordHash, hashedPW) == nil {
		//password correct
		decrypted := aes.Decrypt(string(read_bytes[60:]), hashedPW)

		dBigInt, nBigInt := ConvertKeyToBigInts(decrypted)
		mString := string(msg)
		signed := rsa.FullSign(mString, *nBigInt, *dBigInt)
		return ConvertBigIntToString(signed), nil

	} else {
		//error
		return "That ain't it, Chief", errors.New("Invalid password")
	}
}
func ConvertKeyToBigInts(publicKey string) (*big.Int, *big.Int) {
	secretKeyAsArray := strings.Split(publicKey, ":")
	firstString := secretKeyAsArray[0]
	secondString := secretKeyAsArray[1]
	firstBigInt := ConvertStringToBigInt(firstString)
	secondBigInt := ConvertStringToBigInt(secondString)
	return firstBigInt, secondBigInt
}

/*
func main() {

	//The filename to be written to, the password to be used and the message to sign:
	filename := "hej"
	password := "p4SSw0rd1234!"
	message := "Matti the Welle-boy er den sejeste <3<3<3 for realz j k rowling in the deep"
	messageBytes := []byte(message)

	publicKey := Generate(filename, password)
	eBigInt, nBigInt := ConvertKeyToBigInts(publicKey)

	fmt.Println("First we try to generate and sign correctly")

	signature, err := Sign(filename, password, messageBytes)

	fmt.Println("Checking that signing went well")

	if err != nil {
		fmt.Println("Something went horribly wrong with signing *fire and blood*")
		return
	} else {
		fmt.Println("Signing errors be cancelled maaaannnn *swag*")
	}

	fmt.Println("------------------------------------------------------------------------------------------")
	fmt.Println("Checking verification")

	//Verifying that the public key can be used to verify the signature of the message:
	rsa := MakeRSA(2048)
	if !rsa.VerifyWithKey(message, *ConvertStringToBigInt(signature), *nBigInt, eBigInt) {
		fmt.Println("Something went badly when verifying *surprised pikachu*")
		return
	} else {
		fmt.Println("The signature of the message:", message, ", has been verified with the public key returned from Generate *woah dude that's tight yo*")
	}

	fmt.Println("------------------------------------------------------------------------------------------")
	fmt.Println("Now we verify that Sign fails with an invalid password: ")
	wrongPassword := "NotTheCorrectOneOopsieWhoopsie"

	_, err2 := Sign(filename, wrongPassword, messageBytes)

	if err2 == nil {
		fmt.Println("Mistakes were made")
		return
	} else {
		fmt.Println("A wrong password does not work - yay! XD")
	}

	fmt.Println("------------------------------------------------------------------------------------------")
	fmt.Println("Now we verify that Sign fails with an invalid password: ")
	wrongFilename := "hejhej"

	_, err3 := Sign(wrongFilename, password, messageBytes)

	if err3 == nil {
		fmt.Println("Mistakes were made")
		return
	} else {
		fmt.Println("A wrong filename does not work *happy dance*")
	}

	fmt.Println("------------------------------------------------------------------------------------------")
	fmt.Println("This is how long the bcrypt checking takes:")

	read_bytes, err := os.ReadFile(filename)
	hashedPW := Hash(password).Bytes()

	savedPasswordHash := read_bytes[:60]
	start := time.Now()
	bcrypt.CompareHashAndPassword(savedPasswordHash, hashedPW)
	elapsed := time.Since(start)
	fmt.Println("Time elapsed: ", elapsed)

}
*/
