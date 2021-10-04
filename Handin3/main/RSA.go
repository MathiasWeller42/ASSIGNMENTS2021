package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
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

func Encrypt(m *big.Int, e *big.Int, n big.Int) big.Int {
	//c = m ^ e mod n
	var c big.Int
	c.Exp(m, e, &n)
	return c
}

func Decrypt(c *big.Int, d big.Int, n big.Int) big.Int {
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
		//what should parameter be for ProbablyPrime???
		if prime.ProbablyPrime(1) && (mutuallyPrime.Cmp(big.NewInt(0)) != 0) {
			return prime
		}
	}
}

func GenerateRSAKeys() ([]byte, big.Int, *big.Int) {
	KeyGen(2000)
	return d.Bytes(), n, e
}

/*
func primeTest(k int) {
	prime, err := rand.Prime(rand.Reader, k/2)
	if err != nil {
		fmt.Print(e)
	}
	var x big.Int
	fmt.Println("This is the prime:", prime)
	e = big.NewInt(9159481)
	fmt.Println("subtraction: ", x.Mod((x.Sub(prime, big.NewInt(1))), e))
	mutuallyPrime := big.NewInt(10000000)
	prime = big.NewInt(10000000)
	if prime.ProbablyPrime(1) && (mutuallyPrime.Cmp(big.NewInt(0)) != 0) { //what should parameter be for ProbablyPrime???
		fmt.Println("Test passed with prime: ", prime)
	} else {
		fmt.Println("This ain't prime, y'all")
	}
}*/

func RunRSA() {
	KeyGen(2048)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter the integer to be encrypted:")
	scanner.Scan()
	M := scanner.Text()
	fmt.Println("Encrypting now")
	var x int64
	fmt.Sscan(M, &x)
	encrypted := Encrypt(big.NewInt(x), e, n)
	fmt.Println("Encrypted:", encrypted)
	fmt.Println("Decrypting now")
	decrypted := Decrypt(&encrypted, d, n)
	fmt.Println("Decrypted:", decrypted)
}
