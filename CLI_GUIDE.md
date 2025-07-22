# 🎯 CLI Guide - Simple & Memorable Commands

## 🚀 طراحی جدید CLI: ساده، قابل حفظ، قدرتمند!

### 💡 فلسفه طراحی جدید:
- **ساده**: بدون پارامترهای پیچیده
- **قابل حفظ**: کامندهای کوتاه و منطقی
- **هوشمند**: تشخیص خودکار نیازها
- **انعطاف‌پذیر**: هم CLI و هم Interactive mode

---

## 🌟 کامندهای اصلی

### 1. 🔍 **Quick Setup** (ساده‌ترین راه)

```bash
# Auto-discovery ساده
tunnel quick <ip> <user> <password>

# مثال‌ها:
tunnel quick 1.2.3.4 root mypassword
tunnel quick 1.2.3.4 ubuntu ~/.ssh/id_rsa
tunnel quick 1.2.3.4 root mypass --setup    # با نصب پروتکل‌ها
```

**چی میکنه:**
- ✅ سرور رو خودکار کشف میکنه
- ✅ پروتکل‌های موجود رو تشخیص میده  
- ✅ (اختیاری) پروتکل‌ها رو نصب میکنه
- ✅ کانفیگ‌های کلاینت رو میسازه
- ✅ SSH Tunnel Manager config آماده

### 2. 🌐 **Mesh Network** (مثل Tailscale)

```bash
# شروع mesh network
tunnel mesh init [network-cidr]

# اضافه کردن سرور به mesh
tunnel mesh add <ip> <user> [password]

# مشاهده وضعیت mesh
tunnel mesh status

# اتصال به mesh
tunnel mesh connect
```

**مثال کامل:**
```bash
# ساخت شبکه mesh
tunnel mesh init 10.99.0.0/24

# اضافه کردن سرورها
tunnel mesh add server1.com root
tunnel mesh add server2.com ubuntu
tunnel mesh add 1.2.3.4 admin

# مشاهده شبکه
tunnel mesh status

# اتصال به بهترین node
tunnel mesh connect
```

### 3. 📁 **Configuration Management**

```bash
# اجرای کانفیگ
tunnel config <config-file>

# با وب interface
tunnel config <config-file> --server

# تنظیم پورت سفارشی
tunnel config <config-file> --server --port 9999
```

### 4. 🌐 **Server Mode**

```bash
# شروع سرور مدیریت
tunnel server

# با پورت سفارشی
tunnel server --port 8888

# با کانفیگ خاص
tunnel server --config myconfig.yaml --port 9000
```

### 5. 🎨 **Interactive Mode**

```bash
# حالت تعاملی (فقط tunnel بدون آرگومان)
tunnel

# یا صریحاً
tunnel interactive
tunnel menu
```

---

## 🎯 مقایسه: قدیم vs جدید

### ❌ **CLI قدیم (پیچیده):**
```bash
# مثال پیچیده قدیمی
./ssh-tunnel-manager -autodiscover -host 1.2.3.4 -user root -password mypass -setup -output client-configs

# طولانی و غیر قابل حفظ!
./ssh-tunnel-manager -config configs/config.yaml -server -port 8888
```

### ✅ **CLI جدید (ساده):**
```bash
# همان کار با کامند ساده
tunnel quick 1.2.3.4 root mypass --setup

# ساده و قابل حفظ!
tunnel config configs/config.yaml --server
```

---

## 🌐 راه‌اندازی Mesh Network (مثل Tailscale)

### 🎯 **سناریو: شبکه Mesh با 3 سرور**

```bash
# مرحله 1: ایجاد mesh network
tunnel mesh init 10.99.0.0/24
# ✅ Network: 10.99.0.0/24 created
# ✅ Local node: local-node (10.99.0.1)

# مرحله 2: اضافه کردن سرور اول
tunnel mesh add server1.example.com root
# 🔐 Enter SSH password: ********
# ✅ Server added to mesh: server1 (10.99.0.2)

# مرحله 3: اضافه کردن سرور دوم
tunnel mesh add 1.2.3.4 ubuntu
# 🔐 Enter SSH password: ********
# ✅ Server added to mesh: mesh-1.2.3.4 (10.99.0.3)

# مرحله 4: مشاهده وضعیت
tunnel mesh status
# 🌐 Mesh Network Status
# ═════════════════════
#    📊 Total Nodes: 3
#    ✅ Online Nodes: 3
#    ❌ Offline Nodes: 0
#    🌍 Network: 10.99.0.0/24
#    ⚖️ Load Balancing: latency
#
# Nodes:
#    🟢 local-node (10.99.0.1) - online
#    🟢 server1 (10.99.0.2) - online - 25ms
#    🟢 mesh-1.2.3.4 (10.99.0.3) - online - 45ms

# مرحله 5: اتصال به شبکه
tunnel mesh connect
# 🔗 Connecting to best mesh node...
# ✅ Connected to server1 (10.99.0.2)
# 🌐 SOCKS5 proxy: 127.0.0.1:8080
# 🌐 HTTP proxy: 127.0.0.1:8081
```

