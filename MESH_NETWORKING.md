# 🌐 Mesh Networking - مثل Tailscale اما برای SSH Tunnels!

## 🚀 معرفی Mesh Network

**Mesh Network** امکان اتصال چندین سرور به یکدیگر در یک شبکه هوشمند، خودکار و مقیاس‌پذیر را فراهم می‌کند. درست مثل Tailscale، اما مخصوص SSH Tunnels و پروتکل‌های مختلف!

### 🎯 **چرا Mesh Network؟**

- **🔄 Load Balancing**: توزیع خودکار ترافیک بین سرورها
- **💓 Health Monitoring**: نظارت مداوم بر سلامت سرورها  
- **🔁 Auto Failover**: انتقال خودکار به سرور سالم در صورت قطعی
- **🌍 Geo Distribution**: انتخاب بهترین سرور بر اساس موقعیت جغرافیایی
- **📊 Real-time Metrics**: آمار و عملکرد لحظه‌ای
- **🔐 Encrypted**: ارتباط رمزگذاری شده end-to-end

---

## 🏗️ معماری Mesh Network

```
      ┌─────────────────────────────────────────────────────────┐
      │                  🌐 Mesh Network                        │
      │                 Network: 10.99.0.0/24                  │
      └─────────────────────────────────────────────────────────┘
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
    ┌────▼────┐                ┌────▼────┐                ┌────▼────┐
    │ Node 1  │◄──────────────►│ Node 2  │◄──────────────►│ Node 3  │
    │Local    │                │Server-1 │                │Server-2 │
    │10.99.0.1│                │10.99.0.2│                │10.99.0.3│
    │         │                │   US    │                │   EU    │
    └─────────┘                └─────────┘                └─────────┘
         │                          │                          │
         ▼                          ▼                          ▼
    ┌─────────┐                ┌─────────┐                ┌─────────┐
    │Client   │                │SSH/V2Ray│                │WireGuard│
    │Apps     │                │Tunnels  │                │ Proxy   │
    └─────────┘                └─────────┘                └─────────┘
```

---

## 🚀 راه‌اندازی سریع (5 دقیقه)

### مرحله 1: ایجاد Mesh Network

```bash
tunnel mesh init 10.99.0.0/24
```

```
🌐 Initializing mesh network with CIDR: 10.99.0.0/24
✅ Mesh network initialized!
✅ Local node: local-node (10.99.0.1)
💡 Add servers with: tunnel mesh add <host> <user>
```

### مرحله 2: اضافه کردن سرورها

```bash
# سرور اول (آمریکا)
tunnel mesh add us-server.example.com root
# 🔐 Enter SSH password: ********
# ✅ Server added to mesh: us-server (10.99.0.2)

# سرور دوم (اروپا)  
tunnel mesh add eu-server.example.com ubuntu
# 🔐 Enter SSH password: ********
# ✅ Server added to mesh: eu-server (10.99.0.3)

# سرور سوم (آسیا)
tunnel mesh add asia-server.example.com admin
# 🔐 Enter SSH password: ********
# ✅ Server added to mesh: asia-server (10.99.0.4)
```

### مرحله 3: مشاهده وضعیت شبکه

```bash
tunnel mesh status
```

```
🌐 Mesh Network Status
═════════════════════
   📊 Total Nodes: 4
   ✅ Online Nodes: 4
   ❌ Offline Nodes: 0
   🌍 Network: 10.99.0.0/24
   ⚖️ Load Balancing: latency

Nodes:
   🟢 local-node (10.99.0.1) - online
   🟢 us-server (10.99.0.2) - online - 25ms
   🟢 eu-server (10.99.0.3) - online - 45ms  
   🟢 asia-server (10.99.0.4) - online - 80ms
```

### مرحله 4: اتصال به شبکه

```bash
tunnel mesh connect
```

```
🔗 Connecting to best mesh node...
✅ Connected to us-server (10.99.0.2)
🌐 SOCKS5 proxy: 127.0.0.1:8080
🌐 HTTP proxy: 127.0.0.1:8081
🌐 Mesh Dashboard: http://localhost:8888
```

---

## ⚡ ویژگی‌های پیشرفته

### 🎯 **Auto Load Balancing**

