package main

// Import beberapa package yang diperlukan:
// "bufio" untuk input/output buffered,
// "encoding/json" untuk encoding dan decoding JSON,
// "fmt" untuk output teks,
// "log" untuk logging error,
// "os" untuk operasi sistem file,
// "strconv" untuk konversi string ke tipe lain,
// "sync" untuk penggunaan mutex (penguncian),
// "time" untuk penanganan waktu.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// Definisi struct untuk Account yang merepresentasikan akun pengguna.
type Account struct {
	ID           string           // ID akun
	Password     string           // Password akun
	Balance      float64          // Saldo akun
	Approved     bool             // Status apakah akun disetujui
	Transactions []Transaction    // Daftar transaksi yang terkait dengan akun
}

// Definisi struct untuk Transaction yang merepresentasikan transaksi.
type Transaction struct {
	ID        int     // ID transaksi
	AccountID string  // ID akun yang terkait dengan transaksi
	Type      string  // Jenis transaksi (misalnya transfer atau pembayaran)
	Amount    float64 // Jumlah transaksi
	Date      string  // Tanggal transaksi
	Details   string  // Rincian transaksi
}

// Definisi struct untuk Registration yang menyimpan data registrasi akun.
type Registration struct {
	ID       string // ID akun untuk registrasi
	Password string // Password akun
}

// Definisi struct untuk TopUpRequest yang merepresentasikan permintaan top-up saldo.
type TopUpRequest struct {
	ID        int     // ID permintaan top-up
	AccountID string  // ID akun yang meminta top-up
	Amount    float64 // Jumlah top-up
	Date      string  // Tanggal permintaan
	Approved  bool    // Status apakah permintaan disetujui
}

// Variabel global untuk menyimpan data akun, registrasi, permintaan top-up, mutex untuk sinkronisasi thread, pengguna saat ini, dan ID top-up berikutnya.
var (
	accounts      = make(map[string]*Account)  // Peta untuk menyimpan data akun
	registrations = make(map[string]Registration)  // Peta untuk menyimpan data registrasi
	topUpRequests = []TopUpRequest{}           // Slice untuk menyimpan permintaan top-up
	mu            sync.Mutex                   // Mutex untuk mengamankan akses ke data bersama
	currentUser   *Account                     // Menyimpan data pengguna saat ini
	nextTopUpID   = 1                          // ID top-up berikutnya
)

const accountsFilePath = "accounts.json" // Lokasi file untuk menyimpan data akun

// Fungsi utama program yang memuat data akun, menyediakan pilihan login atau registrasi, dan menampilkan menu utama.
func main() {
	loadAccounts()  // Memuat data akun dari file
	defer saveAccounts()  // Menyimpan data akun ke file saat program berakhir

	scanner := bufio.NewScanner(os.Stdin)  // Membuat scanner untuk menerima input dari pengguna
	fmt.Println("Welcome to the e-money system!")

	// Jika file akun kosong, maka pengguna akan diminta untuk registrasi atau login sebagai admin.
	if isEmptyAccountsFile() {
		fmt.Println("No accounts found. Please register as admin or register a new account.")
		for {
			fmt.Println("1. Admin Login")
			fmt.Println("2. Register Account")
			fmt.Println("3. Exit")

			if !scanner.Scan() {
				break
			}

			option := scanner.Text()  // Membaca pilihan dari pengguna
			switch option {
			case "1":
				// Memeriksa login admin dan menampilkan menu admin jika berhasil.
				if loginadm(scanner) {
					adminMenu(scanner)
				}
			case "2":
				// Registrasi akun baru.
				registerAccount(scanner)
			case "3":
				// Keluar dari program.
				fmt.Println("Goodbye!")
				return
			default:
				// Menangani input yang tidak valid.
				fmt.Println("Invalid option. Please try again.")
			}
		}
	} else {
		// Jika akun sudah ada, pengguna dapat login sebagai user atau admin, atau registrasi akun baru.
		fmt.Println("Choose an option:")
		fmt.Println("1. User Login")
		fmt.Println("2. Admin Login")
		fmt.Println("3. Register Account")
		fmt.Println("4. Exit")

		for {
			if !scanner.Scan() {
				break
			}

			option := scanner.Text()
			switch option {
			case "1":
				// Login sebagai pengguna
				if loginusr(scanner) {
					userMenu(scanner)
				}
			case "2":
				// Login sebagai admin
				if loginadm(scanner) {
					adminMenu(scanner)
				}
			case "3":
				// Registrasi akun baru
				registerAccount(scanner)
			case "4":
				// Keluar dari program
				fmt.Println("Goodbye!")
				return
			default:
				// Menangani input yang tidak valid.
				fmt.Println("Invalid option. Please try again.")
			}
		}
	}
}

