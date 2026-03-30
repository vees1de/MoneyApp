const BASE_URL = 'https://bims.su';

// State
let accessToken = localStorage.getItem('access_token') || '';
let refreshToken = localStorage.getItem('refresh_token') || '';
let history = JSON.parse(localStorage.getItem('request_history') || '[]');

// DOM
const methodSelect = document.getElementById('methodSelect');
const urlInput = document.getElementById('urlInput');
const sendBtn = document.getElementById('sendBtn');
const bodyEditor = document.getElementById('bodyEditor');
const headersEditor = document.getElementById('headersEditor');
const paramsEditor = document.getElementById('paramsEditor');
const responseBody = document.getElementById('responseBody');
const responseHeader = document.getElementById('responseHeader');
const responseStatus = document.getElementById('responseStatus');
const responseTime = document.getElementById('responseTime');
const responseSize = document.getElementById('responseSize');
const tokenDisplay = document.getElementById('tokenDisplay');
const statusDot = document.getElementById('statusDot');
const statusText = document.getElementById('statusText');
const historyList = document.getElementById('historyList');

// Init
updateTokenDisplay();
renderHistory();
checkHealth();

// Health check
async function checkHealth() {
  try {
    const res = await fetch(`${BASE_URL}/healthz`, { mode: 'cors' });
    if (res.ok) {
      statusDot.classList.add('connected');
      statusText.textContent = 'Connected';
    } else {
      statusText.textContent = `Error ${res.status}`;
    }
  } catch (e) {
    statusText.textContent = 'Unreachable';
  }
}

// Token management
function updateTokenDisplay() {
  if (accessToken) {
    tokenDisplay.textContent = `Bearer ${accessToken.substring(0, 32)}...`;
    tokenDisplay.style.color = 'var(--green)';
  } else {
    tokenDisplay.textContent = 'No token';
    tokenDisplay.style.color = 'var(--text2)';
  }
}

function saveTokens(access, refresh) {
  accessToken = access || '';
  refreshToken = refresh || '';
  localStorage.setItem('access_token', accessToken);
  localStorage.setItem('refresh_token', refreshToken);
  updateTokenDisplay();
}

// Quick login
document.getElementById('loginBtn').addEventListener('click', async () => {
  const email = document.getElementById('authEmail').value;
  const password = document.getElementById('authPassword').value;
  if (!email || !password) return;

  methodSelect.value = 'POST';
  urlInput.value = '/api/v1/auth/login';
  bodyEditor.value = JSON.stringify({ email, password }, null, 2);
  await sendRequest();
});

document.getElementById('registerBtn').addEventListener('click', async () => {
  const email = document.getElementById('authEmail').value;
  const password = document.getElementById('authPassword').value;
  if (!email || !password) return;

  methodSelect.value = 'POST';
  urlInput.value = '/api/v1/auth/register';
  bodyEditor.value = JSON.stringify({
    email,
    password,
    first_name: 'Test',
    last_name: 'User'
  }, null, 2);
  await sendRequest();
});

document.getElementById('clearTokenBtn').addEventListener('click', () => {
  saveTokens('', '');
});

document.querySelectorAll('.demo-login').forEach(btn => {
  btn.addEventListener('click', async () => {
    document.getElementById('authEmail').value = btn.dataset.email;
    document.getElementById('authPassword').value = btn.dataset.password;
    await document.getElementById('loginBtn').click();
  });
});

// Send request
sendBtn.addEventListener('click', sendRequest);
urlInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') sendRequest();
});

