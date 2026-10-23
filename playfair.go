// Q replaced by K before encryption

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

var (
	alphabet       = []rune("AĂÂBCDEFGHIÎJKLMNOPQRSȘTȚUVWXYZ") // 31 letters
	matrixAlphabet = []rune("AĂÂBCDEFGHIÎJKLMNOPRSȘTȚUVWXYZ")  // 30 letters (no Q)
	rows           = 5
	cols           = 6
	filler         = 'X'
	fillerAlt      = 'Z'
	minKeyLen      = 7
)

// normalize replaces cedilla forms (Ş→Ș, Ţ→Ț) with comma-below forms.
var normalizeMap = map[rune]rune{
	'Ş': 'Ș', 'ş': 'ș',
	'Ţ': 'Ț', 'ţ': 'ț',
}

func normalize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if rep, ok := normalizeMap[r]; ok {
			b.WriteRune(rep)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func inAlphabet(r rune) bool {
	for _, a := range alphabet {
		if a == r {
			return true
		}
	}
	return false
}

func validate(text string) []rune {
	text = normalize(text)
	var bad []rune
	for _, r := range text {
		if r != ' ' && !inAlphabet(unicode.ToUpper(r)) {
			found := false
			for _, b := range bad {
				if b == r {
					found = true
					break
				}
			}
			if !found {
				bad = append(bad, r)
			}
		}
	}
	return bad
}

func toLetters(text string) []rune {
	text = normalize(text)
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ToUpper(text)
	text = strings.ReplaceAll(text, "Q", "K")
	return []rune(text)
}

func processKey(key string) []rune {
	letters := toLetters(key)
	var result []rune
	seen := make(map[rune]bool)
	for _, r := range letters {
		if !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}
	return result
}

type matrix struct {
	grid [5][6]rune
	pos  map[rune][2]int // letter -> {row, col}
}

func buildMatrix(key string) matrix {
	seq := processKey(key)
	seen := make(map[rune]bool)
	for _, r := range seq {
		seen[r] = true
	}
	for _, r := range matrixAlphabet {
		if !seen[r] {
			seq = append(seq, r)
		}
	}

	var m matrix
	m.pos = make(map[rune][2]int)
	for i, r := range seq {
		row, col := i/cols, i%cols
		m.grid[row][col] = r
		m.pos[r] = [2]int{row, col}
	}
	return m
}

func prepareMessage(message string) []string {
	letters := toLetters(message)
	var pairs []string
	i := 0
	for i < len(letters) {
		a := letters[i]
		if i+1 < len(letters) && letters[i+1] != a {
			pairs = append(pairs, string([]rune{a, letters[i+1]}))
			i += 2
		} else {
			f := filler
			if a == filler {
				f = fillerAlt
			}
			pairs = append(pairs, string([]rune{a, f}))
			i++
		}
	}
	return pairs
}

func transformPair(pair string, m matrix, d int) (string, string) {
	runes := []rune(pair)
	p1 := m.pos[runes[0]]
	p2 := m.pos[runes[1]]
	r1, c1 := p1[0], p1[1]
	r2, c2 := p2[0], p2[1]

	if r1 == r2 { // same row
		return string([]rune{
			m.grid[r1][((c1+d)%cols+cols)%cols],
			m.grid[r2][((c2+d)%cols+cols)%cols],
		}), "row"
	}
	if c1 == c2 { // same column
		return string([]rune{
			m.grid[((r1+d)%rows+rows)%rows][c1],
			m.grid[((r2+d)%rows+rows)%rows][c2],
		}), "column"
	}
	// rectangle
	return string([]rune{m.grid[r1][c2], m.grid[r2][c1]}), "rectangle"
}

func printMatrix(m matrix) {
	fmt.Println("Encryption matrix (5 x 6):")
	for r := 0; r < rows; r++ {
		fmt.Print("    ")
		for c := 0; c < cols; c++ {
			if c > 0 {
				fmt.Print("   ")
			}
			fmt.Printf("%c", m.grid[r][c])
		}
		fmt.Println()
	}
}

func printPairs(title string, pairs []string, perLine int) {
	fmt.Println(title)
	for i := 0; i < len(pairs); i += perLine {
		end := i + perLine
		if end > len(pairs) {
			end = len(pairs)
		}
		fmt.Println("    " + strings.Join(pairs[i:end], " "))
	}
}

func encrypt(key, message string) string {
	pairs := prepareMessage(message)
	m := buildMatrix(key)
	fmt.Printf("Processed key: %s\n", string(processKey(key)))
	printMatrix(m)
	fmt.Printf("Message without spaces, upper case (Q -> K): %s\n", string(toLetters(message)))
	printPairs("Prepared message (pairs):", pairs, 12)

	var out []string
	var steps []string
	for _, p := range pairs {
		q, rule := transformPair(p, m, +1)
		out = append(out, q)
		steps = append(steps, fmt.Sprintf("%s->%s %-9s", p, q, rule))
	}
	printPairs("Pair transformations:", steps, 4)
	printPairs("Encrypted pairs:", out, 12)

	var result strings.Builder
	for _, o := range out {
		result.WriteString(o)
	}
	return result.String()
}

func decrypt(key, cryptogram string) string {
	letters := toLetters(cryptogram)
	var pairs []string
	for i := 0; i < len(letters); i += 2 {
		pairs = append(pairs, string(letters[i:i+2]))
	}

	m := buildMatrix(key)
	fmt.Printf("Processed key: %s\n", string(processKey(key)))
	printMatrix(m)
	printPairs("Cryptogram split into pairs:", pairs, 12)

	var out []string
	var steps []string
	for _, p := range pairs {
		q, rule := transformPair(p, m, -1)
		out = append(out, q)
		steps = append(steps, fmt.Sprintf("%s->%s %-9s", p, q, rule))
	}
	printPairs("Pair transformations:", steps, 4)
	printPairs("Decrypted pairs:", out, 12)

	var result strings.Builder
	for _, o := range out {
		result.WriteString(o)
	}
	return result.String()
}

func readLine(scanner *bufio.Scanner, prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func readText(scanner *bufio.Scanner, prompt string, minLetters int) string {
	for {
		text := readLine(scanner, prompt)
		bad := validate(text)
		if len(bad) > 0 {
			parts := make([]string, len(bad))
			for i, b := range bad {
				parts[i] = fmt.Sprintf("'%c'", b)
			}
			fmt.Printf("  [!] Characters not allowed: %s\n", strings.Join(parts, " "))
			fmt.Println("      Use only 'A'-'Z', 'a'-'z' and the Romanian letters Ă Â Î Ș Ț / ă â î ș ț (and spaces).")
			continue
		}
		n := len(toLetters(text))
		if n < minLetters {
			fmt.Printf("  [!] At least %d letters are required (entered: %d).\n", minLetters, n)
			continue
		}
		return text
	}
}

func readCryptogram(scanner *bufio.Scanner) string {
	for {
		letters := toLetters(readText(scanner, "Cryptogram: ", 1))
		if len(letters)%2 != 0 {
			fmt.Println("  [!] A Playfair cryptogram has an even number of letters.")
			continue
		}
		hasDup := false
		for i := 0; i < len(letters); i += 2 {
			if letters[i] == letters[i+1] {
				hasDup = true
				break
			}
		}
		if hasDup {
			fmt.Println("  [!] A Playfair cryptogram cannot contain a pair of equal letters.")
			continue
		}
		return string(letters)
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("  PLAYFAIR CIPHER  romanian alphabet 31 letters, 5x6 matrix")


	for {
		fmt.Println("\n1 - Encrypt   2 - Decrypt   0 - Exit")
		op := readLine(scanner, "Choose the operation: ")
		if op == "0" {
			break
		}
		if op != "1" && op != "2" {
			fmt.Println("  [!] Enter 1, 2 or 0.")
			continue
		}

		key := readText(scanner, fmt.Sprintf("Key (at least %d letters): ", minKeyLen), minKeyLen)

		if op == "1" {
			message := readText(scanner, "Message: ", 1)
			fmt.Printf("\nCRYPTOGRAM: %s\n", encrypt(key, message))
		} else {
			cryptogram := readCryptogram(scanner)
			fmt.Printf("\nDECRYPTED MESSAGE: %s\n", decrypt(key, cryptogram))
			fmt.Println("(Remove the fillers X/Z and insert the spaces manually.)")
		}
	}
}
