import { createApp } from 'vue';
import App from '@/App.vue';
import router from '@/router';
import VXETable from 'vxe-table';
import 'vxe-table/lib/style.css';
import '@/styles/main.css';

createApp(App).use(router).use(VXETable).mount('#app');