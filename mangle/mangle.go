/*
BAD2BEEF's Stirng Mangler
https://github.com/bad2beef/Toolbox
It just messes with strings. That's about it.
*/

package main

import (
	"bytes"
	"crypto/rc4"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type mangleFunc func([]byte, []byte) []byte

type substitution struct {
	bytesOrig []byte
	bytesSrc  []byte
	bytesDest []byte
}

func mangleRC4(key, value []byte) []byte {
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		log.Fatalf("Could not cretae RC4 ciper. %s", err)
	}

	mangled := make([]byte, len(value))
	cipher.XORKeyStream(mangled, value)

	return mangled
}

func mangleXor(key, value []byte) []byte {
	mangled := make([]byte, len(value))
	for index := range value {
		mangled[index] = value[index] ^ key[index%len(key)]
	}
	return mangled
}

func mangle(fileNameIn, fileNameOut string, substitutions []substitution) {
	fileIn, err := os.OpenFile(fileNameIn, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatalf("Could not open input file. %s", err)
	}
	defer fileIn.Close()

	var fileOut *os.File
	if fileNameOut == "-" {
		fileOut = os.Stdin
	} else {
		fileOut, err = os.OpenFile(fileNameOut, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Fatalf("Could not open output file. %s", err)
		}
		defer fileOut.Close()
	}

	fileBytes, err := ioutil.ReadAll(fileIn)
	if err != nil {
		log.Fatalf("Could not read input file. %s", err)
	}

	for _, tokenSubstitution := range substitutions {
		fileBytes = bytes.ReplaceAll(fileBytes, tokenSubstitution.bytesSrc, tokenSubstitution.bytesDest)
	}

	_, err = fileOut.Write(fileBytes)
	if err != nil {
		log.Fatalf("Could not write to output file. %s", err)
	}
}

func main() {
	// Collect arguments
	if len(os.Args) < 6 {
		fmt.Println("Usage: mangle [xor|rc4] [old_key|\"\"] [new_key] [file_in] [file_out|-] [TOKEN...]")
		os.Exit(2)
	}

	algo := os.Args[1]
	keyOld := os.Args[2]
	keyNew := os.Args[3]
	fileNameIn := os.Args[4]
	fileNameOut := os.Args[5]

	// Select mangling algorithim
	var mangleAlgo mangleFunc
	switch strings.ToLower(algo) {
	case "rc4":
		mangleAlgo = mangleRC4
	case "xor":
		fallthrough
	default:
		mangleAlgo = mangleXor
	}

	// Prepare keys
	keyBytesOld := []byte(keyOld)
	keyBytesNew := []byte(keyNew)

	// Build substitution list
	var substitutions []substitution
	if len(keyBytesOld) > 0 {
		substitutions = append(substitutions, substitution{bytesOrig: keyBytesOld, bytesSrc: keyBytesOld, bytesDest: keyBytesNew})
	}

	var tokenBytesOld, tokenBytesNew []byte
	for _, token := range os.Args[6:] {
		tokenBytes := []byte(token)
		tokenBytesNew = mangleAlgo(keyBytesNew, tokenBytes)

		if len(keyBytesOld) > 0 {
			tokenBytesOld = mangleAlgo(keyBytesOld, tokenBytes)
		} else {
			tokenBytesOld = tokenBytes
		}

		tokenSubstitution := substitution{bytesOrig: tokenBytes, bytesSrc: tokenBytesOld, bytesDest: tokenBytesNew}
		substitutions = append(substitutions, tokenSubstitution)
	}

	if fileNameOut != "-" {
		for _, substitution := range substitutions {
			fmt.Printf("\"%s\": % #x -> % #x\n", substitution.bytesOrig, substitution.bytesSrc, substitution.bytesDest)
		}
	}

	mangle(fileNameIn, fileNameOut, substitutions)
}
