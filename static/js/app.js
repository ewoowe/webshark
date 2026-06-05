// 全局变量
let ws = null;
let sessionId = null;
let packetCount = 0;
let selectedInterfaces = [];

// DOM 元素
const connectBtn = document.getElementById('connect-btn');
const interfaceSection = document.getElementById('interface-section');
const interfaceList = document.getElementById('interface-list');
const filterSection = document.getElementById('filter-section');
const captureSection = document.getElementById('capture-section');
const startCaptureBtn = document.getElementById('start-capture-btn');
const stopCaptureBtn = document.getElementById('stop-capture-btn');
const clearPacketsBtn = document.getElementById('clear-packets-btn');
const packetList = document.getElementById('packet-list');
const packetDetail = document.getElementById('packet-detail');
const packetCountBadge = document.getElementById('packet-count');
const statusMessage = document.getElementById('status-message');

// 显示状态消息
function showStatus(message, type = 'info') {
    statusMessage.textContent = message;
    statusMessage.className = `status-message ${type}`;
    statusMessage.style.display = 'block';
    
    setTimeout(() => {
        statusMessage.style.display = 'none';
    }, 3000);
}

// 获取网卡列表
connectBtn.addEventListener('click', async () => {
    const host = document.getElementById('host').value.trim();
    const username = document.getElementById('username').value.trim();
    const password = document.getElementById('password').value.trim();

    if (!host || !username || !password) {
        showStatus('请填写所有连接信息', 'error');
        return;
    }

    try {
        showStatus('正在获取网卡列表...', 'info');
        
        const response = await fetch(`/api/interfaces?host=${encodeURIComponent(host)}&username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`);
        const data = await response.json();

        if (data.success) {
            renderInterfaceList(data.data);
            interfaceSection.style.display = 'block';
            filterSection.style.display = 'block';
            showStatus(`成功获取 ${data.data.length} 个网卡`, 'success');
        } else {
            showStatus(`获取网卡失败: ${data.error}`, 'error');
        }
    } catch (error) {
        showStatus(`网络错误: ${error.message}`, 'error');
    }
});

// 渲染网卡列表
function renderInterfaceList(interfaces) {
    interfaceList.innerHTML = '';
    
    interfaces.forEach((iface, index) => {
        const div = document.createElement('div');
        div.className = 'checkbox-item';
        
        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.id = `iface-${index}`;
        checkbox.value = iface.name;
        
        const label = document.createElement('label');
        label.htmlFor = `iface-${index}`;
        label.textContent = `${iface.name} (${iface.ip})`;
        
        div.appendChild(checkbox);
        div.appendChild(label);
        interfaceList.appendChild(div);
    });
}

// 全选网卡
document.getElementById('select-all-interfaces').addEventListener('click', () => {
    const checkboxes = interfaceList.querySelectorAll('input[type="checkbox"]');
    checkboxes.forEach(cb => cb.checked = true);
});

// 取消全选
document.getElementById('deselect-all-interfaces').addEventListener('click', () => {
    const checkboxes = interfaceList.querySelectorAll('input[type="checkbox"]');
    checkboxes.forEach(cb => cb.checked = false);
});

// 开始抓包
startCaptureBtn.addEventListener('click', async () => {
    const host = document.getElementById('host').value.trim();
    const username = document.getElementById('username').value.trim();
    const password = document.getElementById('password').value.trim();
    const bpfFilter = document.getElementById('bpf-filter').value.trim();
    const wiresharkFilter = document.getElementById('wireshark-filter').value.trim();

    // 获取选中的网卡
    const checkboxes = interfaceList.querySelectorAll('input[type="checkbox"]:checked');
    selectedInterfaces = Array.from(checkboxes).map(cb => cb.value);

    if (!host || !username || !password) {
        showStatus('请填写连接信息', 'error');
        return;
    }

    try {
        showStatus('正在启动抓包...', 'info');
        
        const response = await fetch('/api/capture/start', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                host: host,
                username: username,
                password: password,
                interfaces: selectedInterfaces,
                bpf_filter: bpfFilter,
                wireshark_filter: wiresharkFilter
            })
        });

        const data = await response.json();

        if (data.success) {
            sessionId = data.message;
            captureSection.style.display = 'block';
            showStatus('抓包已启动', 'success');
            
            // 连接 WebSocket
            connectWebSocket(sessionId);
        } else {
            showStatus(`启动抓包失败: ${data.error}`, 'error');
        }
    } catch (error) {
        showStatus(`网络错误: ${error.message}`, 'error');
    }
});

// 连接 WebSocket
function connectWebSocket(sessionId) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/capture?session_id=${sessionId}`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        console.log('WebSocket 连接已建立');
        showStatus('WebSocket 连接成功', 'success');
    };
    
    ws.onmessage = (event) => {
        try {
            const packet = JSON.parse(event.data);
            addPacketToTable(packet);
        } catch (error) {
            console.error('解析数据包失败:', error);
        }
    };
    
    ws.onerror = (error) => {
        console.error('WebSocket 错误:', error);
        showStatus('WebSocket 连接错误', 'error');
    };
    
    ws.onclose = () => {
        console.log('WebSocket 连接已关闭');
        showStatus('WebSocket 连接已关闭', 'info');
    };
}

// 添加数据包到表格
function addPacketToTable(packet) {
    packetCount++;
    packetCountBadge.textContent = `${packetCount} 个数据包`;
    
    const row = document.createElement('tr');
    row.dataset.packet = JSON.stringify(packet);
    
    row.innerHTML = `
        <td>${packetCount}</td>
        <td>${packet.timestamp || ''}</td>
        <td>${packet.source || ''}</td>
        <td>${packet.dest || ''}</td>
        <td>${packet.protocol || ''}</td>
        <td>${packet.length || 0}</td>
        <td>${packet.info || ''}</td>
    `;
    
    row.addEventListener('click', () => {
        // 移除之前的选中状态
        packetList.querySelectorAll('tr').forEach(r => r.classList.remove('selected'));
        row.classList.add('selected');
        
        // 显示详情
        displayPacketDetail(packet);
    });
    
    packetList.appendChild(row);
    
    // 自动滚动到底部
    const container = document.getElementById('packet-table-container');
    container.scrollTop = container.scrollHeight;
}

// 显示数据包详情
function displayPacketDetail(packet) {
    const detailText = JSON.stringify(packet, null, 2);
    packetDetail.textContent = detailText;
}

// 停止抓包
stopCaptureBtn.addEventListener('click', async () => {
    if (!sessionId) {
        showStatus('没有活跃的抓包会话', 'error');
        return;
    }

    try {
        const response = await fetch(`/api/capture/stop?session_id=${sessionId}`, {
            method: 'POST'
        });

        const data = await response.json();

        if (data.success) {
            showStatus('抓包已停止', 'success');
            
            // 关闭 WebSocket
            if (ws) {
                ws.close();
            }
            
            sessionId = null;
        } else {
            showStatus(`停止抓包失败: ${data.error}`, 'error');
        }
    } catch (error) {
        showStatus(`网络错误: ${error.message}`, 'error');
    }
});

// 清空数据包列表
clearPacketsBtn.addEventListener('click', () => {
    packetList.innerHTML = '';
    packetCount = 0;
    packetCountBadge.textContent = '0 个数据包';
    packetDetail.textContent = '点击数据包查看详细信息';
    showStatus('已清空数据包列表', 'success');
});

// 页面卸载时清理
window.addEventListener('beforeunload', () => {
    if (ws) {
        ws.close();
    }
});
