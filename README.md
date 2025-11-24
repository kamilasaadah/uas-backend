# 🎓 UAS Backend — Go Fiber, PostgreSQL & MongoDB

Backend API untuk memenuhi tugas **UAS Backend**, dibangun dengan arsitektur modular dan clean architecture menggunakan:

🚀 Go Fiber  
🗄 PostgreSQL  
📦 MongoDB  
🔐 JWT Auth  
⚙️ Clean Architecture (Model → Repo → Service → Handler)

---

## 📁 Struktur Project
app/ # model, repository, service
config/ # app env & logger
database/ # postgres & mongo connection
middleware/ # jwt auth
route/ # route admin, mahasiswa, dosen
main.go
.env

## ⭐ Fitur Utama
- Login JWT  
- CRUD Mahasiswa  
- CRUD Dosen Wali  
- CRUD Users & Roles  
- Input pekerjaan alumni (MongoDB)  
- Soft Delete (`deleted` enum)  
- Logging aktivitas  

---

## 🛠 Teknologi
Go Fiber · PostgreSQL · MongoDB · Pgx · JWT-Go · Godotenv · Zap Logger

## ✍️ Author
**Kamilatus Saadah** · UAS Backend 2025  