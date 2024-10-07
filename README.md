# Go EDA (Event-Driven Architecture)

Go EDA adalah sebuah proyek yang menerapkan arsitektur berbasis peristiwa (Event-Driven Architecture) menggunakan bahasa pemrograman Go. Proyek ini mencakup operasi CRUD (Create, Read, Update, Delete) untuk produk dan integrasi dengan RabbitMQ untuk pengiriman pesan.

## Fitur

- CRUD Produk
- Integrasi dengan RabbitMQ untuk event messaging
- Penggunaan PostgreSQL sebagai database
- Penanganan permintaan HTTP menggunakan `net/http`
- Pengujian unit

## Teknologi yang Digunakan

- Go (Golang)
- RabbitMQ
- PostgreSQL
- GORM (ORM untuk Go)
- cURL untuk pengujian API

## Prerequisites

Sebelum menjalankan proyek ini, pastikan Anda telah menginstal:

- [Go](https://golang.org/doc/install) (versi 1.16 atau lebih baru)
- [PostgreSQL](https://www.postgresql.org/download/) dan telah berjalan
- [RabbitMQ](https://www.rabbitmq.com/download.html) dan telah berjalan

## Instalasi

1. **Clone repository**:
   ```bash
   git clone https://github.com/username/go-eda.git
   cd go-eda
