class BridgeComparator {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.bridges = [];
        this.dynasties = [];
        this.selectedBridgeA = null;
        this.selectedBridgeB = null;
        this.comparisonResult = null;
        this.initialized = false;
    }

    async init() {
        if (this.initialized) return;
        this.initialized = true;
        this.renderUI();
        await this.loadDynasties();
        await this.loadBridges();
        this.bindEvents();
    }

    async loadDynasties() {
        try {
            const response = await fetch('/api/v1/comparison/dynasties');
            const data = await response.json();
            if (data.code === 0) {
                this.dynasties = data.data;
            }
        } catch (e) {
            this.dynasties = [
                { id: 'han_jin', name: '汉晋', period: '公元前206年-公元420年', description: '简支木梁桥成熟时期' },
                { id: 'tang', name: '唐代', period: '公元618年-907年', description: '木拱技术萌芽时期' },
                { id: 'song', name: '宋代', period: '公元960年-1279年', description: '贯木拱技术鼎盛时期' },
                { id: 'ming', name: '明代', period: '公元1368年-1644年', description: '木拱廊桥发展时期' },
                { id: 'qing', name: '清代', period: '公元1644年-1912年', description: '工艺精细化与装饰时期' },
            ];
        }
        this.renderDynastyFilters();
    }

    async loadBridges(dynasty = null) {
        try {
            const url = dynasty ? `/api/v1/comparison/bridges?dynasty=${dynasty}` : '/api/v1/comparison/bridges';
            const response = await fetch(url);
            const data = await response.json();
            if (data.code === 0) {
                this.bridges = data.data;
            }
        } catch (e) {
            this.bridges = [
                { id: 101, name: '灞桥', dynasty: 'han_jin', typology: 'beam_bridge', span_length: 18.0, arch_rise: 2.5, deck_width: 4.5, key_innovation: '多跨简支木梁桥，石砌桥墩' },
                { id: 102, name: '枫桥', dynasty: 'tang', typology: 'beam_bridge', span_length: 18.5, arch_rise: 3.8, deck_width: 4.2, key_innovation: '木拱技术萌芽，向拱结构过渡' },
                { id: 103, name: '汴水虹桥', dynasty: 'song', typology: 'through_arch', span_length: 25.6, arch_rise: 5.8, deck_width: 6.5, key_innovation: '贯木拱技术成熟，无柱大跨木拱桥' },
                { id: 104, name: '龙津桥', dynasty: 'ming', typology: 'gallery_bridge', span_length: 28.5, arch_rise: 6.2, deck_width: 5.2, key_innovation: '木拱廊桥，廊屋保护结构' },
                { id: 105, name: '千乘桥', dynasty: 'ming', typology: 'timber_arch', span_length: 27.3, arch_rise: 5.9, deck_width: 5.0, key_innovation: '三节拱五节拱组合技术' },
                { id: 106, name: '飞虹桥', dynasty: 'qing', typology: 'gallery_bridge', span_length: 19.5, arch_rise: 4.2, deck_width: 4.5, key_innovation: '工艺精细化与装饰艺术发展' },
                { id: 107, name: '安澜桥', dynasty: 'song', typology: 'beam_bridge', span_length: 24.0, arch_rise: 5.2, deck_width: 4.8, key_innovation: '竹索加固木梁技术' },
            ];
        }
        this.renderBridgeList();
    }

    renderUI() {
        this.container.innerHTML = `
            <div class="panel-section">
                <h3 class="section-title">
                    <i class="fas fa-landmark"></i> 历史时期桥梁技术对比
                </h3>
                <p class="section-desc">对比不同历史时期木拱桥的结构效率与技术演进</p>
                
                <div class="dynasty-filters" id="hc-dynasty-filters"></div>

                <div class="bridge-selector">
                    <div class="bridge-select-col">
                        <h4>桥梁 A (对比基准)</h4>
                        <div id="hc-bridge-a-list" class="bridge-list"></div>
                    </div>
                    <div class="bridge-select-col">
                        <h4>桥梁 B (对比对象)</h4>
                        <div id="hc-bridge-b-list" class="bridge-list"></div>
                    </div>
                </div>

                <button id="hc-compare" class="btn-primary" disabled>
                    <i class="fas fa-chart-radar"></i> 开始对比
                </button>

                <div id="hc-loading" class="loading" style="display:none;">
                    <div class="spinner"></div>
                    <span>正在进行技术对比分析...</span>
                </div>
            </div>

            <div id="hc-results" class="results-section" style="display:none;">
                <div class="comparison-header">
                    <div class="bridge-info-card">
                        <h5 id="hc-bridge-a-name">-</h5>
                        <p id="hc-bridge-a-dynasty">-</p>
                        <span id="hc-bridge-a-type" class="bridge-type-tag">-</span>
                    </div>
                    <div class="vs-badge">VS</div>
                    <div class="bridge-info-card">
                        <h5 id="hc-bridge-b-name">-</h5>
                        <p id="hc-bridge-b-dynasty">-</p>
                        <span id="hc-bridge-b-type" class="bridge-type-tag">-</span>
                    </div>
                </div>

                <div class="radar-chart-container">
                    <h5>结构效率雷达图</h5>
                    <canvas id="hc-radar-chart" width="500" height="400"></canvas>
                </div>

                <div class="comparison-tables">
                    <div class="metrics-table">
                        <h5>技术指标对比</h5>
                        <table id="hc-metrics-table">
                            <thead>
                                <tr>
                                    <th>指标</th>
                                    <th id="hc-th-a">-</th>
                                    <th id="hc-th-b">-</th>
                                    <th>优势方</th>
                                </tr>
                            </thead>
                            <tbody id="hc-metrics-body"></tbody>
                        </table>
                    </div>
                </div>

                <div class="advantages-section">
                    <div class="advantages-col">
                        <h5><i class="fas fa-check-circle" style="color: #3498db;"></i> 桥梁A优势</h5>
                        <ul id="hc-adv-a"></ul>
                    </div>
                    <div class="advantages-col">
                        <h5><i class="fas fa-check-circle" style="color: #e74c3c;"></i> 桥梁B优势</h5>
                        <ul id="hc-adv-b"></ul>
                    </div>
                </div>

                <div class="evolution-section">
                    <h5>木拱桥技术演进史</h5>
                    <div id="hc-evolution-timeline" class="timeline"></div>
                </div>

                <div class="historical-notes">
                    <h5>历史注解</h5>
                    <ul id="hc-historical-notes"></ul>
                </div>
            </div>
        `;
    }

    renderDynastyFilters() {
        const container = document.getElementById('hc-dynasty-filters');
        if (!container) return;

        container.innerHTML = `
            <button class="dynasty-btn active" data-dynasty="">全部</button>
            ${this.dynasties.map(d => `
                <button class="dynasty-btn" data-dynasty="${d.id}">
                    ${d.name}
                    <span class="dynasty-period">${d.period.split('(')[0]}</span>
                </button>
            `).join('')}
        `;

        container.querySelectorAll('.dynasty-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                container.querySelectorAll('.dynasty-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                this.loadBridges(btn.dataset.dynasty);
            });
        });
    }

    renderBridgeList() {
        const renderList = (containerId, selectedVar, onSelect) => {
            const container = document.getElementById(containerId);
            if (!container) return;

            container.innerHTML = this.bridges.map(bridge => {
                const isSelected = this[selectedVar] === bridge.id;
                const dynastyInfo = this.dynasties.find(d => d.id === bridge.dynasty);
                return `
                    <div class="bridge-card ${isSelected ? 'selected' : ''}" data-id="${bridge.id}">
                        <div class="bridge-card-header">
                            <span class="bridge-name">${bridge.name}</span>
                            <span class="bridge-dynasty">${dynastyInfo ? dynastyInfo.name : bridge.dynasty}</span>
                        </div>
                        <div class="bridge-card-body">
                            <span>跨径: ${bridge.span_length}m</span>
                            <span>矢高: ${bridge.arch_rise}m</span>
                        </div>
                        <div class="bridge-card-footer">
                            <span class="bridge-typology">${this.getTypologyName(bridge.typology)}</span>
                        </div>
                    </div>
                `;
            }).join('');

            container.querySelectorAll('.bridge-card').forEach(card => {
                card.addEventListener('click', () => {
                    const bridgeId = parseInt(card.dataset.id);
                    this[selectedVar] = bridgeId;
                    this.renderBridgeList();
                    this.updateCompareButton();
                    onSelect && onSelect(bridgeId);
                });
            });
        };

        renderList('hc-bridge-a-list', 'selectedBridgeA');
        renderList('hc-bridge-b-list', 'selectedBridgeB');
    }

    getTypologyName(type) {
        const names = {
            'beam_bridge': '木梁桥',
            'timber_arch': '木拱桥',
            'through_arch': '贯木拱',
            'gallery_bridge': '木拱廊桥'
        };
        return names[type] || type;
    }

    updateCompareButton() {
        const btn = document.getElementById('hc-compare');
        btn.disabled = !(this.selectedBridgeA && this.selectedBridgeB && 
                        this.selectedBridgeA !== this.selectedBridgeB);
    }

    bindEvents() {
        document.getElementById('hc-compare').addEventListener('click', () => this.runComparison());
    }

    async runComparison() {
        if (!this.selectedBridgeA || !this.selectedBridgeB) return;

        document.getElementById('hc-compare').disabled = true;
        document.getElementById('hc-loading').style.display = 'flex';
        document.getElementById('hc-results').style.display = 'none';

        try {
            const response = await fetch('/api/v1/comparison/bridges', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    bridge_a_id: this.selectedBridgeA,
                    bridge_b_id: this.selectedBridgeB
                })
            });

            const data = await response.json();
            if (data.code === 0) {
                this.comparisonResult = data.data;
                this.displayResults();
            } else {
                this.runLocalComparison();
            }
        } catch (e) {
            console.error('Comparison failed:', e);
            this.runLocalComparison();
        } finally {
            document.getElementById('hc-compare').disabled = false;
            document.getElementById('hc-loading').style.display = 'none';
        }
    }

    runLocalComparison() {
        const bridgeA = this.bridges.find(b => b.id === this.selectedBridgeA);
        const bridgeB = this.bridges.find(b => b.id === this.selectedBridgeB);

        const calculateMetrics = (bridge) => {
            const riseSpanRatio = bridge.arch_rise / bridge.span_length;
            let materialEfficiency = 60;
            let loadCapacity = 50;
            let complexity = 40;
            let durability = 50;

            if (bridge.typology === 'through_arch') {
                materialEfficiency = 90;
                loadCapacity = 85;
                complexity = 78;
                durability = 65;
            } else if (bridge.typology === 'gallery_bridge') {
                materialEfficiency = 85;
                loadCapacity = 88;
                complexity = 90;
                durability = 85;
            } else if (bridge.typology === 'timber_arch') {
                materialEfficiency = 78;
                loadCapacity = 90;
                complexity = 75;
                durability = 70;
            }

            return {
                span_to_depth_ratio: riseSpanRatio * 100,
                material_efficiency: materialEfficiency,
                load_carrying_capacity: loadCapacity,
                construction_complexity: complexity,
                durability_score: durability,
                weight_to_span_ratio: bridge.typology === 'beam_bridge' ? 40 : 80
            };
        };

        const metricsA = calculateMetrics(bridgeA);
        const metricsB = calculateMetrics(bridgeB);

        const metrics = [
            { name: '跨高比', a: metricsA.span_to_depth_ratio, b: metricsB.span_to_depth_ratio },
            { name: '材料效率', a: metricsA.material_efficiency, b: metricsB.material_efficiency },
            { name: '承载能力', a: metricsA.load_carrying_capacity, b: metricsB.load_carrying_capacity },
            { name: '施工复杂度', a: metricsA.construction_complexity, b: metricsB.construction_complexity },
            { name: '耐久性评分', a: metricsA.durability_score, b: metricsB.durability_score },
            { name: '自重跨度比', a: metricsA.weight_to_span_ratio, b: metricsB.weight_to_span_ratio },
        ];

        const radarData = metrics.map(m => {
            const max = Math.max(m.a, m.b);
            return {
                metric: m.name,
                value_a: (m.a / max) * 100,
                value_b: (m.b / max) * 100
            };
        });

        const normalizedScores = {};
        metrics.forEach(m => {
            const max = Math.max(m.a, m.b);
            normalizedScores[m.name] = {
                [bridgeA.dynasty]: (m.a / max) * 100,
                [bridgeB.dynasty]: (m.b / max) * 100
            };
        });

        const advantagesA = [];
        const advantagesB = [];
        metrics.forEach(m => {
            if (m.a > m.b * 1.1) {
                advantagesA.push(`${m.name}更优: 高出 ${((m.a - m.b) / m.b * 100).toFixed(0)}%`);
            } else if (m.b > m.a * 1.1) {
                advantagesB.push(`${m.name}更优: 高出 ${((m.b - m.a) / m.a * 100).toFixed(0)}%`);
            }
        });

        if (advantagesA.length === 0) advantagesA.push('综合性能均衡');
        if (advantagesB.length === 0) advantagesB.push('综合性能均衡');

        const historicalNotes = [
            `${bridgeA.name}：${bridgeA.key_innovation}`,
            `${bridgeB.name}：${bridgeB.key_innovation}`,
        ];

        if ((bridgeA.dynasty === 'han_jin' && bridgeB.dynasty === 'song') ||
            (bridgeB.dynasty === 'han_jin' && bridgeA.dynasty === 'song')) {
            historicalNotes.push('技术跃迁：从汉晋简支木梁到宋代贯木拱，跨度能力提升约40%');
            historicalNotes.push('材料效率提升：木拱结构相比简支梁可节省约25%的木材用量');
        }

        this.comparisonResult = {
            bridge_a: bridgeA,
            bridge_b: bridgeB,
            metrics_a: metricsA,
            metrics_b: metricsB,
            radar_data: radarData,
            normalized_scores: normalizedScores,
            advantages_a: advantagesA,
            advantages_b: advantagesB,
            historical_notes: historicalNotes,
            tech_evolution: [
                { period: '西周', innovation: '简支木梁桥出现', impact: 30.0, year: -1000 },
                { period: '秦汉', innovation: '石墩木梁桥普及', impact: 45.0, year: -200 },
                { period: '南北朝', innovation: '木拱技术萌芽', impact: 60.0, year: 500 },
                { period: '唐代', innovation: '单孔木拱桥出现', impact: 70.0, year: 700 },
                { period: '北宋', innovation: '贯木拱技术成熟', impact: 95.0, year: 1050 },
                { period: '南宋', innovation: '木拱廊桥发展', impact: 85.0, year: 1200 },
                { period: '明代', innovation: '三节拱五节拱组合', impact: 88.0, year: 1450 },
                { period: '清代', innovation: '工艺精细化与装饰', impact: 80.0, year: 1700 },
            ]
        };

        this.displayResults();
    }

    displayResults() {
        document.getElementById('hc-results').style.display = 'block';

        const r = this.comparisonResult;
        const bridgeA = r.bridge_a || r.BridgeA;
        const bridgeB = r.bridge_b || r.BridgeB;

        document.getElementById('hc-bridge-a-name').textContent = bridgeA.name;
        document.getElementById('hc-bridge-a-dynasty').textContent = bridgeA.historical_era || this.getDynastyName(bridgeA.dynasty);
        document.getElementById('hc-bridge-a-type').textContent = this.getTypologyName(bridgeA.typology);

        document.getElementById('hc-bridge-b-name').textContent = bridgeB.name;
        document.getElementById('hc-bridge-b-dynasty').textContent = bridgeB.historical_era || this.getDynastyName(bridgeB.dynasty);
        document.getElementById('hc-bridge-b-type').textContent = this.getTypologyName(bridgeB.typology);

        document.getElementById('hc-th-a').textContent = bridgeA.name;
        document.getElementById('hc-th-b').textContent = bridgeB.name;

        this.renderMetricsTable(bridgeA, bridgeB);
        this.drawRadarChart(r.radar_data || r.RadarData, bridgeA.name, bridgeB.name);
        this.renderAdvantages(r.advantages_a || r.AdvantagesA, r.advantages_b || r.AdvantagesB);
        this.renderHistoricalNotes(r.historical_notes || r.HistoricalNotes);
        this.renderEvolutionTimeline(r.tech_evolution || r.TechEvolution);
    }

    getDynastyName(dynasty) {
        const d = this.dynasties.find(d => d.id === dynasty);
        return d ? d.name : dynasty;
    }

    renderMetricsTable(bridgeA, bridgeB) {
        const r = this.comparisonResult;
        const metricsA = r.metrics_a || r.MetricsA;
        const metricsB = r.metrics_b || r.MetricsB;

        const rows = [
            { name: '跨径 (m)', a: bridgeA.span_length, b: bridgeB.span_length, higher: true },
            { name: '矢高 (m)', a: bridgeA.arch_rise, b: bridgeB.arch_rise, higher: true },
            { name: '桥面宽 (m)', a: bridgeA.deck_width, b: bridgeB.deck_width, higher: true },
            { name: '矢跨比', a: (bridgeA.arch_rise / bridgeA.span_length).toFixed(3), b: (bridgeB.arch_rise / bridgeB.span_length).toFixed(3), higher: true },
            { name: '材料效率', a: metricsA.material_efficiency || metricsA.MaterialEfficiency, b: metricsB.material_efficiency || metricsB.MaterialEfficiency, higher: true },
            { name: '承载能力', a: metricsA.load_carrying_capacity || metricsA.LoadCarryingCapacity, b: metricsB.load_carrying_capacity || metricsB.LoadCarryingCapacity, higher: true },
            { name: '施工复杂度', a: metricsA.construction_complexity || metricsA.ConstructionComplexity, b: metricsB.construction_complexity || metricsB.ConstructionComplexity, higher: false },
            { name: '耐久性评分', a: metricsA.durability_score || metricsA.DurabilityScore, b: metricsB.durability_score || metricsB.DurabilityScore, higher: true },
        ];

        const tbody = document.getElementById('hc-metrics-body');
        tbody.innerHTML = rows.map(row => {
            const valA = parseFloat(row.a);
            const valB = parseFloat(row.b);
            let winner = '-';
            if (row.higher) {
                if (valA > valB * 1.05) winner = `<span style="color: #3498db;">A</span>`;
                else if (valB > valA * 1.05) winner = `<span style="color: #e74c3c;">B</span>`;
            } else {
                if (valA < valB * 0.95) winner = `<span style="color: #3498db;">A</span>`;
                else if (valB < valA * 0.95) winner = `<span style="color: #e74c3c;">B</span>`;
            }
            return `
                <tr>
                    <td>${row.name}</td>
                    <td>${row.a}</td>
                    <td>${row.b}</td>
                    <td>${winner}</td>
                </tr>
            `;
        }).join('');
    }

    drawRadarChart(radarData, nameA, nameB) {
        const canvas = document.getElementById('hc-radar-chart');
        const ctx = canvas.getContext('2d');
        const centerX = canvas.width / 2;
        const centerY = canvas.height / 2;
        const radius = Math.min(centerX, centerY) - 50;
        const numAxes = radarData.length;

        ctx.clearRect(0, 0, canvas.width, canvas.height);

        for (let level = 1; level <= 5; level++) {
            ctx.beginPath();
            ctx.strokeStyle = '#e0e0e0';
            ctx.lineWidth = 1;
            for (let i = 0; i < numAxes; i++) {
                const angle = (Math.PI * 2 * i / numAxes) - Math.PI / 2;
                const r = radius * level / 5;
                const x = centerX + r * Math.cos(angle);
                const y = centerY + r * Math.sin(angle);
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            }
            ctx.closePath();
            ctx.stroke();
        }

        for (let i = 0; i < numAxes; i++) {
            const angle = (Math.PI * 2 * i / numAxes) - Math.PI / 2;
            ctx.beginPath();
            ctx.strokeStyle = '#ccc';
            ctx.lineWidth = 1;
            ctx.moveTo(centerX, centerY);
            ctx.lineTo(centerX + radius * Math.cos(angle), centerY + radius * Math.sin(angle));
            ctx.stroke();

            ctx.fillStyle = '#333';
            ctx.font = '12px Arial';
            ctx.textAlign = 'center';
            const labelX = centerX + (radius + 25) * Math.cos(angle);
            const labelY = centerY + (radius + 25) * Math.sin(angle);
            ctx.fillText(radarData[i].metric, labelX, labelY);
        }

        const drawPolygon = (data, color, fillColor) => {
            ctx.beginPath();
            ctx.strokeStyle = color;
            ctx.fillStyle = fillColor;
            ctx.lineWidth = 2;
            for (let i = 0; i < numAxes; i++) {
                const angle = (Math.PI * 2 * i / numAxes) - Math.PI / 2;
                const val = data[i].value_a !== undefined ? data[i].value_a : data[i].ValueA;
                const r = radius * (val / 100);
                const x = centerX + r * Math.cos(angle);
                const y = centerY + r * Math.sin(angle);
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            }
            ctx.closePath();
            ctx.fill();
            ctx.stroke();
        };

        drawPolygon(radarData, '#3498db', 'rgba(52, 152, 219, 0.2)');

        const drawPolygonB = (data, color, fillColor) => {
            ctx.beginPath();
            ctx.strokeStyle = color;
            ctx.fillStyle = fillColor;
            ctx.lineWidth = 2;
            for (let i = 0; i < numAxes; i++) {
                const angle = (Math.PI * 2 * i / numAxes) - Math.PI / 2;
                const val = data[i].value_b !== undefined ? data[i].value_b : data[i].ValueB;
                const r = radius * (val / 100);
                const x = centerX + r * Math.cos(angle);
                const y = centerY + r * Math.sin(angle);
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            }
            ctx.closePath();
            ctx.fill();
            ctx.stroke();
        };

        drawPolygonB(radarData, '#e74c3c', 'rgba(231, 76, 60, 0.2)');

        ctx.fillStyle = '#3498db';
        ctx.fillRect(canvas.width - 150, 20, 15, 15);
        ctx.fillStyle = '#333';
        ctx.font = '12px Arial';
        ctx.textAlign = 'left';
        ctx.fillText(nameA, canvas.width - 130, 32);

        ctx.fillStyle = '#e74c3c';
        ctx.fillRect(canvas.width - 150, 45, 15, 15);
        ctx.fillStyle = '#333';
        ctx.fillText(nameB, canvas.width - 130, 57);
    }

    renderAdvantages(advA, advB) {
        const ulA = document.getElementById('hc-adv-a');
        const ulB = document.getElementById('hc-adv-b');

        ulA.innerHTML = (advA || []).map(a => `<li>${a}</li>`).join('');
        ulB.innerHTML = (advB || []).map(b => `<li>${b}</li>`).join('');
    }

    renderHistoricalNotes(notes) {
        const ul = document.getElementById('hc-historical-notes');
        ul.innerHTML = (notes || []).map(n => `<li>${n}</li>`).join('');
    }

    renderEvolutionTimeline(evolution) {
        const container = document.getElementById('hc-evolution-timeline');
        container.innerHTML = (evolution || []).map((point, i) => `
            <div class="timeline-item">
                <div class="timeline-marker" style="height: ${point.impact}%;"></div>
                <div class="timeline-content">
                    <div class="timeline-period">${point.period}</div>
                    <div class="timeline-innovation">${point.innovation}</div>
                    <div class="timeline-year">${point.year > 0 ? '公元' + point.year + '年' : '公元前' + Math.abs(point.year) + '年'}</div>
                </div>
            </div>
        `).join('');
    }
}