```bash
# بر اساس Latency (پیش‌فرض)
tunnel mesh connect

# بر اساس Load (بار سیستم)
tunnel mesh connect --method load

# بر اساس Region (منطقه)
tunnel mesh connect --region us

# Random (تصادفی)
tunnel mesh connect --method random
```

### 💓 **Health Monitoring خودکار**

```yaml
# کانفیگ خودکار health checking
health_monitoring:
  check_interval: 30s          # چک هر 30 ثانیه
  timeout: 10s                 # timeout برای هر چک
  retry_attempts: 3            # تعداد تلاش مجدد
  failover_threshold: 3        # حد آستانه برای failover
  auto_recovery: true          # بازگشت خودکار پس از بهبودی
```

**نمونه Log:**
```
2024-07-22 22:45:12 INFO  Node us-server (10.99.0.2) - Healthy (latency: 25ms)
2024-07-22 22:45:15 WARN  Node eu-server (10.99.0.3) - High latency (120ms)
2024-07-22 22:45:42 ERROR Node asia-server (10.99.0.4) - Connection failed
2024-07-22 22:45:43 INFO  Failover: Switching from asia-server to us-server
2024-07-22 22:46:15 INFO  Node asia-server (10.99.0.4) - Recovered, back online
```

### 🔁 **Smart Failover**

```bash
# مثال: سرور اصلی قطع شد
Current: Connected to eu-server (10.99.0.3)
⚠️  eu-server connection lost
🔄 Automatic failover in progress...
✅ Switched to us-server (10.99.0.2) 
🌐 New proxy: 127.0.0.1:8080
```

### 🌍 **Geographic Routing**

```bash
# اتصال به سرورهای منطقه‌ای
tunnel mesh connect --region us      # سرورهای آمریکا
tunnel mesh connect --region eu      # سرورهای اروپا
tunnel mesh connect --region asia    # سرورهای آسیا

# لیست سرورهای هر منطقه
tunnel mesh list --region us
# 🟢 us-west-1.example.com (10.99.0.5) - 20ms
# 🟢 us-east-1.example.com (10.99.0.6) - 35ms
```

---

## 🔧 سناریوهای واقعی

### 🏢 **سناریو 1: شرکت چند ملیتی**

```bash
# دفاتر شرکت در کشورهای مختلف
tunnel mesh init 192.168.100.0/24

# دفتر تهران
tunnel mesh add office-tehran.company.com admin --region iran

# دفتر دبی  
tunnel mesh add office-dubai.company.com admin --region uae

# دفتر استانبول
tunnel mesh add office-istanbul.company.com admin --region turkey

# کارمندان خودکار به نزدیک‌ترین دفتر وصل می‌شوند
tunnel mesh connect
```

### 🎮 **سناریو 2: Gaming Network**

```bash
# سرورهای گیمینگ با کمترین ping
tunnel mesh init 10.gaming.0.0/24

# سرور اروپا (کم ping برای CS:GO)
tunnel mesh add eu-gaming.provider.com gamer --region eu --tags gaming,csgo

# سرور آمریکا (کم ping برای Valorant)  
tunnel mesh add us-gaming.provider.com gamer --region us --tags gaming,valorant

# سرور آسیا (کم ping برای PUBG)
tunnel mesh add asia-gaming.provider.com gamer --region asia --tags gaming,pubg

# اتصال خودکار به کمترین ping
tunnel mesh connect --method latency
```

### 🎬 **سناریو 3: Streaming & Content**

```bash
# سرورهای streaming
tunnel mesh init 10.stream.0.0/24

# Netflix US
tunnel mesh add netflix-us.provider.com user --tags streaming,netflix

# BBC iPlayer UK
tunnel mesh add uk-streaming.provider.com user --tags streaming,bbc

# Japanese content
tunnel mesh add jp-content.provider.com user --tags streaming,anime

# انتخاب بر اساس محتوا
tunnel mesh connect --tags netflix    # برای Netflix
tunnel mesh connect --tags bbc        # برای BBC
```

### 📈 **سناریو 4: High Availability**

```bash
# شبکه با قابلیت اعتماد بالا
tunnel mesh init 172.16.0.0/24

# سرور اصلی
tunnel mesh add primary.service.com root --priority 1

# سرورهای backup
tunnel mesh add backup1.service.com root --priority 2  
tunnel mesh add backup2.service.com root --priority 3

# Load balancer خودکار
tunnel mesh connect --method priority
```

---

