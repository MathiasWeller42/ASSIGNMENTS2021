package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var n *big.Int
var p *big.Int
var q *big.Int
var d *big.Int
var e *big.Int

func KeyGen(k int) {
	//bit length of pq = n is k
	e = big.NewInt(3) //given by the hand-in text
	p = GeneratePrime(k)
	q = GeneratePrime(k)
	n.Mul(p, q)
	d = GenerateD()
}

func Encrypt(m *big.Int) *big.Int {
	//c = m ^ e mod n
	var c *big.Int
	c.Exp(m, e, n)
	return c
}

func Decrypt(c *big.Int) *big.Int {
	//m = c ^ d mod n
	var m *big.Int
	m.Exp(c, d, n)
	return m
}

func GenerateD() *big.Int {
	/*var x *big.Int
	var y *big.Int
	e.ModInverse(x.Sub(p, big.NewInt(1)),y.Sub(q, big.NewInt(1)))*/
	return big.NewInt(1)
}

func GeneratePrime(k int) *big.Int {
	for {
		prime, err := rand.Prime(rand.Reader, k/2)
		if err != nil {
			fmt.Print(e)
			continue
		}
		var x *big.Int
		x.Mod((x.Sub(prime, big.NewInt(1))), e)
		mutuallyPrime := (*x)

		if prime.ProbablyPrime(1) && (mutuallyPrime.Cmp(big.NewInt(0)) != 0) { //what should parameter be for ProbablyPrime???
			return prime
		}
	}
}

func mult(int1 *big.Int, int2 *big.Int) uint64 {
	res := int1.Mul(int1, int2)
	res2 := res.Mul(res, int1)
	return res2.Uint64()

}

func main() {
	/*
		KeyGen(2048)
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter the integer to be encrypted:")
		scanner.Scan()
		M := scanner.Text()
		fmt.Println("Encrypting now")
		var n int64
		fmt.Sscan(M, &n)
		encrypted := Encrypt(big.NewInt(n))
		fmt.Println("Encrypted:", encrypted)
		fmt.Println("Decrypting now")
		decrypted := Decrypt(encrypted)
		fmt.Println("Decrypted:", decrypted)
	*/
	fmt.Println("Result: ", (mult(big.NewInt(3), big.NewInt(4))))
}
