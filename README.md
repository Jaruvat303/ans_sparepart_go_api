# ANS Sparepart Go API

ระบบ RESTful API สำหรับการจัดการอะไหล่ (Spare Part Management System) พัฒนาด้วยภาษา Go โดยเน้นประสิทธิภาพ (High Performance) และโครงสร้างที่ดูแลรักษาง่าย (Maintainability) ตามหลัก Clean Architecture

![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)
![Fiber Framework](https://img.shields.io/badge/Fiber-v2-black?style=flat&logo=gofiber)

## 🛠 Tech Stack & Libraries

โปรเจกต์นี้เลือกใช้เครื่องมือและไลบรารีที่มีประสิทธิภาพสูงและเป็นมาตรฐานในอุตสาหกรรม:

* **Core Framework:** [Go Fiber](https://gofiber.io/) - Web Framework ที่ทำงานได้รวดเร็ว (Express-inspired)
* **Database:** [PostgreSQL](https://www.postgresql.org/) - ฐานข้อมูล Relational Database หลักของระบบ
* **ORM:** [GORM](https://gorm.io/) - ORM Library สำหรับจัดการฐานข้อมูลและ Migration
* **Authentication:** [JWT-Go](https://github.com/golang-jwt/jwt) - การยืนยันตัวตนแบบ Stateless ด้วย JSON Web Tokens
* **Logging:** [Uber Zap](https://github.com/uber-go/zap) - Logger ที่มีความเร็วสูง (Blazing fast structured logging)
* **Documentation:** [Go-Swagger](https://github.com/go-swagger/go-swagger) - สำหรับสร้างเอกสาร API อัตโนมัติ
* **Testing:** [Testify](https://github.com/stretchr/testify) - Toolkit สำหรับการเขียน Unit Test และ Assertion

## 📂 โครงสร้างโปรเจกต์ (Project Structure)



โปรเจกต์นี้อ้างอิงรูปแบบ **Standard Go Project Layout**:

| Path | รายละเอียด |
|------|------------|
| `cmd/api` | จุดเริ่มต้นของแอปพลิเคชัน (Main entry point) |
| `config` | ไฟล์การตั้งค่าระบบและ Environment Variables |
| `internal` | Business Logic หลัก (Services, Handlers, Repositories) |
| `migrations` | ไฟล์ Database Migration |
| `pkg` | Library กลางที่สามารถแชร์ได้ (เช่น Utils, Helper functions) |
| `docs` | เอกสาร Swagger API |
| `docker-compose.yml` | การตั้งค่า Environment สำหรับ Database |
| `Makefile` | ชุดคำสั่งลัด (Build, Run, Test, Docs) |

## 🚀 การเริ่มต้นใช้งาน (Getting Started)

### สิ่งที่ต้องมี (Prerequisites)
* Go (version 1.18+)
* Docker & Docker Compose (สำหรับ PostgreSQL)

### การติดตั้ง (Installation)

1.  **Clone Repository**
    ```bash
    git clone [https://github.com/Jaruvat303/ans_sparepart_go_api.git](https://github.com/Jaruvat303/ans_sparepart_go_api.git)
    cd ans_sparepart_go_api
    ```

2.  **ติดตั้ง Dependencies**
    ```bash
    go mod tidy
    ```

3.  **สร้างไฟล์ Environment (.env)**
    คัดลอกไฟล์ตัวอย่าง (ถ้ามี) หรือสร้างใหม่เพื่อตั้งค่า DB Connection:
    ```env
    DB_HOST=localhost
    DB_USER=postgres
    DB_PASSWORD=yourpassword
    DB_NAME=ans_sparepart
    DB_PORT=5432
    JWT_SECRET=your_secret_key
    ```

### การรันโปรแกรม (Running the Application)

1.  **Start Database**
    ```bash
    docker-compose up -d
    ```

2.  **Run Application**
    สามารถรันผ่าน Makefile (ถ้ามี) หรือ Go command:
    ```bash
    make run
    # หรือ
    go run cmd/api/main.go
    ```

## 🧪 การทดสอบ (Testing)

โปรเจกต์นี้ใช้ **Testify** ในการทำ Unit Test สามารถรันคำสั่ง:
```bash
go test ./...
