package main

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"testing"
)

func TestRSAEncryptDecrypt(t *testing.T) {
	print("hello")
	rsa := MakeRSA(2048)
	var x int64
	x = 1234567
	encrypted := rsa.Encrypt(big.NewInt(x))
	decrypted := rsa.Decrypt(&encrypted)
	if !(x == decrypted.Int64()) {
		t.Error("Should have decrypted to", x, "but got", decrypted)
	} else {
		fmt.Println("TestRSAEncryptDecrypt passed")
	}
}

func TestShouldNotVerifyIncorrectSignature(t *testing.T) {
	rsa := MakeRSA(2000)
	message := "Hey Mom, what's up?"
	message2 := "Hey Dad, how are you?"

	hashed := rsa.Hash(message)
	signature := rsa.Sign(hashed)
	verified := rsa.Verify(message2, signature)

	if verified {
		t.Error("Verified something wrongly")
	} else {
		fmt.Println("TestShouldNotVerifyIncorrectSignature passed")
	}
}

func TestShouldVerifyCorrectSignature(t *testing.T) {
	rsa := MakeRSA(2000)
	message := "Hey Mom, what's up?"

	hashed := rsa.Hash(message)
	signature := rsa.Sign(hashed)
	verified := rsa.Verify(message, signature)
	if !verified {
		t.Error("Did not verify correctly")
	} else {
		fmt.Println("TestShouldVerifyCorrectSignature passed")
	}
}

func TestCanVerifyTransactionMadeFromSecretKey(t *testing.T) {
	rsa := MakeRSA(2000)
	publicKey := rsa.ConvertBigIntToString(&rsa.n)
	secretKeyD := rsa.ConvertBigIntToString(&rsa.d)
	transaction := MakeSignedTransaction(publicKey, "test", 200, secretKeyD)
	result := rsa.VerifyTransaction(*transaction)
	if result {
		fmt.Println("TestCanVerifyTransactionMadeFromSecretKey PASSED")
	} else {
		fmt.Println("TestCanVerifyTransactionMadeFromSecretKey FAILED")
	}
}

func TestMakeSignedTransaction(t *testing.T) {
	rsa := MakeRSA(2000)
	publicKey := rsa.ConvertBigIntToString(&rsa.n)
	secretKeyD := rsa.ConvertBigIntToString(&rsa.d)
	st := MakeSignedTransaction(publicKey, "test", 200, secretKeyD)
	stringToSign := st.ID + st.From + st.To + strconv.Itoa(st.Amount)
	signAsBig := rsa.FullSign(stringToSign, *rsa.ConvertStringToBigInt(publicKey), *rsa.ConvertStringToBigInt(secretKeyD))
	sign := signAsBig.Bytes()

	if st.From != publicKey {
		fmt.Println("TestMakeSignedTransaction Failed 1")
	} else if st.To != "test" {
		fmt.Println("TestMakeSignedTransaction Failed 2")
	} else if st.Amount != 200 {
		fmt.Println("TestMakeSignedTransaction Failed 3")
	} else if !bytes.Equal(st.Signature, sign) {
		fmt.Println("TestMakeSignedTransaction Failed 4, expected:", sign, "got ", st.Signature)
	} else {
		fmt.Println("TestMakeSignedTransaction Passed")
	}
}
