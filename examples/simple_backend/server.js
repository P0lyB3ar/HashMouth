// Simple Chat Backend for .hmouth hosting
const express = require('express');
const cors = require('cors');
const path = require('path');
const app = express();

app.use(cors());
app.use(express.json());

// Serve the chat frontend
app.use(express.static(path.join(__dirname, '../chat')));

// In-memory storage
let messages = [];
let users = new Set();

// Get all messages
app.get('/api/messages', (req, res) => {
    res.json({ messages });
});

// Send message
app.post('/api/send', (req, res) => {
    const { user, message, to } = req.body;
    
    const username = user || 'Anonymous';
    
    // Add user to users list
    users.add(username);
    
    const msg = {
        id: Date.now(),
        user: username,
        message,
        to: to || 'everyone',
        timestamp: new Date().toISOString()
    };
    
    messages.push(msg);
    res.json({ success: true, message: msg });
});

// Get online users
app.get('/api/users', (req, res) => {
    res.json({ users: Array.from(users) });
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`✅ Chat backend running on http://localhost:${PORT}`);
    console.log(`📋 Host this on .hmouth proxy!`);
});