// Fungsi untuk memeriksa apakah file akun kosong.
func isEmptyAccountsFile() bool {
	data, err := os.ReadFile(accountsFilePath)  // Membaca data dari file
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("Failed to read accounts file:", err)  // Menangani error jika gagal membaca file
		}
		return true  // Mengembalikan true jika file tidak ditemukan atau kosong
	}
	return len(data) == 0  // Mengembalikan true jika ukuran data adalah 0 (file kosong)
}

// Fungsi untuk registrasi akun baru.
func registerAccount(scanner *bufio.Scanner) {
	fmt.Print("Enter account ID: ")
	if !scanner.Scan() {
		return
	}
	id := scanner.Text()

	fmt.Print("Enter password: ")
	if !scanner.Scan() {
		return
	}
	password := scanner.Text()

	mu.Lock()  // Mengunci mutex untuk menghindari race condition saat mengakses peta
	defer mu.Unlock()

	if _, exists := accounts[id]; exists {
		if _, exists := registrations[id]; exists {
			fmt.Println("Account already exists.")  // Menangani kasus jika akun sudah ada
			return
		}
	}

	registrations[id] = Registration{ID: id, Password: password}  // Menambahkan registrasi baru ke peta
	fmt.Println("Registration submitted for approval.")
}

// Fungsi login admin.
func loginadm(scanner *bufio.Scanner) bool {
	fmt.Print("Enter admin ID: ")
	if !scanner.Scan() {
		return false
	}
	id := scanner.Text()

	fmt.Print("Enter password: ")
	if !scanner.Scan() {
		return false
	}
	password := scanner.Text()

	mu.Lock()  // Mengunci mutex sebelum memeriksa data
	defer mu.Unlock()

	if id == "admin" && password == "admin" {
		currentUser = &Account{ID: "admin", Password: "admin"}  // Login berhasil, menetapkan currentUser sebagai admin
		return true
	}

	fmt.Println("Invalid admin credentials.")  // Login gagal
	return false
}

// Fungsi login user.
func loginusr(scanner *bufio.Scanner) bool {
	fmt.Print("Enter account ID: ")
	if !scanner.Scan() {
		return false
	}
	id := scanner.Text()

	fmt.Print("Enter password: ")
	if !scanner.Scan() {
		return false
	}
	password := scanner.Text()

	mu.Lock()
	defer mu.Unlock()

	account, exists := accounts[id]
	if !exists || account.Password != password || !account.Approved {
		fmt.Println("Invalid credentials or account not approved.")  // Menangani error login
		return false
	}

	currentUser = account  // Login berhasil
	return true
}

