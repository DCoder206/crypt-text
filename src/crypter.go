package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: crypter.exe [--read] <filename>")
		return
	}
	var filename string
	readMode := false
	if os.Args[1] == "--read" {
		if len(os.Args) < 3 {
			fmt.Println("Error: Specify the filename to read.")
			return
		}
		readMode = true
		filename = os.Args[2]
	} else {
		filename = os.Args[1]
	}
	if filepath.Ext(filename) != "" {
		fmt.Println("Error: Filename must not contain an extension. Use a plain name (e.g. 'myfile')")
		return
	}
	invalidChars := []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}
	for _, char := range filename {
		if unicode.IsControl(char) || strings.ContainsRune(string(invalidChars), char) {
			fmt.Println("Error: Filename contains invalid characters. Avoid using / \\ : * ? \" < > | and control characters.")
			return
		}
	}
	key := []byte("examplekey123456examplekey123456") // Replace with secure key generation
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Error creating GCM:", err)
		return
	}
	if readMode {
		ciphertext, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			fmt.Println("Error: Corrupted file or invalid content.")
			return
		}
		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			fmt.Println("Error decrypting file:", err)
			return
		}
		fmt.Println("Decrypted Content:")
		fmt.Println(string(plaintext))
		return
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error enabling raw mode:", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	var textBuffer strings.Builder
	fmt.Print("\033[2J\033[H")
	fmt.Println("Simple Text Editor (Press Ctrl+Q to save and exit)")
	fmt.Println("-------------------------------------------------")
	for {
		fmt.Print("\033[u")
		fmt.Print("\033[s")
		buf := make([]byte, 1)
		_, err := os.Stdin.Read(buf)
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}
		if buf[0] == 17 {
			break
		}
		if buf[0] == 127 {
			if textBuffer.Len() > 0 {
				currentText := textBuffer.String()
				currentText = currentText[:len(currentText)-1]
				textBuffer.Reset()
				textBuffer.WriteString(currentText)
				fmt.Print("\b \b")
			}
			continue
		}
		textBuffer.WriteByte(buf[0])
		fmt.Print(string(buf[0]))
	}
	plaintext := []byte(textBuffer.String())
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Error generating nonce:", err)
		return
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	err = os.WriteFile(filename, ciphertext, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Printf("Encrypted text saved to %s\n", filename)
}
