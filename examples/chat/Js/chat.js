// Configuration
const API_URL = 'http://localhost:3000';
let currentChat = 'everyone';
let currentUser = '';
let users = [];

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadUsers();
    loadMessages();
    
    // Auto-refresh
    setInterval(loadUsers, 3000);
    setInterval(loadMessages, 2000);
    
    // Collapsible sections
    document.querySelectorAll('.collapsible').forEach(button => {
        button.addEventListener('click', function() {
            this.classList.toggle('active');
            const content = this.nextElementSibling;
            if (content.style.display === 'block') {
                content.style.display = 'none';
            } else {
                content.style.display = 'block';
            }
        });
    });
});

// Load users
async function loadUsers() {
    try {
        const response = await fetch(`${API_URL}/api/users`);
        const data = await response.json();
        users = data.users || [];
        renderUsers();
    } catch (error) {
        console.error('Failed to load users:', error);
    }
}

// Render users
function renderUsers() {
    const usersList = document.getElementById('usersList');
    const onlineCount = document.getElementById('onlineCount');
    
    if (users.length === 0) {
        usersList.innerHTML = '<p style="color: #999; text-align: center;">No users online</p>';
        if (onlineCount) onlineCount.textContent = '0';
        return;
    }
    
    if (onlineCount) onlineCount.textContent = users.length;
    
    usersList.innerHTML = users.map(user => `
        <div class="chat-item" onclick="selectChat('${user}')">
            <strong>👤 ${user}</strong>
            <p>Click to chat privately</p>
        </div>
    `).join('');
}

// Select chat
function selectChat(target) {
    currentChat = target;
    
    // Update UI
    document.querySelectorAll('.chat-item').forEach(item => item.classList.remove('active'));
    if (event && event.target) {
        event.target.closest('.chat-item').classList.add('active');
    }
    
    const displayName = target === 'everyone' ? '📢 Everyone' : '👤 ' + target;
    document.getElementById('chatDisplayName').textContent = displayName;
    document.getElementById('currentChatInfo').textContent = displayName;
    
    loadMessages();
}

// Load messages
async function loadMessages() {
    try {
        const response = await fetch(`${API_URL}/api/messages`);
        const data = await response.json();
        const messages = data.messages || [];
        
        // Filter messages
        let filtered = messages;
        if (currentChat !== 'everyone') {
            filtered = messages.filter(msg => 
                msg.to === currentChat || 
                (msg.user === currentChat && msg.to === 'everyone')
            );
        }
        
        renderMessages(filtered);
        
        const messageCount = document.getElementById('messageCount');
        if (messageCount) messageCount.textContent = messages.length;
    } catch (error) {
        console.error('Failed to load messages:', error);
    }
}

// Render messages
function renderMessages(messages) {
    const messagesDiv = document.getElementById('messages');
    
    if (messages.length === 0) {
        messagesDiv.innerHTML = `
            <div class="welcome-message">
                <h2>No messages yet</h2>
                <p>Be the first to say something! 👋</p>
            </div>
        `;
        return;
    }
    
    messagesDiv.innerHTML = messages.map(msg => {
        const time = new Date(msg.timestamp).toLocaleTimeString();
        const isSelf = msg.user === currentUser;
        const recipient = msg.to === 'everyone' ? 'Everyone' : msg.to;
        
        return `
            <div class="message ${isSelf ? 'self' : ''}">
                <div class="message-content">
                    <strong>${msg.user}</strong> ${msg.to !== 'everyone' ? `→ ${recipient}` : ''}
                    <p>${escapeHtml(msg.message)}</p>
                    <span class="timestamp">${time}</span>
                </div>
            </div>
        `;
    }).join('');
    
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

// Send message
document.getElementById('sendBtn').addEventListener('click', sendMessage);

function handleEnter(event) {
    if (event.key === 'Enter') {
        sendMessage();
    }
}

async function sendMessage() {
    const input = document.getElementById('msgInput');
    const userInput = document.getElementById('userName');
    
    const message = input.value.trim();
    const user = userInput.value.trim() || 'Anonymous';
    
    if (!message) return;
    
    currentUser = user;
    
    try {
        await fetch(`${API_URL}/api/send`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user: user,
                message: message,
                to: currentChat
            })
        });
        
        input.value = '';
        loadMessages();
    } catch (error) {
        console.error('Failed to send message:', error);
        alert('Failed to send message. Is the backend running?');
    }
}

// Utility function
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