// Fungsi untuk menampilkan menu user setelah login berhasil.
func userMenu(scanner *bufio.Scanner) {
	for {
		fmt.Println("1. Check Balance")
		fmt.Println("2. Transfer Money")
		fmt.Println("3. Make Payment")
		fmt.Println("4. Print Transaction History")
		fmt.Println("5. Top Up Balance")
		fmt.Println("6. Logout")

		if !scanner.Scan() {
			break
		}

		option := scanner.Text()
		switch option {
		case "1":
			checkBalance()  // Memeriksa saldo
		case "2":
			transferMoney(scanner)  // Transfer uang
		case "3":
			makePayment(scanner)  // Membayar
		case "4":
			printTransactionHistory()  // Menampilkan riwayat transaksi
		case "5":
			topUpBalance(scanner)  // Top-up saldo
		case "6":
			currentUser = nil  // Logout
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

// Fungsi untuk menampilkan menu admin setelah login berhasil
func adminMenu(scanner *bufio.Scanner) {
	for {
		fmt.Println("1. Approve/Reject Registration")
		fmt.Println("2. Print Account List")
		fmt.Println("3. Approve/Reject Top Up Requests")
		fmt.Println("4. Logout")

		if !scanner.Scan() {
			break
		}

		option := scanner.Text()
		switch option {
		case "1":
			handleRegistrations(scanner)
		case "2":
			printAccountList()
		case "3":
			handleTopUpRequests(scanner)
		case "4":
			currentUser = nil
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func handleRegistrations(scanner *bufio.Scanner) {
	mu.Lock()
	defer mu.Unlock()

	for id, reg := range registrations {
		fmt.Printf("Approve account %s? (y/n): ", id)
		if !scanner.Scan() {
			return
		}
		if scanner.Text() == "y" {
			accounts[id] = &Account{ID: id, Password: reg.Password, Balance: 0, Approved: true}
			fmt.Println("Account approved.")
		} else {
			fmt.Println("Account rejected.")
		}
		delete(registrations, id)
	}
}

func printAccountList() {
	mu.Lock()
	defer mu.Unlock()

	fmt.Println("Account List:")
	for _, account := range accounts {
		fmt.Printf("ID: %s, Balance: %.2f, Approved: %v\n", account.ID, account.Balance, account.Approved)
	}
}

func checkBalance() {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Balance for account %s: %.2f\n", currentUser.ID, currentUser.Balance)
}

func transferMoney(scanner *bufio.Scanner) {
	fmt.Print("Enter recipient account ID: ")
	if !scanner.Scan() {
		return
	}
	recipientID := scanner.Text()

	fmt.Print("Enter amount to transfer: ")
	if !scanner.Scan() {
		return
	}
	amountStr := scanner.Text()
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		fmt.Println("Invalid amount.")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	recipient, exists := accounts[recipientID]
	if !exists {
		fmt.Println("Recipient account not found.")
		return
	}

	if currentUser.Balance < amount {
		fmt.Println("Insufficient funds.")
		return
	}

	currentUser.Balance -= amount
	recipient.Balance += amount

	transaction := Transaction{
		ID:        len(currentUser.Transactions) + 1,
		AccountID: currentUser.ID,
		Type:      "Transfer",
		Amount:    amount,
		Date:      time.Now().Format(time.RFC3339),
		Details:   fmt.Sprintf("Transferred to %s", recipientID),
	}
	currentUser.Transactions = append(currentUser.Transactions, transaction)

	recipient.Transactions = append(recipient.Transactions, Transaction{
		ID:        len(recipient.Transactions) + 1,
		AccountID: recipient.ID,
		Type:      "Transfer",
		Amount:    amount,
		Date:      time.Now().Format(time.RFC3339),
		Details:   fmt.Sprintf("Received from %s", currentUser.ID),
	})

	fmt.Println("Transfer successful.")
}

func makePayment(scanner *bufio.Scanner) {
	fmt.Print("Enter payment type (e.g., food, phone, electricity, BPJS): ")
	if !scanner.Scan() {
		return
	}
	paymentType := scanner.Text()

	fmt.Print("Enter amount to pay: ")
	if !scanner.Scan() {
		return
	}
	amountStr := scanner.Text()
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		fmt.Println("Invalid amount.")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if currentUser.Balance < amount {
		fmt.Println("Insufficient funds.")
		return
	}

	currentUser.Balance -= amount

	transaction := Transaction{
		ID:        len(currentUser.Transactions) + 1,
		AccountID: currentUser.ID,
		Type:      "Payment",
		Amount:    amount,
		Date:      time.Now().Format(time.RFC3339),
		Details:   paymentType,
	}
	currentUser.Transactions = append(currentUser.Transactions, transaction)

	fmt.Println("Payment successful.")
}

func printTransactionHistory() {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Transaction history for account %s:\n", currentUser.ID)
	for _, transaction := range currentUser.Transactions {
		fmt.Printf("ID: %d, Type: %s, Amount: %.2f, Date: %s, Details: %s\n",
			transaction.ID, transaction.Type, transaction.Amount, transaction.Date, transaction.Details)
	}
}

func saveAccounts() {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		log.Println("Failed to marshal accounts:", err)
		return
	}

	err = os.WriteFile(accountsFilePath, data, 0644)
	if err != nil {
		log.Println("Failed to save accounts:", err)
	}
}

func loadAccounts() {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(accountsFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("Failed to read accounts file:", err)
		}
		return
	}

	err = json.Unmarshal(data, &accounts)
	if err != nil {
		log.Println("Failed to unmarshal accounts:", err)
	}
}

func topUpBalance(scanner *bufio.Scanner) {
	fmt.Print("Enter amount to top up: ")
	if !scanner.Scan() {
		return
	}
	amountStr := scanner.Text()
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		fmt.Println("Invalid amount.")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	topUpRequests = append(topUpRequests, TopUpRequest{
		ID:        nextTopUpID,
		AccountID: currentUser.ID,
		Amount:    amount,
		Date:      time.Now().Format(time.RFC3339),
		Approved:  false,
	})
	nextTopUpID++

	fmt.Println("Top up request submitted.")
}

func handleTopUpRequests(scanner *bufio.Scanner) {
	mu.Lock()
	defer mu.Unlock()

	for i, request := range topUpRequests {
		if !request.Approved {
			fmt.Printf("Approve top up request %d for account %s of amount %.2f? (y/n): ", request.ID, request.AccountID, request.Amount)
			if !scanner.Scan() {
				return
			}
			if scanner.Text() == "y" {
				account := accounts[request.AccountID]
				account.Balance += request.Amount
				topUpRequests[i].Approved = true
				fmt.Println("Top up approved.")
			} else {
				fmt.Println("Top up rejected.")
			}
		}
	}
}
