package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"log"
	"math"
	"strconv"
	"strings"
)

// Decimal to hexadecimal
func DecHex(n int64) string {
	if n < 0 {
		log.Println("Decimal to hexadecimal error: the argument must be greater than zero.")
		return ""
	}
	if n == 0 {
		return "0"
	}
	hex := map[int64]int64{10: 65, 11: 66, 12: 67, 13: 68, 14: 69, 15: 70}
	s := ""
	for q := n; q > 0; q = q / 16 {
		m := q % 16
		if m > 9 && m < 16 {
			m = hex[m]
			s = fmt.Sprintf("%v%v", string(m), s)
			continue
		}
		s = fmt.Sprintf("%v%v", m, s)
	}
	return s
}

// Hexadecimal to decimal
func HexDec(h string) (n int64) {
	s := strings.Split(strings.ToUpper(h), "")
	l := len(s)
	i := 0
	d := float64(0)
	hex := map[string]string{"A": "10", "B": "11", "C": "12", "D": "13", "E": "14", "F": "15"}
	for i = 0; i < l; i++ {
		c := s[i]
		if v, ok := hex[c]; ok {
			c = v
		}
		f, err := strconv.ParseFloat(c, 10)
		if err != nil {
			log.Println("Hexadecimal to decimal error:", err.Error())
			return -1
		}
		d += f * math.Pow(16, float64(l-i-1))
	}
	return int64(d)
}

// Get the hash data from []byte
func GetHashFromBytes(src []byte) ([]byte, error) {
	result := []byte("")
	temp := sha256.Sum256(src)
	result = temp[:]

	if result == nil {
		return []byte(""), errors.New("ERROR:WRONG kHashType")
	}
	return result, nil
}

// int转成[]byte
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err.Error())
	}

	return buff.Bytes()
}