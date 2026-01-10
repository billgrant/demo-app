// Demo App - Frontend JavaScript

// =============================================================================
// API Functions
// =============================================================================

async function fetchHealth() {
    try {
        const response = await fetch('/health');
        return await response.json();
    } catch (error) {
        console.error('Failed to fetch health:', error);
        return null;
    }
}

async function fetchSystem() {
    try {
        const response = await fetch('/api/system');
        return await response.json();
    } catch (error) {
        console.error('Failed to fetch system:', error);
        return null;
    }
}

async function fetchItems() {
    try {
        const response = await fetch('/api/items');
        return await response.json();
    } catch (error) {
        console.error('Failed to fetch items:', error);
        return [];
    }
}

async function createItem(name, description) {
    const response = await fetch('/api/items', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, description })
    });
    return await response.json();
}

async function updateItem(id, name, description) {
    const response = await fetch(`/api/items/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, description })
    });
    return await response.json();
}

async function deleteItem(id) {
    await fetch(`/api/items/${id}`, { method: 'DELETE' });
}

async function fetchDisplay() {
    try {
        const response = await fetch('/api/display');
        return await response.json();
    } catch (error) {
        console.error('Failed to fetch display:', error);
        return {};
    }
}

async function updateDisplay(data) {
    const response = await fetch('/api/display', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
    return await response.json();
}

// =============================================================================
// Render Functions
// =============================================================================

function renderHealth(data) {
    const container = document.getElementById('health-content');

    if (!data) {
        container.innerHTML = `
            <div class="status">
                <span class="status-indicator error"></span>
                <span>Unable to connect</span>
            </div>
        `;
        return;
    }

    container.innerHTML = `
        <div class="status">
            <span class="status-indicator"></span>
            <span>${data.status}</span>
        </div>
        <div class="timestamp">${data.timestamp}</div>
    `;
}

function renderSystem(data) {
    const container = document.getElementById('system-content');

    if (!data) {
        container.innerHTML = '<div class="empty-state">Unable to load system info</div>';
        return;
    }

    const ips = data.ips && data.ips.length > 0
        ? data.ips.join(', ')
        : 'none detected';

    const envVars = data.environment && Object.keys(data.environment).length > 0
        ? Object.entries(data.environment)
            .map(([k, v]) => `<div class="info-row"><span class="info-label">${k}</span><span class="info-value">${v}</span></div>`)
            .join('')
        : '';

    container.innerHTML = `
        <div class="info-row">
            <span class="info-label">Hostname</span>
            <span class="info-value">${data.hostname || 'unknown'}</span>
        </div>
        <div class="info-row">
            <span class="info-label">IPs</span>
            <span class="info-value">${ips}</span>
        </div>
        ${envVars}
    `;
}

function renderItems(items) {
    const container = document.getElementById('items-content');

    if (!items || items.length === 0) {
        container.innerHTML = '<div class="empty-state">No items yet. Click "+ New Item" to create one.</div>';
        return;
    }

    container.innerHTML = `
        <ul class="items-list">
            ${items.map(item => `
                <li class="item-row" data-id="${item.id}">
                    <div class="item-info">
                        <div class="item-name">${escapeHtml(item.name)}</div>
                        ${item.description ? `<div class="item-description">${escapeHtml(item.description)}</div>` : ''}
                    </div>
                    <div class="item-actions">
                        <button class="secondary edit-btn" data-id="${item.id}">Edit</button>
                        <button class="danger delete-btn" data-id="${item.id}">Delete</button>
                    </div>
                </li>
            `).join('')}
        </ul>
    `;

    // Attach event listeners to edit/delete buttons
    container.querySelectorAll('.edit-btn').forEach(btn => {
        btn.addEventListener('click', () => handleEditItem(btn.dataset.id, items));
    });

    container.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', () => handleDeleteItem(btn.dataset.id));
    });
}

function renderDisplay(data) {
    const container = document.getElementById('display-content');

    if (!data || Object.keys(data).length === 0) {
        container.innerHTML = '<div class="empty-state">No display data. Click "Update Display Data" to add some.</div>';
        return;
    }

    container.innerHTML = `<pre>${escapeHtml(JSON.stringify(data, null, 2))}</pre>`;
}

// =============================================================================
// Modal Functions
// =============================================================================

function showModal(title, fields, onSave) {
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';

    const fieldHtml = fields.map(field => {
        if (field.type === 'textarea') {
            return `
                <label for="${field.name}">${field.label}</label>
                <textarea id="${field.name}" name="${field.name}">${field.value || ''}</textarea>
            `;
        }
        return `
            <label for="${field.name}">${field.label}</label>
            <input type="text" id="${field.name}" name="${field.name}" value="${field.value || ''}">
        `;
    }).join('');

    overlay.innerHTML = `
        <div class="modal">
            <h3>${title}</h3>
            ${fieldHtml}
            <div class="modal-actions">
                <button class="secondary cancel-btn">Cancel</button>
                <button class="save-btn">Save</button>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);

    // Focus first input
    const firstInput = overlay.querySelector('input, textarea');
    if (firstInput) firstInput.focus();

    // Event listeners
    overlay.querySelector('.cancel-btn').addEventListener('click', () => {
        overlay.remove();
    });

    overlay.querySelector('.save-btn').addEventListener('click', () => {
        const values = {};
        fields.forEach(field => {
            values[field.name] = overlay.querySelector(`#${field.name}`).value;
        });
        onSave(values);
        overlay.remove();
    });

    // Close on overlay click
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) overlay.remove();
    });
}

