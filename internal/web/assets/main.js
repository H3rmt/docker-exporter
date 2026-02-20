/** @type {string[]} */
let labels = JSON.parse(document.getElementById('Labels').textContent);
/** @type {number[]} */
let cpuData = JSON.parse(document.getElementById('CPUData').textContent);
/** @type {number[]} */
let cpuDataUser = JSON.parse(document.getElementById('CPUDataUser').textContent);
/** @type {number[]} */
let cpuDataSystem = JSON.parse(document.getElementById('CPUDataSystem').textContent);
/** @type {number[]} */
let memData = JSON.parse(document.getElementById('MemData').textContent);

console.log(labels);
console.log(cpuData);
console.log(cpuDataUser);
console.log(cpuDataSystem);
console.log(memData);

async function fetchJSON(url, allowErrors = false) {
    const r = await fetch(url, {cache: 'no-store'});
    if (!allowErrors && !r.ok) throw new Error('HTTP ' + r.status);
    return r.json();
}

function fmtBytesKiB(kib) {
    const bytes = kib * 1024;
    const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
    let u = 0, v = bytes;
    while (v >= 1024 && u < units.length - 1) {
        v /= 1024;
        u++;
    }
    return v.toFixed(1) + ' ' + units[u];
}

function fmtTime(ts) {
    if (!ts) return '-';
    const d = new Date(ts * 1000);
    return d.toLocaleString();
}

async function checkStatus() {
    try {
        /** @type { {status: string, version: string, docker_version?: string, docker_error?: string, errors?: Object<string, string> } } */
        const status = await fetchJSON('/status', true);
        const badge = document.getElementById('status_badge');

        if (status.status === 'starting') {
            badge.className = 'status-badge status-starting';
            badge.textContent = 'Starting...';
            badge.style.display = 'inline-block';
            return false;
        } else if (status.status === 'healthy') {
            badge.className = 'status-badge status-healthy';
            badge.textContent = 'Healthy';
            badge.style.display = 'inline-block';
            return true;
        } else if (status.status === 'unhealthy') {
            badge.className = 'status-badge status-unhealthy';
            badge.textContent = 'Unhealthy';
            badge.style.display = 'inline-block';
            if (status.errors) {
                badge.title = [...Object.entries(status.errors).map(([k, v]) => `${k}: ${v}`),].join('\n');
            } else {
                badge.title = status.docker_error || '';
            }
            return false;
        }
    } catch (e) {
        console.error('Status check failed:', e);
        const badge = document.getElementById('status_badge');
        badge.className = 'status-badge status-unhealthy';
        badge.textContent = 'Error';
        badge.style.display = 'inline-block';
    }
    return false;
}

async function loadInfo() {
    try {
        /** @type { {hostname: string, host_ip?: string, version: string, os_name: string, os_version: string} } */
        const info = await fetchJSON('/api/info');
        document.getElementById('host').textContent = info.hostname;
        document.getElementById('ip').textContent = info.host_ip;
        document.getElementById('version').textContent = '(' + info.version + ')';
        document.getElementById('link').href = 'https://github.com/h3rmt/docker-exporter/tree/' + info.version;
        // Display OS info, handling Unknown values
        if (info.os_name === 'Unknown' || info.os_version === 'Unknown') {
            document.getElementById('os_info').textContent = info.os_name !== 'Unknown' ? info.os_name : 'Unknown';
        } else {
            document.getElementById('os_info').textContent = info.os_name + ' ' + info.os_version;
        }
    } catch (e) {
        console.error(e);
    }
}

const cpuCtx = document.getElementById('cpuChart').getContext('2d');
const memCtx = document.getElementById('memChart').getContext('2d');

