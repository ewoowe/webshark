<template>
  <section class="panel" v-if="visible">
    <h2>选择网卡</h2>
    <div class="checkbox-group">
      <div
        v-for="(iface, index) in interfaces"
        :key="index"
        class="checkbox-item"
        @click="toggleInterface(iface.name)"
      >
        <input
          type="checkbox"
          :id="`iface-${index}`"
          :value="iface.name"
          v-model="selectedInterfaces"
        />
        <label :for="`iface-${index}`">
          {{ iface.name }} ({{ iface.ip }})
        </label>
      </div>
    </div>
    <div class="button-group">
      <button class="btn btn-secondary" @click="selectAll">
        全选
      </button>
      <button class="btn btn-secondary" @click="deselectAll">
        取消全选
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import type { NetworkInterface } from '../types';

interface Props {
  visible: boolean;
  interfaces: NetworkInterface[];
  modelValue: string[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  'update:modelValue': [value: string[]];
}>();

const selectedInterfaces = ref<string[]>(props.modelValue);

watch(() => props.modelValue, (newVal) => {
  selectedInterfaces.value = newVal;
});

watch(selectedInterfaces, (newVal) => {
  emit('update:modelValue', newVal);
});

const toggleInterface = (name: string) => {
  const index = selectedInterfaces.value.indexOf(name);
  if (index > -1) {
    selectedInterfaces.value.splice(index, 1);
  } else {
    selectedInterfaces.value.push(name);
  }
};

const selectAll = () => {
  selectedInterfaces.value = props.interfaces.map(iface => iface.name);
};

const deselectAll = () => {
  selectedInterfaces.value = [];
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

.button-group {
  margin-top: 15px;
}

.checkbox-group {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 10px;
  margin-top: 15px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.checkbox-item:hover {
  background: #e9ecef;
}

.checkbox-item input[type="checkbox"] {
  margin-right: 10px;
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.checkbox-item label {
  cursor: pointer;
  flex: 1;
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

.btn-secondary {
  background: #6c757d;
  color: white;
}
</style>