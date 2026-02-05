package web

import (
	"html/template"
	"net/http"

	"github.com/h3rmt/docker-exporter/internal/log"
)

// HandleRoot serves a single HTML page with charts and containers table
func HandleRoot() http.HandlerFunc {
	tmpl := template.Must(template.New("page").Parse(pageTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataPoints := GetData()
		log.GetLogger().Debug("data points", "dataPoints", len(dataPoints))
		totalMem, _, err := readMemInfo(ctx)
		if err != nil {
			log.GetLogger().Error("failed to read mem info", "error", err)
		}
		_, _, _, _, cpuCount, err := readProcStat(ctx)
		if err != nil {
			log.GetLogger().Error("failed to read cpu", "error", err)
		}

		// Prepare initial data as JSON
		type chartData struct {
			Labels      []string
			CPUData     []float64
			CPUDataUser []float64
			CPUDataSys  []float64
			MemData     []float64
			TotalMem    uint64
			CpuCount    uint64
		}

		cd := chartData{
			Labels:      make([]string, 0),
			CPUData:     make([]float64, 0),
			CPUDataUser: make([]float64, 0),
			CPUDataSys:  make([]float64, 0),
			MemData:     make([]float64, 0),
			TotalMem:    totalMem / 1024, // turn into KiB
			CpuCount:    cpuCount,
		}

		for i := range dataPoints {
			if !dataPoints[i].Time.IsZero() {
				cd.Labels = append(cd.Labels, dataPoints[i].Time.Format("15:04:05"))
				cd.CPUData = append(cd.CPUData, dataPoints[i].Data.CPUPercent)
				cd.CPUDataUser = append(cd.CPUDataUser, dataPoints[i].Data.CPUPercentUser)
				cd.CPUDataSys = append(cd.CPUDataSys, dataPoints[i].Data.CPUPercentSystem)
				cd.MemData = append(cd.MemData, dataPoints[i].Data.MemPercent)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, cd)
	}
}

// language=html
const pageTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Docker Exporter</title>
  <script src="https://cdn.jsdelivr.net/npm/chart.js@4.5.1/dist/chart.umd.min.js" integrity="sha256-SERKgtTty1vsDxll+qzd4Y2cF9swY9BCq62i9wXJ9Uo=" crossorigin="anonymous"></script>
  <style>
	:root { color-scheme: light dark; }
	html { background-color: light-dark(#ffffff, #0e0e0e); color: light-dark(#292524, #f5f5f4); }
	body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; display: flex; flex-direction: column; margin: 0; height: 100dvh; width: 100vw; }
	main { padding: 10px 20px 20px; flex: 1; display: flex; flex-direction: column; gap: 1rem; overflow: auto; }
	h3 { margin:0; }
	h1 { margin:0; }

	.header { display:flex; gap:1rem; justify-content:space-between; align-items: center; }
	.header-right { display:flex; align-items:end; gap: 12px; flex-direction: column; justify-content: center; }

	.card { border: 1px solid light-dark(#ddd, #373737); border-radius: 8px; padding: 10px; box-shadow: 0 0 10px 1px light-dark(rgba(0 0 0 / 15%), rgb(200 200 200 / 15%)); display: flex; flex-direction: column; min-height: 30vh; max-height: 45vh; flex: 1 }
	.card-container { flex: 1; display: flex; }
	.card-container-no-flex { display: flex; justify-content: center }

	.chart-container { flex: 1; overflow: auto; }

	table { width:100%; border-collapse: collapse; }
	th, td { padding: 8px; border-bottom: 1px solid light-dark(#ddd, #373737); text-align: left; }
    th { background: light-dark(#eeeeee, #282828); position: sticky; top: 0; }
    code { background: light-dark(#e1e1e1, #373737); padding:2px 4px; border-radius:4px; }

	.status { padding:2px 6px; border-radius: 10px; font-size: 12px; }
	.running { background:light-dark(#e6ffed, #1a3d1a); color:light-dark(#036400, #4ade80); }
	.exited { background:light-dark(#ffeaea, #3d1a1a); color:light-dark(#8a0000, #f87171); }

	.underline { text-decoration: dashed underline; }
  </style>
 </head>
<body>
<main>
  <div class="header">
	<h1 id="host"></h1>
	<h2 id="ip"></h2>
	<div class="header-right">
      <div>
        <a href="https://github.com/h3rmt/docker-exporter" target="_blank" id="link">Docker Exporter</a>
        <span id="version"></span>
      </div>
      <div>
        <a href="/metrics">metrics</a>
        <a href="/status">status</a>
      </div>
    </div>
  </div>

  <div style="display:flex; gap: 16px; justify-content: center;">
    <div class="card-container"><div class="card">
		<h3>CPU Utilization (<span id="totalCPU"></span>)</h3>
        <div class="chart-container"><canvas id="cpuChart"></canvas></div>
    </div></div>
    <div class="card-container"><div class="card">
        <h3>Memory Utilization (<span id="totalMem"></span>)</h3>
        <div class="chart-container"><canvas id="memChart"></canvas></div>
    </div></div>
  </div>

  <div class="card">
	  <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:8px;">
        <h3 style="margin:0;">Containers<span id="container_count"></span></h3>
		<div style="display:flex; gap:4px;" id="container_loading_div"></div>
        <button id="updateBtn" style="background:#2563eb; color:white; border:none; padding:6px 12px; border-radius:6px; font-size:14px; cursor:pointer; font-family:inherit;">Update</button>
	  </div>
	  <div style="overflow-y:scroll;flex:1;">
	  <table>
	    <thead>
		  <tr>
  	        <th>Name</th>
    	    <th>ID</th>
    	    <th>CPU</th>
    		<th>Created</th>
  		    <th>Mem Usage</th>
  		    <th>Status</th>
		  </tr>
	    </thead>
	    <tbody id="containers"></tbody>
	  </table>
	  </div>
  </div>
</main>
<script>
async function fetchJSON(url){
  const r = await fetch(url, {cache:'no-store'});
  if(!r.ok) throw new Error('HTTP '+r.status);
  return r.json();
}

function fmtBytesKiB(kib){
  const bytes = kib * 1024;
  const units=['B','KiB','MiB','GiB','TiB'];
  let u=0, v=bytes;
  while(v>=1024 && u<units.length-1){ v/=1024; u++; }
  return v.toFixed(1)+' '+units[u];
}

function fmtTime(ts){
  if(!ts) return '-';
  const d = new Date(ts*1000);
  return d.toLocaleString();
}

async function loadInfo(){
  try{
    /** @type { {hostname: string, host_ip?: string, version: string} } */
	const info = await fetchJSON('/api/info');
	document.getElementById('host').textContent = info.hostname;
	document.getElementById('ip').textContent = info.host_ip;
	document.getElementById('version').textContent = '(' + info.version + ')';
	document.getElementById('link').href = 'https://github.com/h3rmt/docker-exporter/tree/' + info.version;
  } catch(e){ console.error(e); }
}

const cpuCtx = document.getElementById('cpuChart').getContext('2d');
const memCtx = document.getElementById('memChart').getContext('2d');

/** @type {string[]} */
let labels={{.Labels}};
/** @type {number[]} */
let cpuData={{.CPUData}};
/** @type {number[]} */
let cpuDataUser={{.CPUDataUser}};
/** @type {number[]} */
let cpuDataSystem={{.CPUDataSys}};
/** @type {number[]} */
let memData={{.MemData}};

console.log(labels);
console.log(cpuData);
console.log(cpuDataUser);
console.log(cpuDataSystem);
console.log(memData);

Chart.defaults.color = window.getComputedStyle(document.body).color;
const cpuChart = new Chart(cpuCtx, { 
    type:'line',
	data: {
		labels,
		datasets: [ {
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
		} ]
	},
	options: { 
        animation: { y: false, x: false },
        scales: { 
            y: { min: 0, max: 100, ticks: { callback: (v) => v + '%' }}
		},
        plugins: {
            tooltip: {
                mode: "index",
                intersect: false
            }
        }
    }
});
const memChart = new Chart(memCtx, { 
    type: 'line',
    data: {
        labels,
        datasets:[{
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
            y: { min: 0, max: 100, ticks: { callback: (v) => v + '%' }}
		},
        plugins: {
            tooltip: {
                intersect: false
            }
        },
    }
});
function addAnimated(values, chart) {
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
  try{
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
  } catch(e){ console.error(e); }
}

async function loadContainers() {
  try{
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
	let list = await fetchJSON('/api/containers');
	list.sort((a,b) => {
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
	for(const c of list){
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
	  code.innerText = c.id.substring(0,12);
	  tdId.appendChild(code);
	  tr.appendChild(tdId);
	  
	  // CPU column
	  const tdCpu = document.createElement('td');
      if (c.max_cpus) {
      	  tdCpu.classList.add('underline');
          tdCpu.innerText = (c.cpu_usage * c.max_cpus) + '% / ' + (c.max_limited_cpus * 100) + '%' + '  (' + c.cpu_limited_usage + '%)';
          tdCpu.title = (c.cpu_usage * c.max_cpus)+ '% / ' + (c.max_cpus * 100) + '%' + '  (' + c.cpu_usage + '%)';
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
  } catch(e){ console.error(e); }
}

loadInfo();
tick();
loadContainers();
setInterval(tick, 2000);
setInterval(loadContainers, 30000);
document.getElementById('updateBtn').addEventListener('click', loadContainers);
// remove flex: 1 css attribute from cards
const cards = Array.from(document.getElementsByClassName('card-container'));
cards.map((i) => i.classList.replace("card-container", "card-container-no-flex"))
document.getElementById('totalMem').innerText = fmtBytesKiB({{.TotalMem}});
document.getElementById('totalCPU').innerText = "{{.CpuCount}} Cores";
</script>
</body>
</html>`
