<template>
  <div class="status-message" :class="type" v-if="visible">
    {{ message }}
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const visible = ref(false);
const message = ref('');
const type = ref<'success' | 'error' | 'info'>('info');

let timeoutId: number | null = null;

const show = (msg: string, msgType: 'success' | 'error' | 'info' = 'info') => {
  message.value = msg;
  type.value = msgType;
  visible.value = true;

  if (timeoutId) {
    clearTimeout(timeoutId);
  }

  timeoutId = window.setTimeout(() => {
    visible.value = false;
  }, 3000);
};

defineExpose({
  show,
});
</script>

<style scoped>
.status-message {
  position: fixed;
  top: 20px;
  right: 20px;
  padding: 15px 25px;
  border-radius: 6px;
  color: white;
  font-weight: 500;
  box-shadow: 0 5px 15px rgba(0,0,0,0.3);
  z-index: 1000;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.status-message.success {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.status-message.error {
  background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
}

.status-message.info {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
</style>