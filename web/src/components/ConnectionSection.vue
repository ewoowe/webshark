<template>
  <section class="panel">
    <h2>远程主机连接</h2>
    <div class="form-group">
      <label for="host">主机地址:</label>
      <input
        type="text"
        id="host"
        v-model="connectionForm.host"
        placeholder="例如: 192.168.1.100"
        required
      />
    </div>
    <div class="form-group">
      <label for="username">用户名:</label>
      <input
        type="text"
        id="username"
        v-model="connectionForm.username"
        placeholder="例如: root"
        required
      />
    </div>
    <div class="form-group">
      <label for="password">密码:</label>
      <input
        type="password"
        id="password"
        v-model="connectionForm.password"
        placeholder="输入密码"
        required
      />
    </div>
    <button class="btn btn-primary" @click="handleConnect">
      获取网卡列表
    </button>
  </section>
</template>

<script setup lang="ts">
import { reactive } from 'vue';

const emit = defineEmits<{
  connect: [host: string, username: string, password: string];
}>();

const connectionForm = reactive({
  host: '',
  username: '',
  password: '',
});

const handleConnect = () => {
  if (!connectionForm.host || !connectionForm.username || !connectionForm.password) {
    return;
  }
  emit('connect', connectionForm.host, connectionForm.username, connectionForm.password);
};
</script>

<style scoped>
.panel {
  background: white;
  border-radius: 10px;
  padding: 25px;
  margin-bottom: 20px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.2);
}

.panel h2 {
  color: #333;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 2px solid #667eea;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #555;
  font-weight: 500;
}

.form-group input[type="text"],
.form-group input[type="password"] {
  width: 100%;
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.3s;
  box-sizing: border-box;
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  margin-right: 10px;
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(0,0,0,0.2);
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}
</style>