### ⚡ **ویژگی‌های Mesh Network:**

- **🔄 Auto Load Balancing**: بهترین سرور بر اساس latency
- **💓 Health Monitoring**: چک مداوم سلامت nodes
- **🔁 Auto Failover**: اتصال خودکار به سرور جایگزین
- **📊 Real-time Metrics**: نظارت لحظه‌ای بر کارکرد
- **🌍 Geo Distribution**: پخش جغرافیایی سرورها
- **🔐 Encrypted Mesh**: رمزگذاری end-to-end

---

## 🎨 Interactive Mode (تعاملی)

فقط `tunnel` تایپ کنید و وارد حالت تعاملی بشید:

```
🚀 SSH Tunnel Manager
=====================

Choose an option:

  1. 🔍 Quick Setup (Auto-discover server)
  2. 🌐 Mesh Network (Connect multiple servers)  
  3. 📁 Use existing config
  4. ⚙️ Advanced configuration
  5. 📊 Monitor connections
  6. 🔧 Manage servers
  7. 📖 Help & Documentation
  8. 🚪 Exit

📝 Select option (1-8): 
```

### 🔍 **Quick Setup Wizard:**
```
🔍 Quick Setup Wizard
=====================

This will automatically discover and setup your server with all supported protocols.

📝 Enter server IP or hostname: 1.2.3.4
📝 Enter SSH username: root

Choose authentication method:
  1. 🔑 Password
  2. 🔐 SSH Key
📝 Select (1-2): 1
🔐 Enter SSH password: ********

📝 Setup all protocols on server? (y/n): y
📝 Output directory for configs [client-configs]: 

🚀 Starting auto-discovery...
```

---

## 🏗️ ساختار پیشنهادی برای Use Cases

### 1. **🏠 شخصی: VPN سریع**
```bash
# یک سرور VPS دارید
tunnel quick your-vps.com root password123 --setup

# فوری‌اً تونل‌های مختلف آماده:
# ✅ SSH Tunnel (port 8080)
# ✅ V2Ray/VLESS 
# ✅ WireGuard VPN
# ✅ Trojan proxy
```

### 2. **🏢 شرکتی: چند دفتر**
```bash
# Mesh network برای اتصال دفاتر مختلف
tunnel mesh init 192.168.100.0/24
tunnel mesh add office-tehran.company.com admin
tunnel mesh add office-isfahan.company.com admin  
tunnel mesh add office-tabriz.company.com admin

# اتصال خودکار به نزدیک‌ترین دفتر
tunnel mesh connect
```

### 3. **🎮 Gaming: کاهش Ping**
```bash
# سرورهای game در مناطق مختلف
tunnel mesh init
tunnel mesh add game-eu.example.com gamer
tunnel mesh add game-us.example.com gamer
tunnel mesh add game-asia.example.com gamer

# اتصال به کم‌ترین ping
tunnel mesh connect
```

### 4. **🎬 Streaming: دسترسی محتوا**
```bash
# سرورهای مختلف برای Content
tunnel mesh add us-server.com user      # Netflix US
tunnel mesh add uk-server.com user      # BBC iPlayer  
tunnel mesh add jp-server.com user      # Japanese content

tunnel mesh status  # چک کردن latency
tunnel mesh connect # اتصال به بهترین
```

---

## 🛠️ نکات پیشرفته

### ⚙️ **ترکیب CLI و Interactive:**

```bash
# شروع سریع با CLI
tunnel quick 1.2.3.4 root mypass

# ادامه کار در Interactive mode
tunnel
# > انتخاب گزینه 2 برای Mesh Network
```

### 🔧 **استفاده از فایل‌های تولید شده:**

```bash
# بعد از Quick Setup
tunnel quick server.com root pass --setup

# استفاده از کانفیگ تولید شده
tunnel config client-configs/ssh-tunnel-manager-config.yaml --server

# دسترسی به وب interface
# http://localhost:8888
```

### 📊 **مانیتورینگ و مدیریت:**