## 📊 مانیتورینگ پیشرفته

### 🖥️ **Dashboard تحت وب**

```bash
# شروع dashboard
tunnel mesh dashboard --port 8888
```

دسترسی به: `http://localhost:8888`

**ویژگی‌های Dashboard:**
- 📊 Real-time metrics
- 🌍 نقشه جغرافیایی nodes
- 📈 نمودار latency و throughput  
- ⚠️ هشدارها و alerts
- 📋 لاگ‌های سیستم
- ⚙️ تنظیمات mesh

### 📈 **Metrics و آمار**

```bash
# آمار کلی
tunnel mesh metrics

# آمار هر node
tunnel mesh metrics --node us-server

# آمار بر اساس زمان
tunnel mesh metrics --since 1h
tunnel mesh metrics --since 24h
```

**نمونه خروجی:**
```
📊 Mesh Network Metrics (Last 1 hour)
════════════════════════════════════

🌐 Network Overview:
   Total Nodes: 4
   Online: 4 (100%)
   Total Connections: 1,245
   Data Transferred: 15.2 GB
   Average Latency: 35ms

📡 Node Performance:
   us-server    (10.99.0.2): 450 conn, 8.2GB, 25ms avg
   eu-server    (10.99.0.3): 380 conn, 4.1GB, 45ms avg  
   asia-server  (10.99.0.4): 280 conn, 2.1GB, 80ms avg
   backup       (10.99.0.5): 135 conn, 0.8GB, 120ms avg

🔄 Load Balancing:
   us-server:    36.1% traffic
   eu-server:    30.5% traffic
   asia-server:  22.6% traffic
   backup:       10.8% traffic

⚡ Events (Last 1h):
   22:15:23 - Failover: asia-server → us-server (60 connections)
   22:18:45 - Node asia-server recovered
   22:32:10 - High traffic detected on eu-server
   22:45:22 - Load balancer redistributed 150 connections
```

---

## 🔐 امنیت در Mesh Network

### 🛡️ **Multi-layer Security**

```yaml
mesh_security:
  # رمزگذاری کانال‌های ارتباطی
  inter_node_encryption: AES-256-GCM
  
  # احراز هویت nodes
  node_authentication: 
    method: certificate
    ca_cert: /etc/mesh/ca.pem
    
  # کنترل دسترسی
  access_control:
    allow_regions: [us, eu, asia]
    deny_ips: [192.168.1.100]
    require_tags: [trusted]
    
  # مانیتورینگ امنیتی  
  security_monitoring:
    detect_anomalies: true
    alert_failed_auth: true
    log_all_connections: true
```

### 🔑 **Certificate Management**

```bash
# ایجاد CA برای mesh
tunnel mesh create-ca --name "MyMesh CA"

# ایجاد certificate برای node
tunnel mesh create-cert --node us-server --ca MyMesh

# اعتبارسنجی certificates
tunnel mesh verify-certs

# تمدید certificate
tunnel mesh renew-cert --node us-server
```

---

## 🚀 API و Automation

### 🌐 **RESTful API**

```bash
# Health check
curl http://localhost:8888/api/v1/mesh/health

# وضعیت کلی mesh
curl http://localhost:8888/api/v1/mesh/status

# لیست nodes
curl http://localhost:8888/api/v1/mesh/nodes

# اضافه کردن node
curl -X POST http://localhost:8888/api/v1/mesh/nodes \
  -d '{"host": "new-server.com", "user": "root", "region": "us"}'

# اتصال به node خاص
curl -X POST http://localhost:8888/api/v1/mesh/connect/node-id
```

### 🤖 **Webhook Integration**

```yaml
webhooks:
  - name: slack_alerts
    url: https://hooks.slack.com/services/xxx
    events: [node_down, failover, high_latency]
    
  - name: monitoring_system  
    url: https://monitoring.company.com/webhook
    events: [metrics, status_change]
    headers:
      Authorization: "Bearer your-token"
```

### 📜 **Infrastructure as Code**

```yaml
# mesh-infrastructure.yaml
apiVersion: tunnel.mesh/v1
kind: MeshNetwork
metadata:
  name: production-mesh
spec:
  network_cidr: "10.99.0.0/24"
  
  nodes:
    - name: us-west
      host: us-west.company.com
      region: us
      priority: 1
      tags: [production, web]
      
    - name: eu-central
      host: eu-central.company.com  
      region: eu
      priority: 2
      tags: [production, api]
      
  load_balancing:
    method: latency
    health_check_interval: 30s
    
  security:
    encryption: true
    authentication: certificate
```

