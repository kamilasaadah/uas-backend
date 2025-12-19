// middleware/blocklist_test_helper.go
package middleware

// Fungsi KHUSUS UNTUK TESTING saja
// Membersihkan seluruh JWT blocklist
func ClearBlocklistForTest() {
	mutex.Lock()
	defer mutex.Unlock()
	jwtBlocklist = make(map[string]blockedToken)
}

// Opsional: fungsi untuk cek langsung isi blocklist tanpa auto-clean
func IsTokenInBlocklistForTest(token string) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	_, exists := jwtBlocklist[token]
	return exists
}
