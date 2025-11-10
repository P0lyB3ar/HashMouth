# 🚀 HashMouth Quick Start

## Option 1: Simple Chat (Easiest!)

```powershell
.\start.ps1
# Choose: 1
```

Or manually:
```bash
cd examples/simple_backend
npm install
npm start
```

Open: **http://localhost:3000**

- Enter your name
- Chat with everyone or specific users
- Simple and fast!

## Option 2: Host .hmouth Domains

```powershell
.\start.ps1
# Choose: 2
```

Or manually:
```bash
go run cmd/hmouth_proxy.go
```

Open: **http://localhost:8888**

### Configure Browser Proxy:
- **Firefox**: Settings → Network → Manual proxy
  - HTTP Proxy: `localhost`
  - Port: `8888`

### Host Your Backend:
1. Start your backend (e.g., `npm start` on port 3000)
2. Open control panel: http://localhost:8888
3. Select "Backend Application"
4. Enter: `http://localhost:3000`
5. Domain: `chat` (becomes `chat.hmouth`)
6. Click "Host Site"
7. Visit: `http://chat.hmouth` (with proxy configured)

## That's It!

**Simple Chat** - For easy group chat
**HMouth Proxy** - For anonymous hosting

Choose what you need! 🎉