```bash
# اعمال کانفیگ
tunnel mesh apply -f mesh-infrastructure.yaml
```

---

## 📈 بهینه‌سازی Performance

### ⚡ **Tuning Tips**

```bash
# بهینه‌سازی شبکه
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_congestion_control = bbr' >> /etc/sysctl.conf
sysctl -p

# بهینه‌سازی mesh
tunnel mesh optimize --profile high-throughput
tunnel mesh optimize --profile low-latency
tunnel mesh optimize --profile battery-saving
```

### 📊 **Benchmarking**

```bash
# تست کارکرد mesh
tunnel mesh benchmark --duration 5m

# تست latency بین nodes
tunnel mesh ping --all-nodes

# تست throughput
tunnel mesh speedtest --node us-server
```

---

## 🎯 مقایسه با رقبا

### 🆚 **Mesh Network vs Tailscale**

| ویژگی | SSH Tunnel Manager Mesh | Tailscale |
|--------|------------------------|-----------|
| **پروتکل‌ها** | SSH, V2Ray, WireGuard, Trojan, Hysteria | فقط WireGuard |
| **Load Balancing** | ✅ چندین روش | ❌ خیر |
| **Failover** | ✅ خودکار | ❌ دستی |
| **Geo Routing** | ✅ با tags | ❌ محدود |
| **Custom Protocols** | ✅ قابل توسعه | ❌ محدود |
| **Cost** | ✅ رایگان | 💰 پولی |
| **Privacy** | ✅ کاملاً خصوصی | ⚠️ محدودیت |

### 🆚 **مزایای کلیدی**

✅ **Protocol Flexibility**: پشتیبانی از 9+ پروتکل  
✅ **Smart Load Balancing**: توزیع هوشمند ترافیک  
✅ **Auto Failover**: تضمین uptime بالا  
✅ **Cost Effective**: هیچ هزینه اضافی  
✅ **Full Control**: کنترل کامل بر infrastructure  
✅ **Enterprise Ready**: قابلیت‌های سازمانی  
✅ **Open Source**: کد باز و قابل توسعه  

---

## 🔮 آینده Mesh Network

### 🚀 **ویژگی‌های در دست توسعه**

- **🤖 AI-Powered Routing**: مسیریابی هوشمند با ML
- **🌊 Traffic Shaping**: کنترل هوشمند bandwidth  
- **🔄 Dynamic Scaling**: افزایش خودکار nodes
- **📱 Mobile Apps**: اپ موبایل برای مدیریت
- **🌍 CDN Integration**: ادغام با CDN providers
- **⚡ Edge Computing**: پردازش در edge nodes

### 💡 **Roadmap**

**Q1 2024:**
- GraphQL API
- Advanced Metrics
- Multi-tenant Support

**Q2 2024:**  
- AI Load Balancing
- Mobile Management App
- Kubernetes Integration

**Q3 2024:**
- Edge Computing
- CDN Integration  
- Advanced Security

---

## 🎉 نتیجه‌گیری

**Mesh Network** در SSH Tunnel Manager ترکیبی از سادگی Tailscale، قدرت enterprise solutions، و انعطاف‌پذیری کامل ارائه می‌دهد.

### 🏆 **چرا انتخاب کنید:**

- **🕐 5 دقیقه راه‌اندازی**: سریع‌ترین setup ممکن
- **🔄 Zero Maintenance**: مدیریت خودکار کامل  
- **💰 Cost Efficient**: هیچ هزینه ماهانه
- **🚀 Enterprise Grade**: آماده برای سازمان‌ها
- **🌍 Global Scale**: مقیاس‌پذیری جهانی

### 📞 **شروع کنید:**

```bash
# 1. نصب
go build -o tunnel ./cmd/main.go

# 2. ایجاد mesh  
tunnel mesh init

# 3. اضافه کردن سرور
tunnel mesh add your-server.com root

# 4. اتصال
tunnel mesh connect

# 🎉 تبریک! شبکه mesh شما آماده است!
```

**پروژه شما از یک SSH tunnel ساده به یک Mesh Network Platform تبدیل شده است!** 🌐✨ 