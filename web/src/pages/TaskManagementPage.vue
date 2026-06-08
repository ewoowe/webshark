<template>
  <div class="task-management-page">
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">任务管理</h3>
      </div>
      <div class="card-body">
        <p class="placeholder-text">任务管理功能开发中...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.task-management-page {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f2f5;
}

.card-title {
  margin: 0;
  font-size: 16px;
  color: #333;
  font-weight: 600;
}

.card-body {
  padding: 60px 24px;
  text-align: center;
}

.placeholder-text {
  font-size: 16px;
  color: #999;
}
</style>