let cpuChart, memChart;
try {
    Chart.defaults.color = window.getComputedStyle(document.body).color;
    cpuChart = new Chart(cpuCtx, {
        type: 'line',
        data: {
            labels,
            datasets: [{
                label: 'CPU %',
                data: cpuData,
                borderColor: '#3b82f6',
                backgroundColor: 'rgba(59,130,246,0.15)',
                fill: true,
                tension: 0.25
            }, {
                label: 'CPU % (User)',
                data: cpuDataUser,
                borderColor: '#8b5cf6',
                backgroundColor: 'rgba(139,92,246,0.15)',
                fill: false,
                tension: 0.25
            }, {
                label: 'CPU % (System)',
                data: cpuDataSystem,
                borderColor: '#06b6d4',
                backgroundColor: 'rgba(6,182,212,0.15)',
                fill: false,
                tension: 0.25
            }]
        },
        options: {
            animation: {y: false, x: false},
            scales: {
                y: {min: 0, max: 100, ticks: {callback: (v) => v + '%'}}
            },
            plugins: {
                tooltip: {
                    mode: "index",
                    intersect: false
                }
            }
        }
    });
    memChart = new Chart(memCtx, {
        type: 'line',
        data: {
            labels,
            datasets: [{
                label: 'Mem %',
                data: memData,
                borderColor: '#16a34a',
                backgroundColor: 'rgba(22,163,74,0.15)',
                fill: true,
                tension: 0.25
            }]
        },
        options: {
            animation: {
                duration: 500,
                easing: 'easeOutQuart'
            },
            scales: {
                y: {min: 0, max: 100, ticks: {callback: (v) => v + '%'}}
            },
            plugins: {
                tooltip: {
                    intersect: false
                }
            },
        }
    });
} catch (e) {
    console.error('Chart initialization failed:', e);
}

function addAnimated(values, chart) {
    if (!chart) return;
    for (const array of values) {
        const lastY = array[0][memData.length - 1];
        array[0].push(lastY);
    }
    chart.update('none');

    for (const array of values) {
        array[0][array[0].length - 1] = array[1];
    }
    chart.update({
        duration: 1600,
        easing: 'easeOutCubic'
    });
}

async function tick() {
    try {
        /** @type { {cpu_percent: number, cpu_percent_user: number, cpu_percent_system: number, mem_percent: number} } */
        const u = await fetchJSON('/api/usage');
        const now = new Date();
        const label = now.toLocaleTimeString(undefined, {
            hour12: false
        });

        labels.shift();
        cpuData.shift();
        cpuDataUser.shift();
        cpuDataSystem.shift();
        memData.shift();

        console.log("adding: ", u);
        labels.push(label);
        addAnimated([[cpuData, u.cpu_percent], [cpuDataUser, u.cpu_percent_user], [cpuDataSystem, u.cpu_percent_system]], cpuChart);
        addAnimated([[memData, u.mem_percent]], memChart);
    } catch (e) {
        console.error(e);
    }
}

