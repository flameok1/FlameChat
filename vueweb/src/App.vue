<template>
  <div class="app">
    <div v-if="!inChatRoom" class="room-list">
      <h1>聊天室列表</h1>
      <div class="rooms">
        <div v-for="room in rooms" :key="room.roomid" class="room-item" @click="joinRoom(room.roomid)">
          {{ room.roomname }}
        </div>
      </div>
      <button class="create-room-btn" @click="showCreateRoom = true">創建新房間</button>
      
      <div v-if="showCreateRoom" class="create-room-modal">
        <div class="modal-content">
          <h2>創建新房間</h2>
          <input v-model="newRoomName" placeholder="輸入房間名稱" />
          <div class="modal-buttons">
            <button @click="createRoom">確認</button>
            <button @click="showCreateRoom = false">取消</button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="chat-room">
      <div class="chat-header">
        <h2>{{ currentRoom.roomname }}</h2>
        <button @click="leaveRoom">離開房間</button>
      </div>
      
      <div class="nickname-section">
        <input v-model="nickname" placeholder="設置暱稱" />
      </div>

      <div class="messages" ref="messagesContainer">
        <div v-for="(msg, index) in messages" 
             :key="index" 
             :class="['message-container', msg.nickname === nickname ? 'self' : 'other']">
          <div class="message-box">
            <div class="nickname">{{ msg.nickname }}</div>
            <div class="content">{{ msg.message }}</div>
            <div class="time">{{ msg.time }}</div>
          </div>
        </div>
      </div>

      <div class="input-area">
        <input v-model="messageInput" @keyup.enter="sendMessage" placeholder="輸入訊息..." />
        <button @click="sendMessage">發送</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'

const rooms = ref([])
const showCreateRoom = ref(false)
const newRoomName = ref('')
const inChatRoom = ref(false)
const currentRoom = ref(null)
const nickname = ref('')
const messageInput = ref('')
const messages = ref([])
const ws = ref(null)
const messagesContainer = ref(null)

// 獲取房間列表
const fetchRooms = async () => {
  try {
    const response = await fetch('http://localhost:8080/getrooms')
    rooms.value = await response.json()
  } catch (error) {
    console.error('Error fetching rooms:', error)
  }
}

// 創建新房間
const createRoom = () => {
  if (!newRoomName.value) return
  
  ws.value = new WebSocket('ws://localhost:8080/ws')
  
  ws.value.onopen = () => {
    ws.value.send(JSON.stringify({
      protocol: 'openroom',
      roomname: newRoomName.value
    }))
  }
  
  ws.value.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.protocol === 'resopenroom' && data.status === 'ok') {
      currentRoom.value = {
        roomid: data.roomid,
        roomname: newRoomName.value
      }
      inChatRoom.value = true
      showCreateRoom.value = false
      newRoomName.value = ''
      setupChatHandlers()
    }
  }
}

// 加入房間
const joinRoom = (roomId) => {
  ws.value = new WebSocket('ws://localhost:8080/ws')
  
  ws.value.onopen = () => {
    ws.value.send(JSON.stringify({
      protocol: 'joinroom',
      roomid: roomId
    }))
  }
  
  ws.value.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.protocol === 'resjoinroom' && data.status === 'ok') {
      const room = rooms.value.find(r => r.roomid === roomId)
      if (room) {
        currentRoom.value = room
        inChatRoom.value = true
        setupChatHandlers()
      }
    }
  }
}

// 設置聊天消息處理器
const setupChatHandlers = () => {
  if (!ws.value) return
  
  ws.value.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.protocol === 'message') {
      messages.value.push({
        nickname: data.nickname,
        message: data.message,
        time: new Date().toLocaleTimeString()
      })
      scrollToBottom()
    }
  }
}

// 發送訊息
const sendMessage = () => {
  if (!messageInput.value || !nickname.value || !ws.value) return
  
  ws.value.send(JSON.stringify({
    protocol: 'message',
    nickname: nickname.value,
    message: messageInput.value,
    time: new Date().toLocaleTimeString()
  }))
  
  messageInput.value = ''
}

// 離開房間
const leaveRoom = () => {
  if (ws.value) {
    ws.value.close()
    ws.value = null
  }
  inChatRoom.value = false
  currentRoom.value = null
  messages.value = []
  fetchRooms()
}

// 滾動到最新消息
const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

onMounted(() => {
  fetchRooms()
})
</script>

<style scoped>
.app {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.room-list {
  text-align: center;
}

.rooms {
  display: grid;
  gap: 10px;
  margin: 20px 0;
}

.room-item {
  padding: 15px;
  background: #f5f5f5;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.3s;
}

.room-item:hover {
  background: #e0e0e0;
}

.create-room-btn {
  padding: 10px 20px;
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.create-room-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  width: 300px;
}

.modal-buttons {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.chat-room {
  display: flex;
  flex-direction: column;
  height: 80vh;
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background: #f5f5f5;
}

.nickname-section {
  padding: 10px;
  background: #eee;
}

.messages {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
  background: #fff;
  border: 1px solid #ddd;
}

.message-container {
  display: flex;
  margin: 10px 0;
  width: 100%;
}

.message-container.self {
  justify-content: flex-start;
}

.message-container.other {
  justify-content: flex-end;
}

.message-box {
  max-width: 70%;
  padding: 8px;
  border-radius: 12px;
  background: #f0f0f0;
}

.message-container.self .message-box {
  background: #e3f2fd;
  margin-right: 20%;
}

.message-container.other .message-box {
  background: #f5f5f5;
  margin-left: 20%;
}

.nickname {
  font-weight: bold;
  font-size: 0.9em;
  margin-bottom: 4px;
  color: #333;
}

.content {
  word-break: break-word;
  margin: 4px 0;
}

.time {
  font-size: 0.8em;
  color: #666;
  text-align: right;
}

.input-area {
  display: flex;
  gap: 10px;
  padding: 10px;
  background: #f5f5f5;
}

input {
  flex: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

button {
  padding: 8px 15px;
  background: #2196F3;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background: #1976D2;
}
</style>