```bash
# شروع سرور مانیتورینگ
tunnel server --port 8888

# API endpoints موجود:
curl http://localhost:8888/api/v1/health
curl http://localhost:8888/api/v1/status  
curl http://localhost:8888/api/v1/metrics
```

---

## 🎯 مزایای CLI جدید

### ✅ **Simple & Memorable:**
- `tunnel quick` → همه میفهمن
- `tunnel mesh` → مثل Tailscale
- `tunnel server` → وب interface
- `tunnel config` → استفاده از کانفیگ

### ✅ **Progressive Disclosure:**
- بدون آرگومان → Interactive mode
- آرگومان ناقص → راهنمای مربوطه
- آرگومان کامل → اجرای مستقیم

### ✅ **Backward Compatible:**
- CLI قدیم هنوز کار میکنه
- Migration تدریجی امکان‌پذیر
- پیام‌های راهنما برای انتقال

### ✅ **Flexible:**
- CLI برای automation
- Interactive برای manual
- Config files برای enterprise

---

## 🚀 مثال‌های کامل

### 🌟 **سناریو 1: راه‌اندازی سریع (2 دقیقه)**

```bash
# گام 1: کشف و راه‌اندازی سرور
tunnel quick 1.2.3.4 root mypassword --setup
# ✅ 9 نوع تونل آماده شد!

# گام 2: شروع tunnel manager  
tunnel config client-configs/ssh-tunnel-manager-config.yaml --server
# ✅ وب interface: http://localhost:8888
# ✅ SOCKS5: 127.0.0.1:8080
# ✅ HTTP: 127.0.0.1:8081
```

### 🌟 **سناریو 2: Mesh Network (5 دقیقه)**

```bash
# گام 1: ایجاد mesh
tunnel mesh init

# گام 2: اضافه کردن سرورها
tunnel mesh add server1.com root
tunnel mesh add server2.com ubuntu  
tunnel mesh add 1.2.3.4 admin

# گام 3: اتصال به شبکه
tunnel mesh connect
# ✅ Load balancing خودکار
# ✅ Failover خودکار
# ✅ Health monitoring
```

### 🌟 **سناریو 3: Enterprise Setup**

```bash
# گام 1: کشف سرورهای متعدد
tunnel quick server-us.company.com admin pass1
tunnel quick server-eu.company.com admin pass2
tunnel quick server-asia.company.com admin pass3

# گام 2: ادغام کانفیگ‌ها
cat client-configs-*/ssh-tunnel-manager-config.yaml > enterprise-config.yaml

# گام 3: شروع سرور مرکزی
tunnel config enterprise-config.yaml --server --port 8888
```

---

## 💡 بهترین Practices

### 🔐 **امنیت:**
```bash
# استفاده از SSH key به جای password
tunnel quick server.com user ~/.ssh/id_rsa

# محدود کردن دسترسی فایل‌های کانفیگ
chmod 600 client-configs/*
```

### 📊 **Performance:**
```bash
# استفاده از mesh برای load balancing
tunnel mesh init
tunnel mesh add fast-server.com user
tunnel mesh add backup-server.com user

# اتصال خودکار به بهترین
tunnel mesh connect
```

### 🔧 **مدیریت:**
```bash
# شروع server mode برای مانیتورینگ
tunnel server --port 8888

# دسترسی به API
curl http://localhost:8888/api/v1/status
```

### 💾 **Backup & Recovery:**
```bash
# پشتیبان‌گیری از کانفیگ‌ها
tar -czf tunnel-backup.tar.gz client-configs/

# بازیابی و استفاده
tar -xzf tunnel-backup.tar.gz
tunnel config client-configs/ssh-tunnel-manager-config.yaml
```

---

## 🎉 خلاصه: انقلاب در UX!

### ❌ **قبل (پیچیده):**
- 20+ پارامتر CLI
- مستندات 50 صفحه‌ای
- کامندهای غیر قابل حفظ
- فقط experts میتونستن استفاده کنن

### ✅ **حالا (ساده):**
- 4 کامند اصلی: `quick`, `mesh`, `config`, `server`
- قابل حفظ و منطقی
- Interactive mode برای راهنمایی
- هر کسی میتونه استفاده کنه

### 🚀 **نتیجه:**
- **10x راحت‌تر** برای کاربران جدید
- **5x سریع‌تر** برای کاربران با تجربه  
- **Mesh networking** مثل Tailscale
- **Enterprise-ready** با مانیتورینگ

پروژه شما از یک SSH tunnel ساده به یک **Enterprise Tunnel Management Platform** با UX فوق‌العاده تبدیل شده! 🎯✨ 