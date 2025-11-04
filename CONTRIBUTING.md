# Pedoman Kontribusi untuk Topup Store Microservice

Terima kasih telah meluangkan waktu untuk berkontribusi pada proyek `topupstore-microservice`! Saya sangat menghargai bantuan Anda.

Pedoman berikut akan membantu Anda memahami cara berpartisipasi dalam proyek ini.

## Kode Etik (Code of Conduct)

Proyek ini dan semua orang yang berpartisipasi di dalamnya diatur oleh [Kode Etik](CODE_OF_CONDUCT.md) kami. Dengan berpartisipasi, Anda diharapkan untuk menjunjung tinggi kode etik ini.

## Bagaimana Saya Bisa Berkontribusi?

Saya menyambut baik semua jenis kontribusi, termasuk:

* **Melaporkan Bug:** Jika Anda menemukan *bug* atau perilaku yang tidak terduga.
* **Menyarankan Fitur:** Jika Anda memiliki ide untuk fungsionalitas baru.
* **Mengirimkan Pull Request:** Jika Anda ingin memperbaiki *bug*, menambah fitur, atau meningkatkan dokumentasi.

## Melaporkan Bug

Jika Anda menemukan *bug*, silakan buka **"Issue"** baru di repositori GitHub. Saat melaporkan *bug*, pastikan untuk menyertakan:

* **Langkah-langkah untuk mereproduksi:** Penjelasan yang jelas dan rinci tentang cara memicu *bug*.
* **Perilaku yang diharapkan:** Apa yang seharusnya terjadi.
* **Perilaku aktual:** Apa yang sebenarnya terjadi (termasuk pesan kesalahan atau *log*).
* **Lingkungan Anda:** Versi Go, sistem operasi, dan detail relevan lainnya.

## Menyarankan Fitur

1.  Buka **"Issue"** baru untuk mendiskusikan ide Anda.
2.  Jelaskan masalah yang ingin Anda selesaikan dan bagaimana fitur baru Anda akan menyelesaikannya.
3.  Ini memberi kesempatan bagi kami untuk mendiskusikan fitur tersebut sebelum Anda mulai mengerjakannya.

## Mengirimkan Pull Request (PR)

Ini adalah alur kerja yang kami sarankan untuk mengirimkan kontribusi kode:

### 1. Persiapan Awal

1.  **Fork** repositori ini ke akun GitHub Anda.
2.  **Clone** *fork* Anda ke mesin lokal Anda:
    ```bash
    git clone [https://github.com/USERNAME_ANDA/topupstore-microservice.git](https://github.com/USERNAME_ANDA/topupstore-microservice.git)
    cd topupstore-microservice
    ```
3.  **Tambahkan Upstream:** Tambahkan repositori asli sebagai *remote* bernama `upstream`:
    ```bash
    git remote add upstream [https://github.com/akmmp241/topupstore-microservice.git](https://github.com/akmmp241/topupstore-microservice.git)
    ```
4.  **Unduh Dependensi:** Pastikan Anda memiliki dependensi modul Go yang terbaru.

### 2. Buat Branch Baru

Selalu buat *branch* baru untuk setiap pekerjaan Anda. Jangan pernah bekerja langsung di *branch* `main`.

1.  Sinkronkan *branch* `main` Anda dengan `upstream`:
    ```bash
    git fetch upstream
    git checkout main
    git rebase upstream/main
    ```
2.  Buat *branch* baru dengan nama yang deskriptif:
    ```bash
    # Untuk fitur baru
    git checkout -b feature/nama-fitur-anda
    
    # Untuk perbaikan bug
    git checkout -b fix/deskripsi-bug
    ```

### 3. Lakukan Perubahan dan Tes

1.  Tulis kode Anda.
2.  **Penting (Gaya Kode):** Pastikan kode Anda telah diformat dengan benar menggunakan alat standar Go.
    ```bash
    gofmt -w .
    # atau
    goimports -w .
    ```

### 4. Commit dan Push

1.  Buat *commit* dengan pesan yang jelas dan deskriptif. Kami sangat menyarankan untuk mengikuti [Conventional Commits](https://www.conventionalcommits.org/).
    * `feat:` untuk fungsionalitas baru.
    * `fix:` untuk perbaikan *bug*.
    * `docs:` untuk perubahan dokumentasi.
    
    Contoh:
    ```bash
    git commit -m "feat: add endpoint payment validation"
    ```
2.  *Push* *branch* Anda ke *fork* Anda di GitHub:
    ```bash
    git push origin feature/nama-fitur-anda
    ```

### 5. Buka Pull Request

1.  Buka repositori *fork* Anda di GitHub.
2.  Klik tombol "Compare & pull request".
3.  Pastikan *base repository* adalah `akmmp241/topupstore-microservice` di *branch* `main` dan *head repository* adalah *fork* Anda di *branch* fitur Anda.
4.  Berikan judul yang jelas dan deskripsi singkat tentang apa yang dilakukan oleh PR Anda. Jika PR Anda menutup *issue* tertentu, tautkan *issue* tersebut di deskripsi (misal: `Resolve #123`).
5.  Tunggu ulasan dari *maintainer* (Saya 😆).

Sekali lagi, terima kasih telah berkontribusi!
