/**
 * Workbench Web GUI Controller & Terminal Emulator
 */

// --- ANSI Escape Code Parser & HTML Formatter ---
function ansiToHtml(text) {
    if (!text) return '';
    
    // Replace standard ANSI color escape sequences
    const ansiMap = {
        '\u001b[0m': '</span>',
        '\u001b[1m': '<span class="font-bold">',
        '\u001b[2m': '<span class="opacity-60">',
        '\u001b[3m': '<span class="italic">',
        '\u001b[4m': '<span class="underline">',
        '\u001b[30m': '<span class="text-base-content/40">',
        '\u001b[31m': '<span class="text-error font-semibold">',
        '\u001b[32m': '<span class="text-success font-semibold">',
        '\u001b[33m': '<span class="text-warning font-semibold">',
        '\u001b[34m': '<span class="text-info font-semibold">',
        '\u001b[35m': '<span class="text-secondary font-semibold">',
        '\u001b[36m': '<span class="text-accent font-semibold">',
        '\u001b[37m': '<span class="text-base-content">',
        '\u001b[90m': '<span class="text-base-content/50">',
        '\u001b[91m': '<span class="text-error">',
        '\u001b[92m': '<span class="text-success">',
        '\u001b[93m': '<span class="text-warning">',
        '\u001b[94m': '<span class="text-info">',
        '\u001b[95m': '<span class="text-secondary">',
        '\u001b[96m': '<span class="text-accent">',
        '\u001b[97m': '<span class="text-base-100">'
    };

    let result = text;
    for (const [code, html] of Object.entries(ansiMap)) {
        result = result.split(code).join(html);
    }
    // Strip any remaining uncaught ANSI codes
    result = result.replace(/\u001b\[[0-9;]*m/g, '');
    return result;
}

// --- Terminal Emulator State & Controls ---
window.TerminalEmulator = {
    autoScroll: true,
    isOpen: false,

    init() {
        const terminalOutput = document.getElementById('terminal-output');
        const autoScrollCheckbox = document.getElementById('term-autoscroll-toggle');
        
        if (autoScrollCheckbox) {
            autoScrollCheckbox.checked = this.autoScroll;
            autoScrollCheckbox.addEventListener('change', (e) => {
                this.autoScroll = e.target.checked;
            });
        }
    },

    appendLine(htmlContent) {
        const container = document.getElementById('terminal-output');
        if (!container) return;

        const line = document.createElement('div');
        line.innerHTML = ansiToHtml(htmlContent);
        container.appendChild(line);

        if (this.autoScroll) {
            container.scrollTop = container.scrollHeight;
        }
    },

    clear() {
        const container = document.getElementById('terminal-output');
        if (container) {
            container.innerHTML = '<div class="text-base-content/40 text-xs italic">[Terminal geleert]</div>';
        }
    },

    toggleModal(open) {
        const modal = document.getElementById('terminal-modal');
        if (!modal) return;
        if (open === undefined) {
            this.isOpen = !this.isOpen;
        } else {
            this.isOpen = Boolean(open);
        }
        if (this.isOpen) {
            modal.showModal();
        } else {
            modal.close();
        }
    },

    copyContent() {
        const container = document.getElementById('terminal-output');
        if (!container) return;
        const text = container.innerText;
        navigator.clipboard.writeText(text).then(() => {
            const btn = document.getElementById('term-copy-btn');
            if (btn) {
                const orig = btn.innerText;
                btn.innerText = 'Kopiert!';
                setTimeout(() => btn.innerText = orig, 1500);
            }
        });
    }
};

// --- Job Filter ---
window.filterJobs = function(query) {
    const q = (query || '').toLowerCase().trim();
    const items = document.querySelectorAll('#jobs-list .job-item');
    items.forEach(item => {
        const sig = (item.getAttribute('data-signature') || '').toLowerCase();
        const title = (item.getAttribute('data-title') || '').toLowerCase();
        const status = (item.getAttribute('data-status') || '').toLowerCase();
        if (!q || sig.includes(q) || title.includes(q) || status.includes(q)) {
            item.style.display = '';
        } else {
            item.style.display = 'none';
        }
    });
};

// --- Theme Selector Integration ---
const THEMES = [
    { id: 'dark', name: 'Dark', type: 'dark' },
    { id: 'light', name: 'Light', type: 'light' },
    { id: 'corporate', name: 'Corporate', type: 'light' },
    { id: 'emerald', name: 'Emerald', type: 'light' },
    { id: 'nord', name: 'Nord', type: 'light' },
    { id: 'night', name: 'Night', type: 'dark' },
    { id: 'dracula', name: 'Dracula', type: 'dark' },
    { id: 'business', name: 'Business', type: 'dark' },
    { id: 'dim', name: 'Dim', type: 'dark' },
    { id: 'synthwave', name: 'Synthwave', type: 'dark' },
    { id: 'forest', name: 'Forest', type: 'dark' },
    { id: 'black', name: 'Black', type: 'dark' },
    { id: 'luxury', name: 'Luxury', type: 'dark' },
    { id: 'sunset', name: 'Sunset', type: 'dark' }
];

window.setAppTheme = function(themeId) {
    document.documentElement.setAttribute('data-theme', themeId);
    try {
        localStorage.setItem('rfid_theme', themeId);
        document.cookie = `rfid_theme=${encodeURIComponent(themeId)}; path=/; max-age=31536000; SameSite=Lax`;
    } catch(e) {}
    
    const indicator = document.getElementById('current-theme-name');
    if (indicator) {
        indicator.innerText = themeId;
    }
};

window.setThemeMode = function(mode) {
    try {
        localStorage.setItem('rfid_theme_mode', mode);
        document.cookie = `rfid_theme_mode=${encodeURIComponent(mode)}; path=/; max-age=31536000; SameSite=Lax`;
    } catch(e) {}
    
    if (mode === 'dark') {
        window.setAppTheme('dark');
    } else if (mode === 'light') {
        window.setAppTheme('corporate');
    } else {
        const isDark = Boolean(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches);
        window.setAppTheme(isDark ? 'dark' : 'corporate');
    }
};

// --- Initialize Event Listeners & SSE Integration on DOM Load ---
document.addEventListener('DOMContentLoaded', () => {
    window.TerminalEmulator.init();

    // Listen for HTMX custom trigger to open terminal
    document.body.addEventListener('terminalOpen', () => {
        window.TerminalEmulator.toggleModal(true);
    });

    // Handle SSE events from hx-sse or direct EventSource
    document.body.addEventListener('htmx:sseMessage', (e) => {
        const eventName = e.detail.type;
        const data = e.detail.data;

        if (eventName === 'terminal-line') {
            window.TerminalEmulator.appendLine(data);
        } else if (eventName === 'terminal-start') {
            window.TerminalEmulator.appendLine(`<div class="text-primary font-bold my-1">=== ${data} ===</div>`);
            window.TerminalEmulator.toggleModal(true);
        } else if (eventName === 'terminal-done') {
            window.TerminalEmulator.appendLine(`<div class="text-success font-bold my-1">=== Vorgang beendet ===</div>`);
        } else if (eventName === 'refresh') {
            // Reload active batch jobs
            const activeBatch = new URLSearchParams(window.location.search).get('batch');
            if (activeBatch) {
                htmx.ajax('GET', `/batch/${encodeURIComponent(activeBatch)}`, {
                    target: '#jobs-container',
                    swap: 'innerHTML'
                });
            }
        }
    });
});
