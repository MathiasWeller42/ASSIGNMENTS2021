package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"os"
	"time"
)

var n big.Int
var p *big.Int
var q *big.Int
var d big.Int
var e *big.Int

func KeyGen(k int) {
	//bit length of p*q = n is k
	e = big.NewInt(3) //given by the hand-in text
	p = GeneratePrime(k)
	q = GeneratePrime(k)
	n.Mul(p, q)
	GenerateD()
}

func Encrypt(m *big.Int) big.Int {
	//c = m ^ e mod n
	var c big.Int
	c.Exp(m, e, &n)
	return c
}

func Decrypt(c *big.Int) big.Int {
	//m = c ^ d mod n
	var m big.Int
	m.Exp(c, &d, &n)
	return m
}

func GenerateD() {
	var x, y, z big.Int
	x.Mul(y.Sub(p, big.NewInt(1)), z.Sub(q, big.NewInt(1)))
	d.ModInverse(e, &x)
}

func GeneratePrime(k int) *big.Int {
	for {
		prime, err := rand.Prime(rand.Reader, k/2) //set the length to k/2
		if err != nil {
			fmt.Print(e) //print the error and try again
			continue
		}
		var x big.Int
		x.Mod(x.Sub(prime, big.NewInt(1)), e)
		mutuallyPrime := (x) //this is (prime-1) mod e

		//prime should be a probable prime
		//mutuallyPrime should not be 0 (this means that gcd(prime-1,e) is 1 when e=3)
		if prime.ProbablyPrime(1) && (mutuallyPrime.Cmp(big.NewInt(0)) != 0) {
			return prime
		}
	}
}

func GenerateRSAKeys() ([]byte, big.Int, *big.Int) {
	KeyGen(2000)
	return d.Bytes(), n, e
}

func RunRSAEncryptionDecryption() {
	KeyGen(2048)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter the integer to be encrypted:")
	scanner.Scan()
	M := scanner.Text()
	fmt.Println("Encrypting now")
	var x int64
	fmt.Sscan(M, &x)
	encrypted := Encrypt(big.NewInt(x))
	fmt.Println("Encrypted:", encrypted)
	fmt.Println("Decrypting now")
	decrypted := Decrypt(&encrypted)
	fmt.Println("Decrypted:", decrypted)
}

func hash(message string) *big.Int {
	messageAsByte := []byte(message)
	hashed := sha256.Sum256(messageAsByte)
	x := big.NewInt(0).SetBytes(hashed[:])
	return x
}

func sign(m *big.Int) big.Int {
	signature := Decrypt(m)
	return signature
}

func verify(m string, s big.Int) bool {
	decryptedSignature := Encrypt(&s)
	hashed := hash(m)
	var verified bool
	if bytes.Compare(hashed.Bytes(), decryptedSignature.Bytes()) == 0 {
		verified = true
	} else {
		verified = false
	}
	return verified
}

func testVerification(m1, m2 string) bool {
	message := m1
	hashed := hash(message)
	signature := sign(hashed)
	verified := verify(m2, signature)
	return verified
}

func timeVerification(noOfBytes int) {
	fmt.Println("Measuring time to hash", noOfBytes, "bytes")
	randomArray := make([]byte, noOfBytes)
	rand.Read(randomArray)
	start := time.Now()
	hash(string(randomArray))
	elapsed := time.Since(start)
	fmt.Println("Time elapsed: ", elapsed)
	secs := float64(elapsed.Milliseconds()) / 1000.0
	if secs == 0 {
		fmt.Println("Elapsed time was rounded to 0")
	} else {
		bitsPrSec := float64(noOfBytes*8) / secs
		fmt.Println("This gives ", bitsPrSec, " bits per second")
	}
}

func main() {
	//Generating keys for all the tests
	KeyGen(2000)

	fmt.Println("Test 1:")
	message := "Hey Mom, what's up?"
	fmt.Println("Hashing and signing message", message)
	verified := testVerification(message, message)
	fmt.Println("Verification worked?", verified)

	message2 := "Hey Dad, how are you?"
	fmt.Println("Using the wrong signature for message", message2)
	verified2 := testVerification(message2, message)
	fmt.Println("Wrong verification worked?", verified2)

	fmt.Println("\n Test 2:")
	timeVerification(1000000)
	timeVerification(10000000)
	timeVerification(100000000)
	timeVerification(1000000000)

	fmt.Println("\n Test 3: ")
	fmt.Println("Testing how much time it takes to sign a hashed message: ")
	message3 := "Hello bro"
	hashed := hash(message3)
	start := time.Now()
	sign(hashed)
	elapsed := time.Since(start)
	fmt.Println("Time elapsed: ", elapsed)

	fmt.Println("What's up bros?")
	fmt.Println("The fucking ceiling")
	fmt.Println("kThXbaYi uwu luv u bae RAWR XD l0l ROFLMAO 10hi f9s ur my <3 simzebazze (sowy im sooo random :P *blushes* *tips fedora*)")
}