async function sendRequest() {
  const method = methodSelect.value;
  let path = urlInput.value.trim();
  if (!path) return;

  // Prompt for {id} placeholders
  if (path.includes('{id}') || path.includes('{roleId}')) {
    const id = prompt('Enter ID value:');
    if (!id) return;
    path = path.replace('{id}', id);
    path = path.replace('{roleId}', id);
    urlInput.value = path;
  }

  // Build query params
  let queryString = '';
  try {
    const paramsText = paramsEditor.value.trim();
    if (paramsText) {
      const params = JSON.parse(paramsText);
      const sp = new URLSearchParams(params);
      queryString = '?' + sp.toString();
    }
  } catch (e) { /* ignore */ }

  const url = BASE_URL + path + queryString;

  // Build headers
  const headers = {};
  try {
    const customHeaders = headersEditor.value.trim();
    if (customHeaders) Object.assign(headers, JSON.parse(customHeaders));
  } catch (e) { /* ignore */ }

  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`;
  }

  // Build options
  const options = { method, headers, mode: 'cors' };

  if (method !== 'GET' && bodyEditor.value.trim()) {
    options.body = bodyEditor.value.trim();
    if (!headers['Content-Type']) {
      headers['Content-Type'] = 'application/json';
    }
  }

  // Send
  sendBtn.disabled = true;
  sendBtn.textContent = '...';
  const startTime = performance.now();

  try {
    const res = await fetch(url, options);
    const elapsed = Math.round(performance.now() - startTime);
    const contentType = res.headers.get('content-type') || '';
    const contentDisposition = res.headers.get('content-disposition') || '';

    if (contentDisposition.includes('attachment') || contentType.includes('application/vnd.ms-excel')) {
      const blob = await res.blob();
      const size = blob.size;
      const filenameMatch = contentDisposition.match(/filename="([^"]+)"/i);
      const filename = filenameMatch ? filenameMatch[1] : 'download.xls';
      const downloadUrl = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = downloadUrl;
      anchor.download = filename;
      anchor.click();
      URL.revokeObjectURL(downloadUrl);

      responseHeader.style.display = 'flex';
      responseStatus.textContent = `${res.status} ${res.statusText}`;
      responseStatus.className = 'response-status';
      if (res.status >= 200 && res.status < 300) responseStatus.classList.add('status-2xx');
      else if (res.status >= 400 && res.status < 500) responseStatus.classList.add('status-4xx');
      else if (res.status >= 500) responseStatus.classList.add('status-5xx');
      responseTime.textContent = `${elapsed}ms`;
      responseSize.textContent = formatBytes(size);
      responseBody.innerHTML = `<pre>Downloaded file: ${filename}</pre>`;
      addHistory(method, path, res.status);
      return;
    }

    const text = await res.text();
    const size = new Blob([text]).size;

    // Show response
    responseHeader.style.display = 'flex';

    responseStatus.textContent = `${res.status} ${res.statusText}`;
    responseStatus.className = 'response-status';
    if (res.status >= 200 && res.status < 300) responseStatus.classList.add('status-2xx');
    else if (res.status >= 400 && res.status < 500) responseStatus.classList.add('status-4xx');
    else if (res.status >= 500) responseStatus.classList.add('status-5xx');

    responseTime.textContent = `${elapsed}ms`;
    responseSize.textContent = formatBytes(size);

    // Format JSON
    let formatted = text;
    try {
      const json = JSON.parse(text);
      formatted = syntaxHighlight(JSON.stringify(json, null, 2));

      // Auto-save tokens from login/register/refresh
      if (path.includes('/auth/login') || path.includes('/auth/register') || path.includes('/auth/refresh')) {
        if (json.access_token) {
          saveTokens(json.access_token, json.refresh_token || refreshToken);
        } else if (json.data && json.data.access_token) {
          saveTokens(json.data.access_token, json.data.refresh_token || refreshToken);
        }
      }
    } catch (e) { /* not json */ }

    responseBody.innerHTML = `<pre>${formatted}</pre>`;

    // Save to history
    addHistory(method, path, res.status);

  } catch (err) {
    responseHeader.style.display = 'flex';
    responseStatus.textContent = 'ERROR';
    responseStatus.className = 'response-status status-5xx';
    responseTime.textContent = '';
    responseSize.textContent = '';
    responseBody.innerHTML = `<pre style="color:var(--red)">${err.message}\n\nCheck CORS settings or network connectivity.</pre>`;

    addHistory(method, path, 'ERR');
  }

  sendBtn.disabled = false;
  sendBtn.textContent = 'Send';
}

// Sidebar items
document.querySelectorAll('.sidebar-item').forEach(item => {
  item.addEventListener('click', () => {
    const method = item.dataset.method;
    const path = item.dataset.path;
    const body = item.dataset.body;

    if (method) methodSelect.value = method;
    if (path) urlInput.value = path;
    if (body) {
      try {
        bodyEditor.value = JSON.stringify(JSON.parse(body), null, 2);
      } catch (e) {
        bodyEditor.value = body;
      }
    } else {
      bodyEditor.value = '';
    }

    document.querySelectorAll('.sidebar-item').forEach(i => i.classList.remove('active'));
    item.classList.add('active');
  });
});

// Quick action buttons
document.querySelectorAll('.quick-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    urlInput.value = btn.dataset.path;
    methodSelect.value = 'GET';
    bodyEditor.value = '';
    sendRequest();
  });
});

// Tabs
document.querySelectorAll('.tab').forEach(tab => {
  tab.addEventListener('click', () => {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    tab.classList.add('active');
    document.getElementById('tab-' + tab.dataset.tab).classList.add('active');
  });
});

// History
function addHistory(method, path, status) {
  history.unshift({ method, path, status, time: Date.now() });
  if (history.length > 50) history.pop();
  localStorage.setItem('request_history', JSON.stringify(history));
  renderHistory();
}

function renderHistory() {
  historyList.innerHTML = history.slice(0, 20).map(h => {
    const methodClass = `method-${h.method.toLowerCase()}`;
    const statusColor = typeof h.status === 'number'
      ? (h.status < 300 ? 'var(--green)' : h.status < 500 ? 'var(--orange)' : 'var(--red)')
      : 'var(--red)';
    return `
      <div class="history-item" data-method="${h.method}" data-path="${h.path}">
        <span class="method ${methodClass}">${h.method}</span>
        <span class="path">${h.path}</span>
        <span class="status-code" style="color:${statusColor}">${h.status}</span>
      </div>`;
  }).join('');

  historyList.querySelectorAll('.history-item').forEach(item => {
    item.addEventListener('click', () => {
      methodSelect.value = item.dataset.method;
      urlInput.value = item.dataset.path;
    });
  });
}

// Syntax highlighting for JSON
function syntaxHighlight(json) {
  return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
    let cls = 'color: var(--orange)'; // number
    if (/^"/.test(match)) {
      if (/:$/.test(match)) {
        cls = 'color: var(--accent)'; // key
      } else {
        cls = 'color: var(--green)'; // string
      }
    } else if (/true|false/.test(match)) {
      cls = 'color: #c084fc'; // boolean
    } else if (/null/.test(match)) {
      cls = 'color: var(--text2)'; // null
    }
    return `<span style="${cls}">${match}</span>`;
  });
}

function formatBytes(bytes) {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
  // Ctrl/Cmd + Enter to send
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    e.preventDefault();
    sendRequest();
  }
});
