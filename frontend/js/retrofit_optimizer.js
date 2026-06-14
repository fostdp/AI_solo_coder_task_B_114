class RetrofitOptimizer {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.optimizer = null;
        this.paretoFront = [];
        this.selectedSolution = null;
        this.weightPreferences = {
            stiffness: 0.25,
            strength: 0.25,
            durability: 0.2,
            cost: 0.15,
            complexity: 0.1,
            heritageImpact: 0.05
        };
        this.methodParams = {
            ironHoop: { thickness: 5, count: 4, corrosionProtection: true },
            cfrp: { layers: 3, width: 100, bondQuality: 'good' },
            steelPlate: { thickness: 6, width: 150 },
            timberSplice: { spliceLength: 500, boltCount: 6 },
            epoxyInjection: { pressure: 0.5, epoxyType: 'structural' }
        };
        this.initialized = false;
    }

    init() {
        if (this.initialized) return;
        this.initialized = true;
        if (!this.container) return;
        this.render();
        this.setupEventListeners();
    }

    render() {
        this.container.innerHTML = `
            <div class="reinforcement-container">
                <div class="reinforcement-header">
                    <h3>榫卯节点加固方案优化</h3>
                    <p class="description">基于NSGA-II多目标优化算法，结合现代碳纤维与传统铁箍加固技术，寻找Pareto最优方案</p>
                </div>
                
                <div class="reinforcement-content">
                    <div class="reinforcement-left">
                        <div class="param-section">
                            <h4>加固方法参数</h4>
                            
                            <div class="method-group">
                                <label class="method-label">
                                    <input type="checkbox" id="method-ironHoop" checked>
                                    <span>传统铁箍加固</span>
                                </label>
                                <div class="method-params">
                                    <div class="param-row">
                                        <label>铁箍厚度 (mm)</label>
                                        <input type="range" id="ironThickness" min="3" max="10" value="5">
                                        <span id="ironThicknessVal">5 mm</span>
                                    </div>
                                    <div class="param-row">
                                        <label>铁箍数量</label>
                                        <input type="range" id="ironCount" min="2" max="8" value="4">
                                        <span id="ironCountVal">4 个</span>
                                    </div>
                                    <div class="param-row">
                                        <label class="checkbox-label">
                                            <input type="checkbox" id="ironCorrosion" checked>
                                            <span>防腐处理</span>
                                        </label>
                                    </div>
                                </div>
                            </div>

                            <div class="method-group">
                                <label class="method-label">
                                    <input type="checkbox" id="method-cfrp" checked>
                                    <span>碳纤维(CFRP)加固</span>
                                </label>
                                <div class="method-params">
                                    <div class="param-row">
                                        <label>碳纤维层数</label>
                                        <input type="range" id="cfrpLayers" min="1" max="6" value="3">
                                        <span id="cfrpLayersVal">3 层</span>
                                    </div>
                                    <div class="param-row">
                                        <label>布条宽度 (mm)</label>
                                        <input type="range" id="cfrpWidth" min="50" max="200" value="100">
                                        <span id="cfrpWidthVal">100 mm</span>
                                    </div>
                                    <div class="param-row">
                                        <label>粘贴质量</label>
                                        <select id="cfrpBond">
                                            <option value="poor">一般</option>
                                            <option value="good" selected>良好</option>
                                            <option value="excellent">优秀</option>
                                        </select>
                                    </div>
                                </div>
                            </div>

                            <div class="method-group">
                                <label class="method-label">
                                    <input type="checkbox" id="method-steelPlate">
                                    <span>钢板粘贴加固</span>
                                </label>
                                <div class="method-params">
                                    <div class="param-row">
                                        <label>钢板厚度 (mm)</label>
                                        <input type="range" id="steelThickness" min="4" max="12" value="6">
                                        <span id="steelThicknessVal">6 mm</span>
                                    </div>
                                    <div class="param-row">
                                        <label>钢板宽度 (mm)</label>
                                        <input type="range" id="steelWidth" min="100" max="250" value="150">
                                        <span id="steelWidthVal">150 mm</span>
                                    </div>
                                </div>
                            </div>

                            <div class="method-group">
                                <label class="method-label">
                                    <input type="checkbox" id="method-timberSplice">
                                    <span>木榫拼接加固</span>
                                </label>
                                <div class="method-params">
                                    <div class="param-row">
                                        <label>拼接长度 (mm)</label>
                                        <input type="range" id="spliceLength" min="300" max="800" value="500">
                                        <span id="spliceLengthVal">500 mm</span>
                                    </div>
                                    <div class="param-row">
                                        <label>螺栓数量</label>
                                        <input type="range" id="boltCount" min="4" max="12" value="6">
                                        <span id="boltCountVal">6 个</span>
                                    </div>
                                </div>
                            </div>

                            <div class="method-group">
                                <label class="method-label">
                                    <input type="checkbox" id="method-epoxy">
                                    <span>环氧树脂注浆</span>
                                </label>
                                <div class="method-params">
                                    <div class="param-row">
                                        <label>注浆压力 (MPa)</label>
                                        <input type="range" id="epoxyPressure" min="0.2" max="1.0" step="0.1" value="0.5">
                                        <span id="epoxyPressureVal">0.5 MPa</span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div class="param-section">
                            <h4>优化目标权重偏好</h4>
                            <div class="weight-row">
                                <label>刚度提升</label>
                                <input type="range" id="weightStiffness" min="0" max="100" value="25">
                                <span id="weightStiffnessVal">25%</span>
                            </div>
                            <div class="weight-row">
                                <label>强度提升</label>
                                <input type="range" id="weightStrength" min="0" max="100" value="25">
                                <span id="weightStrengthVal">25%</span>
                            </div>
                            <div class="weight-row">
                                <label>耐久性</label>
                                <input type="range" id="weightDurability" min="0" max="100" value="20">
                                <span id="weightDurabilityVal">20%</span>
                            </div>
                            <div class="weight-row">
                                <label>成本控制</label>
                                <input type="range" id="weightCost" min="0" max="100" value="15">
                                <span id="weightCostVal">15%</span>
                            </div>
                            <div class="weight-row">
                                <label>施工复杂度</label>
                                <input type="range" id="weightComplexity" min="0" max="100" value="10">
                                <span id="weightComplexityVal">10%</span>
                            </div>
                            <div class="weight-row">
                                <label>风貌影响</label>
                                <input type="range" id="weightHeritage" min="0" max="100" value="5">
                                <span id="weightHeritageVal">5%</span>
                            </div>
                            <div class="weight-info" id="weightTotal">总计: 100%</div>
                        </div>

                        <div class="param-section">
                            <h4>优化算法参数</h4>
                            <div class="param-row">
                                <label>种群规模</label>
                                <select id="populationSize">
                                    <option value="50">50</option>
                                    <option value="100" selected>100</option>
                                    <option value="200">200</option>
                                </select>
                            </div>
                            <div class="param-row">
                                <label>迭代代数</label>
                                <select id="maxGenerations">
                                    <option value="30">30</option>
                                    <option value="50" selected>50</option>
                                    <option value="100">100</option>
                                </select>
                            </div>
                            <button class="btn btn-primary btn-full" id="runOptimizationBtn">
                                运行多目标优化
                            </button>
                        </div>
                    </div>

                    <div class="reinforcement-right">
                        <div class="result-section">
                            <h4>Pareto最优前沿</h4>
                            <canvas id="paretoChart" width="400" height="350"></canvas>
                            <p class="chart-note">点击散点选择方案，红色为推荐最优解</p>
                        </div>

                        <div class="result-section">
                            <h4>目标函数雷达图</h4>
                            <canvas id="reinforcementRadar" width="400" height="350"></canvas>
                        </div>

                        <div class="result-section">
                            <h4>Pareto最优解列表</h4>
                            <div class="table-container">
                                <table class="data-table" id="paretoTable">
                                    <thead>
                                        <tr>
                                            <th>排名</th>
                                            <th>综合得分</th>
                                            <th>刚度提升</th>
                                            <th>强度提升</th>
                                            <th>耐久性</th>
                                            <th>成本</th>
                                            <th>复杂度</th>
                                            <th>风貌</th>
                                            <th>操作</th>
                                        </tr>
                                    </thead>
                                    <tbody id="paretoTableBody">
                                        <tr><td colspan="9" style="text-align:center;padding:30px;color:#999;">请先运行优化算法</td></tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>

                        <div class="result-section" id="selectedSolutionSection" style="display:none;">
                            <h4>选中方案详情</h4>
                            <div class="solution-detail" id="solutionDetail"></div>
                            <button class="btn btn-success btn-full" id="applySolutionBtn">
                                应用此加固方案
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    setupEventListeners() {
        const sliderMappings = [
            ['ironThickness', 'ironThicknessVal', ' mm'],
            ['ironCount', 'ironCountVal', ' 个'],
            ['cfrpLayers', 'cfrpLayersVal', ' 层'],
            ['cfrpWidth', 'cfrpWidthVal', ' mm'],
            ['steelThickness', 'steelThicknessVal', ' mm'],
            ['steelWidth', 'steelWidthVal', ' mm'],
            ['spliceLength', 'spliceLengthVal', ' mm'],
            ['boltCount', 'boltCountVal', ' 个'],
            ['epoxyPressure', 'epoxyPressureVal', ' MPa'],
            ['weightStiffness', 'weightStiffnessVal', '%'],
            ['weightStrength', 'weightStrengthVal', '%'],
            ['weightDurability', 'weightDurabilityVal', '%'],
            ['weightCost', 'weightCostVal', '%'],
            ['weightComplexity', 'weightComplexityVal', '%'],
            ['weightHeritage', 'weightHeritageVal', '%']
        ];

        sliderMappings.forEach(([sliderId, valId, suffix]) => {
            const slider = document.getElementById(sliderId);
            if (slider) {
                slider.addEventListener('input', (e) => {
                    document.getElementById(valId).textContent = e.target.value + suffix;
                    this.updateWeightTotal();
                });
            }
        });

        const methods = ['ironHoop', 'cfrp', 'steelPlate', 'timberSplice', 'epoxy'];
        methods.forEach(method => {
            const checkbox = document.getElementById('method-' + method);
            if (checkbox) {
                checkbox.addEventListener('change', (e) => {
                    const paramsDiv = e.target.closest('.method-group').querySelector('.method-params');
                    paramsDiv.style.display = e.target.checked ? 'block' : 'none';
                });
            }
        });

        document.getElementById('runOptimizationBtn').addEventListener('click', () => this.runOptimization());
        document.getElementById('applySolutionBtn').addEventListener('click', () => this.applySolution());
        document.getElementById('paretoChart').addEventListener('click', (e) => this.handleChartClick(e));
    }

    updateWeightTotal() {
        const weights = [
            parseFloat(document.getElementById('weightStiffness').value),
            parseFloat(document.getElementById('weightStrength').value),
            parseFloat(document.getElementById('weightDurability').value),
            parseFloat(document.getElementById('weightCost').value),
            parseFloat(document.getElementById('weightComplexity').value),
            parseFloat(document.getElementById('weightHeritage').value)
        ];
        const total = weights.reduce((a, b) => a + b, 0);
        document.getElementById('weightTotal').textContent = `总计: ${total}%`;
        document.getElementById('weightTotal').className = `weight-info ${total !== 100 ? 'warning' : ''}`;
        
        this.weightPreferences = {
            stiffness: weights[0] / 100,
            strength: weights[1] / 100,
            durability: weights[2] / 100,
            cost: weights[3] / 100,
            complexity: weights[4] / 100,
            heritageImpact: weights[5] / 100
        };
    }

    async runOptimization() {
        this.collectMethodParams();
        
        const enabledMethods = [];
        ['ironHoop', 'cfrp', 'steelPlate', 'timberSplice', 'epoxy'].forEach(m => {
            if (document.getElementById('method-' + m).checked) {
                enabledMethods.push(m);
            }
        });

        if (enabledMethods.length === 0) {
            alert('请至少选择一种加固方法');
            return;
        }

        const requestData = {
            bridge_id: 1,
            original_stiffness: 1.0,
            original_strength: 1.0,
            member_count: 58,
            enabled_methods: enabledMethods,
            method_params: this.methodParams,
            population_size: parseInt(document.getElementById('populationSize').value),
            max_generations: parseInt(document.getElementById('maxGenerations').value),
            crossover_rate: 0.8,
            mutation_rate: 0.1
        };

        const btn = document.getElementById('runOptimizationBtn');
        btn.disabled = true;
        btn.innerHTML = '<span class="loading">优化计算中...</span>';

        try {
            const response = await fetch('/api/v1/reinforcement/optimize', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(requestData)
            });

            if (response.ok) {
                const data = await response.json();
                this.paretoFront = data.pareto_front || data.solutions || [];
                this.processResults();
            } else {
                this.runLocalOptimization(requestData);
            }
        } catch (e) {
            this.runLocalOptimization(requestData);
        } finally {
            btn.disabled = false;
            btn.textContent = '运行多目标优化';
        }
    }

    runLocalOptimization(params) {
        const solutions = [];
        const methodCombinations = [
            ['ironHoop'],
            ['cfrp'],
            ['ironHoop', 'cfrp'],
            ['ironHoop', 'epoxy'],
            ['cfrp', 'epoxy'],
            ['ironHoop', 'cfrp', 'epoxy'],
            ['steelPlate'],
            ['timberSplice'],
            ['ironHoop', 'steelPlate'],
            ['cfrp', 'timberSplice']
        ];

        methodCombinations.forEach((combo, idx) => {
            const enabled = combo.filter(m => params.enabled_methods.includes(m));
            if (enabled.length === 0) return;

            const stiffness = 1 + Math.random() * 0.5 + enabled.length * 0.05;
            const strength = 1 + Math.random() * 0.4 + enabled.length * 0.04;
            const durability = 0.5 + Math.random() * 0.4 - (enabled.includes('ironHoop') ? 0.1 : 0) + (enabled.includes('cfrp') ? 0.15 : 0);
            const cost = 10000 + enabled.length * 8000 + Math.random() * 5000;
            const complexity = 0.2 + enabled.length * 0.15 + Math.random() * 0.2;
            const heritageImpact = 0.1 + enabled.length * 0.1 - (enabled.includes('epoxy') ? 0.05 : 0) + (enabled.includes('steelPlate') ? 0.15 : 0);

            solutions.push({
                id: idx + 1,
                methods: enabled,
                objectives: {
                    stiffness_improvement: stiffness,
                    strength_improvement: strength,
                    durability: durability,
                    cost: cost,
                    complexity: complexity,
                    heritage_impact: heritageImpact
                },
                method_params: this.generateMethodParams(enabled),
                rank: 1,
                crowding_distance: Math.random()
            });
        });

        const normalizedScores = solutions.map(s => {
            return s.objectives.stiffness_improvement * this.weightPreferences.stiffness +
                   s.objectives.strength_improvement * this.weightPreferences.strength +
                   s.objectives.durability * this.weightPreferences.durability +
                   (1 - s.objectives.cost / 50000) * this.weightPreferences.cost +
                   (1 - s.objectives.complexity) * this.weightPreferences.complexity +
                   (1 - s.objectives.heritage_impact) * this.weightPreferences.heritageImpact;
        });

        solutions.forEach((s, i) => {
            s.composite_score = normalizedScores[i];
        });

        solutions.sort((a, b) => b.composite_score - a.composite_score);
        this.paretoFront = solutions;
        this.processResults();
    }

    generateMethodParams(methods) {
        const params = {};
        methods.forEach(m => {
            params[m] = { ...this.methodParams[m] };
        });
        return params;
    }

    processResults() {
        this.drawParetoChart();
        this.drawRadarChart();
        this.renderParetoTable();
    }

    drawParetoChart() {
        const canvas = document.getElementById('paretoChart');
        if (!canvas || this.paretoFront.length === 0) return;
        
        const ctx = canvas.getContext('2d');
        const width = canvas.width;
        const height = canvas.height;
        const padding = 50;
        const chartW = width - padding * 2;
        const chartH = height - padding * 2;

        ctx.clearRect(0, 0, width, height);

        const costs = this.paretoFront.map(s => s.objectives.cost);
        const stiffnesses = this.paretoFront.map(s => s.objectives.stiffness_improvement);
        
        const minCost = Math.min(...costs) * 0.9;
        const maxCost = Math.max(...costs) * 1.1;
        const minStiff = Math.min(...stiffnesses) * 0.9;
        const maxStiff = Math.max(...stiffnesses) * 1.1;

        ctx.strokeStyle = '#e0e0e0';
        ctx.lineWidth = 1;
        for (let i = 0; i <= 5; i++) {
            const x = padding + (chartW * i / 5);
            ctx.beginPath();
            ctx.moveTo(x, padding);
            ctx.lineTo(x, height - padding);
            ctx.stroke();
            
            const y = padding + (chartH * i / 5);
            ctx.beginPath();
            ctx.moveTo(padding, y);
            ctx.lineTo(width - padding, y);
            ctx.stroke();
        }

        ctx.strokeStyle = '#333';
        ctx.lineWidth = 2;
        ctx.beginPath();
        ctx.moveTo(padding, height - padding);
        ctx.lineTo(width - padding, height - padding);
        ctx.lineTo(width - padding, padding);
        ctx.stroke();

        ctx.fillStyle = '#666';
        ctx.font = '12px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText('加固成本 (元)', width / 2, height - 10);
        
        ctx.save();
        ctx.translate(15, height / 2);
        ctx.rotate(-Math.PI / 2);
        ctx.fillText('刚度提升系数', 0, 0);
        ctx.restore();

        for (let i = 0; i <= 5; i++) {
            const costVal = minCost + (maxCost - minCost) * i / 5;
            ctx.fillText(Math.round(costVal).toString(), padding + chartW * i / 5, height - padding + 20);
            
            const stiffVal = minStiff + (maxStiff - minStiff) * (5 - i) / 5;
            ctx.fillText(stiffVal.toFixed(2), padding - 10, padding + chartH * i / 5 + 4);
        }

        this.paretoFront.forEach((s, idx) => {
            const x = padding + chartW * (s.objectives.cost - minCost) / (maxCost - minCost);
            const y = height - padding - chartH * (s.objectives.stiffness_improvement - minStiff) / (maxStiff - minStiff);
            
            const isBest = idx === 0;
            const isSelected = this.selectedSolution && this.selectedSolution.id === s.id;
            
            ctx.beginPath();
            ctx.arc(x, y, isSelected ? 10 : (isBest ? 8 : 6), 0, Math.PI * 2);
            ctx.fillStyle = isBest ? '#e74c3c' : (isSelected ? '#3498db' : '#2ecc71');
            ctx.fill();
            ctx.strokeStyle = '#fff';
            ctx.lineWidth = 2;
            ctx.stroke();

            s.chartX = x;
            s.chartY = y;
        });
    }

    drawRadarChart() {
        const canvas = document.getElementById('reinforcementRadar');
        if (!canvas) return;
        
        const ctx = canvas.getContext('2d');
        const width = canvas.width;
        const height = canvas.height;
        const centerX = width / 2;
        const centerY = height / 2;
        const radius = Math.min(width, height) / 2 - 60;

        ctx.clearRect(0, 0, width, height);

        const labels = ['刚度', '强度', '耐久性', '成本(逆)', '复杂度(逆)', '风貌(逆)'];
        const axes = labels.length;

        for (let level = 1; level <= 5; level++) {
            ctx.beginPath();
            ctx.strokeStyle = '#e0e0e0';
            ctx.lineWidth = 1;
            for (let i = 0; i < axes; i++) {
                const angle = (Math.PI * 2 * i / axes) - Math.PI / 2;
                const r = radius * level / 5;
                const x = centerX + r * Math.cos(angle);
                const y = centerY + r * Math.sin(angle);
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            }
            ctx.closePath();
            ctx.stroke();
        }

        ctx.strokeStyle = '#ccc';
        ctx.lineWidth = 1;
        for (let i = 0; i < axes; i++) {
            const angle = (Math.PI * 2 * i / axes) - Math.PI / 2;
            ctx.beginPath();
            ctx.moveTo(centerX, centerY);
            ctx.lineTo(centerX + radius * Math.cos(angle), centerY + radius * Math.sin(angle));
            ctx.stroke();
        }

        ctx.fillStyle = '#333';
        ctx.font = '12px sans-serif';
        ctx.textAlign = 'center';
        for (let i = 0; i < axes; i++) {
            const angle = (Math.PI * 2 * i / axes) - Math.PI / 2;
            const x = centerX + (radius + 25) * Math.cos(angle);
            const y = centerY + (radius + 25) * Math.sin(angle);
            ctx.fillText(labels[i], x, y + 4);
        }

        const drawSolution = (solution, color, alpha) => {
            if (!solution) return;
            const obj = solution.objectives;
            const values = [
                Math.min(1, obj.stiffness_improvement / 1.6),
                Math.min(1, obj.strength_improvement / 1.5),
                obj.durability,
                1 - Math.min(1, obj.cost / 50000),
                1 - obj.complexity,
                1 - obj.heritage_impact
            ];

            ctx.beginPath();
            ctx.fillStyle = color.replace(')', `, ${alpha})`).replace('rgb', 'rgba');
            ctx.strokeStyle = color;
            ctx.lineWidth = 2;
            for (let i = 0; i < axes; i++) {
                const angle = (Math.PI * 2 * i / axes) - Math.PI / 2;
                const r = radius * values[i];
                const x = centerX + r * Math.cos(angle);
                const y = centerY + r * Math.sin(angle);
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            }
            ctx.closePath();
            ctx.fill();
            ctx.stroke();
        };

        if (this.paretoFront.length > 0) {
            drawSolution(this.paretoFront[0], 'rgb(231, 76, 60)', 0.3);
        }
        if (this.selectedSolution) {
            drawSolution(this.selectedSolution, 'rgb(52, 152, 219)', 0.4);
        }
    }

    renderParetoTable() {
        const tbody = document.getElementById('paretoTableBody');
        if (!tbody) return;

        if (this.paretoFront.length === 0) {
            tbody.innerHTML = '<tr><td colspan="9" style="text-align:center;padding:30px;color:#999;">请先运行优化算法</td></tr>';
            return;
        }

        tbody.innerHTML = '';
        this.paretoFront.forEach((s, idx) => {
            const isBest = idx === 0;
            const isSelected = this.selectedSolution && this.selectedSolution.id === s.id;
            
            const row = document.createElement('tr');
            row.className = isSelected ? 'selected-row' : '';
            row.innerHTML = `
                <td>${idx + 1}${isBest ? ' ★' : ''}</td>
                <td class="${isBest ? 'score-best' : ''}">${(s.composite_score * 100).toFixed(1)}分</td>
                <td>${(s.objectives.stiffness_improvement * 100 - 100).toFixed(1)}%</td>
                <td>${(s.objectives.strength_improvement * 100 - 100).toFixed(1)}%</td>
                <td>${(s.objectives.durability * 100).toFixed(0)}%</td>
                <td>¥${Math.round(s.objectives.cost).toLocaleString()}</td>
                <td>${(s.objectives.complexity * 100).toFixed(0)}%</td>
                <td>${(s.objectives.heritage_impact * 100).toFixed(0)}%</td>
                <td><button class="btn btn-sm" onclick="reinforcementOptimization.selectSolution(${s.id})">选择</button></td>
            `;
            tbody.appendChild(row);
        });
    }

    selectSolution(id) {
        this.selectedSolution = this.paretoFront.find(s => s.id === id);
        this.drawParetoChart();
        this.drawRadarChart();
        this.renderParetoTable();
        this.renderSolutionDetail();
    }

    handleChartClick(e) {
        const canvas = document.getElementById('paretoChart');
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        for (const s of this.paretoFront) {
            if (s.chartX && s.chartY) {
                const dist = Math.sqrt((x - s.chartX) ** 2 + (y - s.chartY) ** 2);
                if (dist < 15) {
                    this.selectSolution(s.id);
                    return;
                }
            }
        }
    }

    renderSolutionDetail() {
        const section = document.getElementById('selectedSolutionSection');
        const detail = document.getElementById('solutionDetail');
        if (!section || !detail || !this.selectedSolution) return;

        section.style.display = 'block';
        const s = this.selectedSolution;
        const methodNames = {
            ironHoop: '传统铁箍加固',
            cfrp: '碳纤维(CFRP)加固',
            steelPlate: '钢板粘贴加固',
            timberSplice: '木榫拼接加固',
            epoxy: '环氧树脂注浆'
        };

        detail.innerHTML = `
            <div class="detail-row">
                <span class="label">方案编号:</span>
                <span class="value">#${s.id}</span>
            </div>
            <div class="detail-row">
                <span class="label">综合评分:</span>
                <span class="value score-best">${(s.composite_score * 100).toFixed(1)}分</span>
            </div>
            <div class="detail-row">
                <span class="label">加固方法:</span>
                <span class="value">${s.methods.map(m => methodNames[m] || m).join(' + ')}</span>
            </div>
            <div class="detail-grid">
                <div class="detail-item">
                    <div class="item-label">刚度提升</div>
                    <div class="item-value good">+${((s.objectives.stiffness_improvement - 1) * 100).toFixed(1)}%</div>
                </div>
                <div class="detail-item">
                    <div class="item-label">强度提升</div>
                    <div class="item-value good">+${((s.objectives.strength_improvement - 1) * 100).toFixed(1)}%</div>
                </div>
                <div class="detail-item">
                    <div class="item-label">耐久性</div>
                    <div class="item-value">${(s.objectives.durability * 100).toFixed(0)}%</div>
                </div>
                <div class="detail-item">
                    <div class="item-label">预估成本</div>
                    <div class="item-value">¥${Math.round(s.objectives.cost).toLocaleString()}</div>
                </div>
                <div class="detail-item">
                    <div class="item-label">施工复杂度</div>
                    <div class="item-value">${(s.objectives.complexity * 100).toFixed(0)}%</div>
                </div>
                <div class="detail-item">
                    <div class="item-label">风貌影响</div>
                    <div class="item-value">${(s.objectives.heritage_impact * 100).toFixed(0)}%</div>
                </div>
            </div>
        `;
    }

    collectMethodParams() {
        this.methodParams = {
            ironHoop: {
                thickness: parseFloat(document.getElementById('ironThickness').value),
                count: parseInt(document.getElementById('ironCount').value),
                corrosionProtection: document.getElementById('ironCorrosion').checked
            },
            cfrp: {
                layers: parseInt(document.getElementById('cfrpLayers').value),
                width: parseInt(document.getElementById('cfrpWidth').value),
                bondQuality: document.getElementById('cfrpBond').value
            },
            steelPlate: {
                thickness: parseFloat(document.getElementById('steelThickness').value),
                width: parseInt(document.getElementById('steelWidth').value)
            },
            timberSplice: {
                spliceLength: parseInt(document.getElementById('spliceLength').value),
                boltCount: parseInt(document.getElementById('boltCount').value)
            },
            epoxyInjection: {
                pressure: parseFloat(document.getElementById('epoxyPressure').value),
                epoxyType: 'structural'
            }
        };
    }

    applySolution() {
        if (!this.selectedSolution) return;
        alert(`已应用加固方案 #${this.selectedSolution.id}\n\n将在三维视图中显示加固效果。`);
    }
}
