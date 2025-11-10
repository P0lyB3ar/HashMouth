# 🔒 HashMouth - Anonymous P2P Communication

HashMouth is a decentralized, anonymous communication platform with DHT peer discovery, .hmouth domains (like Tor's .onion), and multiple anonymity layers.

## 🚀 Quick Start

### Simple Chat (Easiest!)
```bash
# Start backend
cd examples/simple_backend
npm install
npm start

# Open browser: http://localhost:3000
# Enter your name and start chatting!
```

### Host on .hmouth Domain
```bash
# Start proxy
go run cmd/hmouth_proxy.go

# Open control panel: http://localhost:8888
# Host your backend as chat.hmouth
# Configure browser proxy: localhost:8888
# Visit: http://chat.hmouth
```

## 🌟 Features

### 1. **DHT Peer Discovery**
- Automatic peer discovery using BitTorrent DHT
- No central server needed
- Connects to public DHT bootstrap nodes
- Works like torrent peer discovery

### 2. **.hmouth Domains** (Like Tor .onion)
- Host anonymous websites on .hmouth domains
- Custom domain names (unlike Tor's random)
- DHT-based discovery (no directories)
- Backend hosting support (Node.js, Python, Go, etc.)

### 3. **Ultra Anonymous Chat**
- Multi-hop routing (3-6 hops, more than Tor)
- Cover traffic (fake messages)
- ID rotation (every 15 minutes)
- Timing obfuscation (random delays)
- Better anonymity than Tor!

### 4. **Simple Chat**
- Beautiful web UI
- Send to everyone or specific users
- Real-time messaging
- Easy to use

### 5. **End-to-End Encryption**
- ChaCha20-Poly1305 encryption
- Onion routing
- Encrypted P2P connections

## 📋 What You Can Do

### 1. Simple Chat
```bash
cd examples/simple_backend
npm install
npm start
# Open: http://localhost:3000
```
- Enter your name
- Chat with everyone or specific users
- Simple and fast

### 2. Host .hmouth Websites
```bash
go run cmd/hmouth_proxy.go
```
- Host static sites or backends
- Access via .hmouth domains
- Anonymous hosting
- Like Tor hidden services

### 3. Ultra Anonymous Chat
```bash
go run cmd/ultra_anonymous.go
```
- Maximum anonymity
- Multi-hop routing
- Cover traffic
- ID rotation
- Better than Tor

### 4. DHT Chat
```bash
go run cmd/dht_chat.go
```
- Automatic peer discovery
- No manual setup
- DHT-based

## 🏗️ Architecture

### Core Components

**DHT Network** (`network/dht.go`)
- BitTorrent DHT implementation
- Connects to public bootstrap nodes
- Automatic peer discovery
- Announces presence

**P2P Network** (`network/node.go`)
- TCP-based P2P connections
- Encrypted communication
- Peer management

**Relay Network** (`network/relay.go`)
- Multi-hop message routing
- Anonymous communication
- Path building

**Crypto** (`crypto/`)
- ChaCha20-Poly1305 encryption
- Onion packet creation
- Key management

### Applications

**HMouth Proxy** (`cmd/hmouth_proxy.go`)
- Local HTTP proxy
- Resolves .hmouth domains
- Hosts websites
- Backend proxying

**Ultra Anonymous Chat** (`cmd/ultra_anonymous.go`)
- Maximum anonymity features
- Cover traffic generation
- ID rotation
- Timing obfuscation

**DHT Chat** (`cmd/dht_chat.go`)
- DHT peer discovery
- Automatic connections
- Simple chat interface

**Simple Chat Backend** (`examples/simple_backend/`)
- Node.js Express server
- REST API
- Serves web UI
- Message storage

## 🎯 Use Cases

### For Friends/Family
- **Simple Chat**: Easy group chat
- No setup required
- Just enter name and chat

### For Privacy
- **Ultra Anonymous Chat**: Maximum anonymity
- Hidden IP addresses
- Encrypted messages

### For Hosting
- **.hmouth Domains**: Anonymous websites
- Host backends (APIs, apps)
- No DNS needed

### For Public Networks
- **DHT Chat**: Automatic discovery
- No central server
- Decentralized

## 📚 Documentation

- **SIMPLE_BACKEND_GUIDE.md** - Simple chat setup
- **HMOUTH_DOMAINS_GUIDE.md** - .hmouth domain hosting
- **BACKEND_HOSTING_GUIDE.md** - Backend hosting guide
- **ULTRA_ANONYMOUS_GUIDE.md** - Ultra anonymous features
- **DHT_GUIDE.md** - DHT peer discovery

## 🔧 Requirements

**Go Applications:**
- Go 1.23 or higher
- golang.org/x/crypto

**Node.js Backend:**
- Node.js 14+
- npm packages: express, cors

## 🚨 Security Notice

### Privacy Levels:

**Simple Chat:**
- ⚠️ IP visible to backend
- ✅ Encrypted messages
- ✅ No central server

**.hmouth Domains:**
- ✅ Server IP hidden
- ✅ Anonymous hosting
- ⚠️ Visitor IP visible to proxy

**Ultra Anonymous:**
- ✅ Maximum anonymity
- ✅ IP hidden
- ✅ Timing obfuscation
- ✅ Cover traffic

**DHT Chat:**
- ⚠️ IP visible in DHT
- ✅ Encrypted messages
- ✅ Decentralized

## 🆚 Comparison

| Feature | Simple Chat | .hmouth | Ultra Anonymous | DHT Chat |
|---------|-------------|---------|-----------------|----------|
| Setup | Easy | Medium | Medium | Easy |
| Anonymity | Low | Medium | Maximum | Medium |
| Speed | Fast | Medium | Slow | Fast |
| Discovery | Manual | DHT | Manual | DHT |
| Best For | Friends | Hosting | Privacy | Public |

## 🎉 Examples

### Example 1: Quick Chat
```bash
cd examples/simple_backend
npm install && npm start
# Open: http://localhost:3000
```

### Example 2: Host Website
```bash
go run cmd/hmouth_proxy.go
# Open: http://localhost:8888
# Host folder as mysite.hmouth
```

### Example 3: Anonymous Chat
```bash
go run cmd/ultra_anonymous.go
# Maximum anonymity active!
```

## 🤝 Contributing

Contributions welcome! This is an experimental project.

## ⚠️ Disclaimer

This is experimental software. Not audited for production use. Use at your own risk.

## 📄 License

See LICENSE file for details.

---

**Made with ❤️ for anonymous communication**