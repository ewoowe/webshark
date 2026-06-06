<template>
  <section class="panel" v-if="visible">
    <h2>过滤器配置</h2>
    <div class="form-group">
      <label for="bpf-filter">BPF 过滤器 (tcpdump):</label>
      <input
        type="text"
        id="bpf-filter"
        v-model="filters.bpfFilter"
        placeholder="例如: tcp port 80"
      />
      <small>常用示例: tcp port 80, host 192.168.1.1, not port 22</small>
    </div>
    <div class="form-group">
      <label for="wireshark-filter">Wireshark 过滤器:</label>
      <input
        type="text"
        id="wireshark-filter"
        v-model="filters.wiresharkFilter"
        placeholder="例如: http.request.method == GET"
      />
      <small>常用示例: http, dns, ip.addr == 192.168.1.1</small>
    </div>
    <button class="btn btn-success" @click="handleStartCapture">
      开始抓包
    </button>
  </section>
</template>

<script setup lang="ts">
import { reactive } from 'vue';

interface Props {
  visible: boolean;
}

defineProps<Props>();
const emit = defineEmits<{
  startCapture: [];
}>();

const filters = reactive({
  bpfFilter: '',
  wiresharkFilter: '',
});

const handleStartCapture = () => {
  emit('startCapture');
};

const getFilters = () => ({
  bpfFilter: filters.bpfFilter,
  wiresharkFilter: filters.wiresharkFilter,
});

defineExpose({
  getFilters,
});
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

.form-group small {
  display: block;
  margin-top: 5px;
  color: #888;
  font-size: 12px;
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

.btn-success {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: white;
}
</style>