async function loadContainers() {
    try {
        const tbody = document.getElementById('containers');
        const loading = document.getElementById('container_loading_div');
        loading.innerHTML = '<svg width="20" height="20" viewBox="0 0 50 50" style="vertical-align:middle;margin-right:8px;" xmlns="http://www.w3.org/2000/svg">' +
            '<circle cx="25" cy="25" r="20" fill="none" stroke="#2563eb" stroke-width="4" stroke-linecap="round" stroke-dasharray="31.4 31.4" transform="rotate(-90 25 25)">' +
            '<animateTransform attributeName="transform" type="rotate" from="0 25 25" to="360 25 25" dur="1s" repeatCount="indefinite"/>' +
            '</circle></svg>' +
            '<span style="vertical-align:middle;">Loading containers...</span>';
        /** @type { {
         *   exited: boolean, names: string[],
         *   id: string, created: number, mem_usage_kib: number,
         *   mem_limit_kib: number, state: string,
         *   exit_code: number, restart_count: number,
         *   cpu_usage: number, max_cpus: number,
         *   max_limited_cpus: number, cpu_limited_usage: number
         * }[] } */
        await new Promise(resolve => setTimeout(resolve, 100));
        let list = await fetchJSON('/api/containers');
        list.sort((a, b) => {
            const exitA = a.exited ? a.exit_code : -1;
            const exitB = b.exited ? b.exit_code : -1;
            if (exitA !== exitB) {
                return exitA - exitB;
            }
            return a.created - b.created;
        });
        loading.innerHTML = "";
        tbody.innerHTML = "";
        document.getElementById('container_count').innerText = " (" + list.length + ")";
        for (const c of list) {
            const tr = document.createElement('tr');
            const stateClass = c.exited ? 'exited' : 'running';

            // Name column
            const tdName = document.createElement('td');
            tdName.innerText = (c.names && c.names.length) ? c.names[0] : '';
            tr.appendChild(tdName);

            // ID column
            const tdId = document.createElement('td');
            const code = document.createElement('code');
            code.title = c.id + '\n';
            code.classList.add('underline');
            code.innerText = c.id.substring(0, 12);
            tdId.appendChild(code);
            tr.appendChild(tdId);

            // CPU column
            const tdCpu = document.createElement('td');
            if (c.max_cpus) {
                tdCpu.classList.add('underline');
                tdCpu.innerText = (c.cpu_usage * c.max_cpus) + '% / ' + (c.max_limited_cpus * 100) + '%' + '  (' + c.cpu_limited_usage + '%)';
                tdCpu.title = (c.cpu_usage * c.max_cpus) + '% / ' + (c.max_cpus * 100) + '%' + '  (' + c.cpu_usage + '%)';
            } else {
                tdCpu.innerText = '-';
            }
            tr.appendChild(tdCpu);

            // Created column
            const tdCreated = document.createElement('td');
            tdCreated.classList.add('underline');
            const now = Date.now();
            const createdMs = c.created * 1000;
            const diffMs = now - createdMs;
            const diffSec = Math.floor(diffMs / 1000);
            const diffMin = Math.floor(diffSec / 60);
            const diffHour = Math.floor(diffMin / 60);
            const diffDay = Math.floor(diffHour / 24);

            let durationStr = '';
            if (diffDay > 0) {
                durationStr = diffDay + 'd ' + (diffHour % 24) + 'h';
            } else if (diffHour > 0) {
                durationStr = diffHour + 'h ' + (diffMin % 60) + 'm';
            } else if (diffMin > 0) {
                durationStr = diffMin + 'm ' + (diffSec % 60) + 's';
            } else {
                durationStr = diffSec + 's';
            }

            tdCreated.innerText = fmtTime(c.created);
            tdCreated.title = durationStr + " Ago"
            tr.appendChild(tdCreated);

            // Memory usage column
            const tdMem = document.createElement('td');
            if (c.mem_usage_kib && c.mem_limit_kib) {
                const memPercent = ((c.mem_usage_kib / c.mem_limit_kib) * 100).toFixed(1) + '%';
                tdMem.innerText = fmtBytesKiB(c.mem_usage_kib) + ' / ' + fmtBytesKiB(c.mem_limit_kib) + ' (' + memPercent + ')';
            } else {
                tdMem.innerText = ' - ';
            }
            tr.appendChild(tdMem);

            // Status column
            const tdStatus = document.createElement('td');
            const statusSpan = document.createElement('span');
            statusSpan.className = 'status ' + stateClass;
            statusSpan.innerText = c.state + (c.exited ? (' (exit=' + c.exit_code + ')') : '') + (c.restart_count ? ' (' + c.restart_count + ')' : '');
            tdStatus.appendChild(statusSpan);
            tr.appendChild(tdStatus);

            tbody.appendChild(tr);
        }
    } catch (e) {
        console.error(e);
    }
}

// Initialize page
(async function init() {
    document.getElementById('totalMem').innerText = fmtBytesKiB(parseInt(document.getElementById("TotalMemData").textContent));
    document.getElementById('totalCPU').innerText = document.getElementById("CpuCountData").textContent + " Cores";

    // both are already available before the server reports ready
    loadInfo().then();
    tick().then();
    setInterval(tick, 2000);

    // remove flex: 1 css attribute from cards
    const cards = Array.from(document.getElementsByClassName('card-container'));
    cards.map((i) => i.classList.replace("card-container", "card-container-no-flex"))

    // Check status first and poll until ready
    let isOk = await checkStatus();

    if (!isOk) {
        // Poll status 30 times every second (max 30 attempts = 30 seconds)
        let attempts = 0;
        const maxAttempts = 30;

        while (!isOk && attempts < maxAttempts) {
            await new Promise(resolve => setTimeout(resolve, 1000));
            attempts++;
            isOk = await checkStatus();
        }
    }


    if (!isOk) {
        console.error('Server failed to become ready after 30 seconds');
        const badge = document.getElementById('status_badge');
        badge.className = 'status-badge status-unhealthy';
        badge.textContent = 'Startup timeout';
        badge.style.display = 'inline-block';
    } else {
        setTimeout(loadContainers, 750);
        setInterval(loadContainers, 30000);
        document.getElementById('updateBtn').addEventListener('click', loadContainers);
    }
})();