// =============================================================================
// Event Handlers
// =============================================================================

async function handleAddItem() {
    showModal('New Item', [
        { name: 'name', label: 'Name', type: 'text' },
        { name: 'description', label: 'Description', type: 'text' }
    ], async (values) => {
        if (!values.name.trim()) {
            alert('Name is required');
            return;
        }
        await createItem(values.name, values.description);
        await refreshItems();
    });
}

async function handleEditItem(id, items) {
    const item = items.find(i => i.id == id);
    if (!item) return;

    showModal('Edit Item', [
        { name: 'name', label: 'Name', type: 'text', value: item.name },
        { name: 'description', label: 'Description', type: 'text', value: item.description || '' }
    ], async (values) => {
        if (!values.name.trim()) {
            alert('Name is required');
            return;
        }
        await updateItem(id, values.name, values.description);
        await refreshItems();
    });
}

async function handleDeleteItem(id) {
    if (!confirm('Delete this item?')) return;
    await deleteItem(id);
    await refreshItems();
}

async function handleUpdateDisplay() {
    const currentData = await fetchDisplay();
    const currentJson = Object.keys(currentData).length > 0
        ? JSON.stringify(currentData, null, 2)
        : '{\n  \n}';

    showModal('Update Display Data', [
        { name: 'json', label: 'JSON Data', type: 'textarea', value: currentJson }
    ], async (values) => {
        try {
            const data = JSON.parse(values.json);
            await updateDisplay(data);
            await refreshDisplay();
        } catch (e) {
            alert('Invalid JSON: ' + e.message);
        }
    });
}

// =============================================================================
// Refresh Functions
// =============================================================================

async function refreshHealth() {
    const data = await fetchHealth();
    renderHealth(data);
}

async function refreshSystem() {
    const data = await fetchSystem();
    renderSystem(data);
}

async function refreshItems() {
    const items = await fetchItems();
    renderItems(items);
}

async function refreshDisplay() {
    const data = await fetchDisplay();
    renderDisplay(data);
}

async function refreshAll() {
    await Promise.all([
        refreshHealth(),
        refreshSystem(),
        refreshItems(),
        refreshDisplay()
    ]);
}

// =============================================================================
// Utilities
// =============================================================================

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// =============================================================================
// Initialization
// =============================================================================

document.addEventListener('DOMContentLoaded', () => {
    // Initial load
    refreshAll();

    // Button event listeners
    document.getElementById('add-item-btn').addEventListener('click', handleAddItem);
    document.getElementById('update-display-btn').addEventListener('click', handleUpdateDisplay);

    // Auto-refresh health every 10 seconds
    setInterval(refreshHealth, 10000);
});
