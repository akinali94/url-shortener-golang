package main


//Excluded 0 (zero), I, O (ou), l
const Base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz" 

func Base10toBase58(num uint64) string {

	var Base58Alphabet string
	
	if num == 0 {
		return "1" // Base-58 equivalent of zero
	}

	base := uint64(58)
	result := ""

	for num > 0 {
		remainder := num % base
		num /= base
		result = string(Base58Alphabet[remainder]) + result
	}

	return result
}